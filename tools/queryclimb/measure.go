package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

type caseSpec struct {
	Name        string              `json:"name"`
	Query       string              `json:"query"`
	Tenant      string              `json:"tenant"`
	Args        []*string           `json:"args"`
	TieColumns  []int               `json:"tie_columns"`
	MaskColumns []int               `json:"mask_columns,omitempty"`
	TableMasks  map[string][]string `json:"table_masks,omitempty"`
	EmptyEdge   bool                `json:"empty_edge,omitempty"`
	WarmArgs    [][]*string         `json:"warm_args,omitempty"`
	StateTables []string            `json:"state_tables,omitempty"`
}

type metrics struct {
	QueryID       int64           `json:"query_id"`
	Query         string          `json:"query"`
	Case          string          `json:"case"`
	RowsRead      float64         `json:"rows_read"`
	RowsProcessed float64         `json:"rows_processed"`
	SharedBuffers float64         `json:"shared_buffers"`
	TempBuffers   float64         `json:"temp_buffers"`
	PlanHash      string          `json:"plan_hash"`
	Fingerprint   string          `json:"fingerprint"`
	Indexes       []string        `json:"indexes"`
	Signatures    []string        `json:"signatures"`
	Plan          json.RawMessage `json:"plan"`
	GenericPlans  int64           `json:"generic_plans"`
	CustomPlans   int64           `json:"custom_plans"`
	Returned      int             `json:"returned"`
	Affected      int64           `json:"affected"`
}

func inspect(ctx context.Context, root string, args []string) error {
	flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
	raw := flags.String("database", localDSN, "owned loopback administrator DSN")
	database := flags.String("name", "qc_schema", "database to inspect")
	casesPath := flags.String("cases", "", "JSON case definitions with text-format arguments")
	diagnostic := flags.Bool("diagnostic", false, "continue after case failures; never use this output as a passing gate")
	if err := flags.Parse(args); err != nil {
		return err
	}
	conn, err := connect(ctx, *raw, *database, "tadoku")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	queries, err := inventory(root)
	if err != nil {
		return err
	}
	cases, err := readCases(*casesPath)
	if err != nil {
		return err
	}
	results := []metrics{}
	for _, c := range cases {
		q, err := findQuery(queries, c.Query)
		if err != nil {
			return err
		}
		result, err := measure(ctx, conn, q, c)
		if err != nil {
			if *diagnostic {
				fmt.Fprintf(os.Stderr, "%s/%s: %v\n", c.Query, c.Name, err)
				continue
			}
			return fmt.Errorf("%s/%s: %w", c.Query, c.Name, err)
		}
		results = append(results, result)
	}
	return output(results)
}

func readCases(path string) ([]caseSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cases []caseSpec
	if err := json.Unmarshal(b, &cases); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, c := range cases {
		key := c.Query + "/" + c.Name
		if c.Tenant == "" || c.Name == "" || seen[key] {
			return nil, fmt.Errorf("invalid or duplicate case %s", key)
		}
		seen[key] = true
	}
	return cases, nil
}

func findQuery(queries []query, name string) (query, error) {
	for _, q := range queries {
		if q.Key == name {
			return q, nil
		}
	}
	return query{}, fmt.Errorf("unknown query %s", name)
}

func scoped(ctx context.Context, conn *pgx.Conn, tenant string, work func() error) (err error) {
	if _, err := conn.Exec(ctx, "begin"); err != nil {
		return err
	}
	defer func() {
		_, rollbackErr := conn.Exec(context.WithoutCancel(ctx), "rollback")
		err = errors.Join(err, rollbackErr)
	}()
	if _, err := conn.Exec(ctx, "select set_config('tadoku.tenant', $1, true)", tenant); err != nil {
		return err
	}
	return work()
}

func execute(ctx context.Context, conn *pgx.Conn, c caseSpec) ([][]*string, int64, error) {
	params := make([][]byte, len(c.Args))
	for i, arg := range c.Args {
		if arg != nil {
			params[i] = []byte(*arg)
		}
	}
	result := conn.PgConn().ExecPrepared(ctx, "qc_query", params, nil, []int16{0}).Read()
	if result.Err != nil {
		return nil, 0, result.Err
	}
	rows := make([][]*string, len(result.Rows))
	for i, raw := range result.Rows {
		rows[i] = make([]*string, len(raw))
		for j, v := range raw {
			if v != nil {
				text := string(v)
				rows[i][j] = &text
			}
		}
	}
	return rows, result.CommandTag.RowsAffected(), nil
}

