package migrationsafety

import (
	"fmt"
	"slices"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type Finding struct {
	Rule    string
	Line    int
	Message string
}

// Check reports migration patterns that require review before deployment.
func Check(sql string) ([]Finding, error) {
	parsed, err := pg_query.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("parse migration: %w", err)
	}
	scanned, err := pg_query.Scan(sql)
	if err != nil {
		return nil, fmt.Errorf("scan migration: %w", err)
	}
	tokens := scanned.GetTokens()
	next := 0
	var findings []Finding
	for _, raw := range parsed.Stmts {
		for next < len(tokens) && (tokens[next].GetStart() < raw.GetStmtLocation() ||
			tokens[next].GetToken() == pg_query.Token_SQL_COMMENT || tokens[next].GetToken() == pg_query.Token_C_COMMENT) {
			next++
		}
		start := len(sql)
		if next < len(tokens) {
			start = int(tokens[next].GetStart())
		}
		line := 1 + strings.Count(sql[:start], "\n")
		add := func(rule, message string) {
			findings = append(findings, Finding{Rule: rule, Line: line, Message: message})
		}
		stmt := raw.Stmt
		if index := stmt.GetIndexStmt(); index != nil {
			if index.GetConcurrent() && len(parsed.Stmts) > 1 {
				add("concurrently-inside-transaction", "CREATE INDEX CONCURRENTLY must be the only statement in its migration file")
			} else if !index.GetConcurrent() {
				add("create-index-without-concurrently", "CREATE INDEX without CONCURRENTLY blocks writes")
			}
		}
		if alter := stmt.GetAlterTableStmt(); alter != nil {
			for _, command := range alter.GetCmds() {
				checkAlter(command.GetAlterTableCmd(), add)
			}
		}
		if stmt.GetRenameStmt() != nil {
			add("rename", "RENAME can break code using the old name")
		}
		if stmt.GetTruncateStmt() != nil {
			add("truncate", "TRUNCATE takes an exclusive table lock")
		}
		if stmt.GetClusterStmt() != nil || vacuumFull(stmt.GetVacuumStmt()) {
			add("rewrite-command", "CLUSTER or VACUUM FULL rewrites table storage")
		}
	}
	return findings, nil
}

func checkAlter(cmd *pg_query.AlterTableCmd, add func(string, string)) {
	if cmd == nil {
		return
	}
	switch cmd.GetSubtype() {
	case pg_query.AlterTableType_AT_AddColumn:
		column := cmd.GetDef().GetColumnDef()
		notNull, filled := false, serial(column.GetTypeName())
		for _, node := range column.GetConstraints() {
			constraint := node.GetConstraint()
			switch constraint.GetContype() {
			case pg_query.ConstrType_CONSTR_NOTNULL, pg_query.ConstrType_CONSTR_PRIMARY:
				notNull = true
			case pg_query.ConstrType_CONSTR_DEFAULT, pg_query.ConstrType_CONSTR_IDENTITY, pg_query.ConstrType_CONSTR_GENERATED:
				filled = true
			}
			checkConstraint(constraint, add)
		}
		if notNull && !filled {
			add("add-column-not-null-without-default", "ADD COLUMN NOT NULL without a default breaks inserts from application versions that omit the column")
		}
	case pg_query.AlterTableType_AT_AlterColumnType:
		add("alter-column-type", "ALTER COLUMN TYPE can rewrite the table")
	case pg_query.AlterTableType_AT_AddConstraint:
		checkConstraint(cmd.GetDef().GetConstraint(), add)
	case pg_query.AlterTableType_AT_SetNotNull:
		add("set-not-null", "SET NOT NULL can scan the table")
	case pg_query.AlterTableType_AT_DropColumn:
		add("drop-column", "DROP COLUMN can break code using the old column")
	}
}

func checkConstraint(constraint *pg_query.Constraint, add func(string, string)) {
	switch constraint.GetContype() {
	case pg_query.ConstrType_CONSTR_DEFAULT:
		if !literal(constraint.GetRawExpr()) {
			add("add-column-expression-default", "ADD COLUMN with an expression default can rewrite existing rows")
		}
	case pg_query.ConstrType_CONSTR_FOREIGN:
		if !constraint.GetSkipValidation() {
			add("foreign-key-validates-immediately", "adding a foreign key without NOT VALID scans existing rows")
		}
	case pg_query.ConstrType_CONSTR_UNIQUE:
		if constraint.GetIndexname() == "" {
			add("unique-constraint", "adding UNIQUE directly builds a backing index")
		}
	}
}

func serial(typeName *pg_query.TypeName) bool {
	names := typeName.GetNames()
	return len(names) == 1 && slices.Contains([]string{"smallserial", "serial", "bigserial", "serial2", "serial4", "serial8"}, names[0].GetString_().GetSval())
}

func literal(node *pg_query.Node) bool {
	if node.GetAConst() != nil {
		return true
	}
	if cast := node.GetTypeCast(); cast != nil {
		return literal(cast.GetArg())
	}
	return false
}

func vacuumFull(stmt *pg_query.VacuumStmt) bool {
	for _, node := range stmt.GetOptions() {
		option := node.GetDefElem()
		if option.GetDefname() != "full" {
			continue
		}
		arg := option.GetArg()
		if value := arg.GetString_(); value != nil {
			return !strings.EqualFold(value.GetSval(), "false") && !strings.EqualFold(value.GetSval(), "off")
		}
		if value := arg.GetInteger(); value != nil {
			return value.GetIval() != 0
		}
		return true
	}
	return false
}