func measure(ctx context.Context, conn *pgx.Conn, q query, c caseSpec) (metrics, error) {
	result := metrics{Query: q.Key, Case: c.Name, Indexes: []string{}, Signatures: []string{}}
	if err := conn.Deallocate(ctx, "qc_query"); err != nil {
		return result, err
	}
	desc, err := conn.Prepare(ctx, "qc_query", q.SQL)
	if err != nil {
		return result, err
	}
	if len(desc.ParamOIDs) != len(c.Args) {
		return result, errors.New("case argument count differs from prepared statement")
	}

	if len(c.WarmArgs) != 0 && len(c.WarmArgs) != 5 {
		return result, errors.New("warm_args must have five executions")
	}
	for i := 0; i < 5; i++ {
		warm := c
		if len(c.WarmArgs) != 0 {
			warm.Args = c.WarmArgs[i]
		}
		err := scoped(ctx, conn, c.Tenant, func() error { _, _, err := execute(ctx, conn, warm); return err })
		if err != nil {
			return result, err
		}
	}
	literals := make([]string, len(c.Args))
	for i, arg := range c.Args {
		var typeName string
		if err := conn.QueryRow(ctx, "select format_type($1::oid, null)", desc.ParamOIDs[i]).Scan(&typeName); err != nil {
			return result, err
		}
		literals[i] = "null::" + typeName
		if arg != nil {
			literals[i] = "'" + strings.ReplaceAll(*arg, "'", "''") + "'::" + typeName
		}
	}
	executeSQL := "execute qc_query"
	if len(literals) > 0 {
		executeSQL += "(" + strings.Join(literals, ",") + ")"
	}
	err = scoped(ctx, conn, c.Tenant, func() error {
		return conn.QueryRow(ctx,
			"explain (analyze, buffers, verbose, format json) "+executeSQL,
			pgx.QueryExecModeSimpleProtocol,
		).Scan(&result.Plan)
	})
	if err != nil {
		return result, err
	}
	if err := conn.QueryRow(ctx, `select generic_plans, custom_plans from pg_prepared_statements
  where name = 'qc_query'`).Scan(&result.GenericPlans, &result.CustomPlans); err != nil {
		return result, err
	}
	if result.GenericPlans+result.CustomPlans != 6 {
		return result, errors.New("plan was not measured on execution six")
	}
	var identity []struct {
		QueryID int64 `json:"Query Identifier"`
	}
	if err := json.Unmarshal(result.Plan, &identity); err != nil {
		return result, err
	}
	if len(identity) != 1 || identity[0].QueryID == 0 {
		return result, errors.New("EXPLAIN did not return a query identifier")
	}
	result.QueryID = identity[0].QueryID
	if err := analyzePlan(&result); err != nil {
		return result, err
	}

	err = scoped(ctx, conn, c.Tenant, func() error {
		rows, affected, err := execute(ctx, conn, c)
		if err != nil {
			return err
		}
		result.Returned, result.Affected = len(rows), affected
		result.Fingerprint, err = fingerprint(rows, c.TieColumns, c.MaskColumns)
		if err != nil {
			return err
		}
		if len(rows) == 0 && affected == 0 && !c.EmptyEdge {
			return errors.New("empty result requires an explicit empty_edge case")
		}
		if q.Write {
			tables := c.StateTables
			if len(tables) == 0 {
				tables = q.Mutations
			}
			expanded, err := affectedTables(ctx, conn, tables)
			if err != nil {
				return err
			}
			state, err := tableState(ctx, conn, expanded, c.TableMasks)
			if err != nil {
				return err
			}
			result.Fingerprint = digest(result.Fingerprint + "/" + strconv.FormatInt(affected, 10) + "/" + state)
		}
		return nil
	})
	return result, err
}

func fingerprint(rows [][]*string, ties, masks []int) (string, error) {
	encoded := make([]string, len(rows))
	keys := make([]string, len(rows))
	for i, row := range rows {
		copyRow := slices.Clone(row)
		for _, col := range masks {
			if col < 0 || col >= len(row) {
				return "", errors.New("mask column out of range")
			}
			if copyRow[col] != nil {
				value := "<volatile>"
				copyRow[col] = &value
			}
		}
		key := []*string{}
		for _, col := range ties {
			if col < 0 || col >= len(row) {
				return "", errors.New("tie column out of range")
			}
			key = append(key, row[col])
		}
		b, _ := json.Marshal(copyRow)
		encoded[i] = string(b)
		b, _ = json.Marshal(key)
		keys[i] = string(b)
	}
	for start := 0; start < len(rows); {
		end := start + 1
		for end < len(rows) && keys[end] == keys[start] {
			end++
		}
		slices.Sort(encoded[start:end])
		start = end
	}
	b, _ := json.Marshal(encoded)
	return digest(string(b)), nil
}

func tableState(ctx context.Context, conn *pgx.Conn, tables []string, masks map[string][]string) (string, error) {
	values := []string{}
	for _, table := range tables {

		projection := "to_jsonb(t)"
		for _, name := range masks[table] {
			literal := "'" + strings.ReplaceAll(name, "'", "''") + "'"
			projection += " || jsonb_build_object(" + literal + ",\n case\n when to_jsonb(t)->" + literal +
				" = 'null'::jsonb then 'null'::jsonb\n else to_jsonb('<volatile>'::text)\n end)"
		}
		sql := "select row_data from (select (" + projection + ")::text as row_data from " +
			pgx.Identifier{table}.Sanitize() + " t) s order by row_data collate \"C\""
		rows, err := conn.Query(ctx, sql)
		if err != nil {
			return "", err
		}
		hash := sha256.New()
		var frame [8]byte
		for rows.Next() {
			var row string
			if err := rows.Scan(&row); err != nil {
				rows.Close()
				return "", err
			}
			binary.BigEndian.PutUint64(frame[:], uint64(len(row)))
			hash.Write(frame[:])
			hash.Write([]byte(row))
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return "", err
		}
		values = append(values, table+":"+hex.EncodeToString(hash.Sum(nil)))
	}
	slices.Sort(values)
	return digest(strings.Join(values, "\n")), nil
}

func digest(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func analyzePlan(result *metrics) error {
	var document []map[string]any
	if err := json.Unmarshal(result.Plan, &document); err != nil {
		return err
	}
	if len(document) != 1 {
		return errors.New("expected exactly one explain plan")
	}
	root, ok := document[0]["Plan"].(map[string]any)
	if !ok {
		return errors.New("missing plan")
	}
	number := func(node map[string]any, name string) float64 { v, _ := node[name].(float64); return v }
	result.SharedBuffers = number(root, "Shared Hit Blocks") + number(root, "Shared Read Blocks")
	result.TempBuffers = number(root, "Temp Read Blocks") + number(root, "Temp Written Blocks")
	indexes, signatures := map[string]bool{}, map[string]bool{}
	var shape func(map[string]any, bool) map[string]any
	shape = func(node map[string]any, underLimit bool) map[string]any {
		result.RowsProcessed += number(node, "Actual Rows") * number(node, "Actual Loops")
		stable := map[string]any{}
		for _, key := range []string{"Node Type", "Parent Relationship", "Relation Name", "Alias", "Index Name",
			"Join Type", "Strategy", "CTE Name", "Subplan Name", "Sort Key", "Group Key", "Filter", "Index Cond",
			"Hash Cond", "Merge Cond", "Join Filter", "Parallel Aware", "Scan Direction", "Operation"} {
			if v, exists := node[key]; exists {
				stable[key] = v
			}
		}
		if index, ok := node["Index Name"].(string); ok {
			indexes[index] = true
		}
		kind, _ := node["Node Type"].(string)
		if _, relation := node["Relation Name"]; relation && kind != "Bitmap Index Scan" {
			result.RowsRead += (number(node, "Actual Rows") + number(node, "Rows Removed by Filter") +
				number(node, "Rows Removed by Index Recheck")) * number(node, "Actual Loops")
		}
		if node["Parent Relationship"] == "SubPlan" && number(node, "Actual Loops") > 100 {
			signatures["per-row-subplan"] = true
		}
		if node["Sort Space Type"] == "Disk" {
			signatures["sort-spill"] = true
		}
		if kind == "Seq Scan" {
			filter, _ := node["Filter"].(string)
			if strings.Contains(filter, "IS NULL") && strings.Contains(filter, " OR ") {
				signatures["optional-filter-seq-scan"] = true
			}
		}
		if kind == "CTE Scan" {
			signatures["materialized-cte"] = true
		}
		if underLimit && kind == "Aggregate" {
			signatures["aggregate-under-limit"] = true
		}
		if children, ok := node["Plans"].([]any); ok {
			childShapes := []any{}
			for _, child := range children {
				childShapes = append(childShapes, shape(child.(map[string]any), underLimit || kind == "Limit"))
			}
			stable["Plans"] = childShapes
		}
		return stable
	}
	encoded, err := json.Marshal(shape(root, false))
	if err != nil {
		return err
	}
	result.PlanHash = digest(string(encoded))
	for name := range indexes {
		result.Indexes = append(result.Indexes, name)
	}
	for name := range signatures {
		result.Signatures = append(result.Signatures, name)
	}
	slices.Sort(result.Indexes)
	slices.Sort(result.Signatures)
	return nil
}

func affectedTables(ctx context.Context, conn *pgx.Conn, tables []string) ([]string, error) {
	if len(tables) == 0 {
		return nil, errors.New("write has no declared affected tables")
	}
	rows, err := conn.Query(ctx, `with recursive affected(oid) as (
   select c.oid from pg_class c join pg_namespace n on n.oid = c.relnamespace
   where n.nspname = 'public' and c.relname = any($1::text[])
   union
   select f.conrelid from pg_constraint f join affected a on a.oid = f.confrelid
   where f.contype = 'f' and (f.confdeltype = 'c' or f.confupdtype = 'c')
  )
  select c.relname::text from affected a join pg_class c on c.oid = a.oid order by c.relname`, tables)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		result = append(result, name)
	}
	return result, rows.Err()
}
