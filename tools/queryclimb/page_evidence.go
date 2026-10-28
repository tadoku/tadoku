package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
)

func estimateAtSix(
	ctx context.Context,
	conn *pgx.Conn,
	q query,
	c caseSpec,
	executeSQL string,
	measured metrics,
) (json.RawMessage, error) {
	const name = "qc_estimate"
	if err := conn.Deallocate(ctx, name); err != nil {
		return nil, err
	}
	if _, err := conn.Prepare(ctx, name, q.SQL); err != nil {
		return nil, err
	}
	for i := 0; i < 5; i++ {
		warm := c
		if len(c.WarmArgs) != 0 {
			warm.Args = c.WarmArgs[i]
		}
		err := scoped(ctx, conn, c.Tenant, func() error {
			_, _, err := executeNamed(ctx, conn, warm, name)
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	var plan json.RawMessage
	sql := "explain (verbose, costs, format json) " + strings.Replace(executeSQL, "qc_query", name, 1)
	err := scoped(ctx, conn, c.Tenant, func() error {
		return conn.QueryRow(ctx, sql, pgx.QueryExecModeSimpleProtocol).Scan(&plan)
	})
	if err != nil {
		return nil, err
	}
	var generic, custom int64
	err = conn.QueryRow(ctx, `select generic_plans, custom_plans
		from pg_prepared_statements
		where name = $1`, name).Scan(&generic, &custom)
	if err != nil {
		return nil, err
	}
	if generic+custom != 6 || generic != measured.GenericPlans || custom != measured.CustomPlans {
		return nil, errors.New("estimated plan did not reproduce execution-six plan selection")
	}
	return plan, nil
}

func pageEvidence(ctx context.Context, root string, args []string) error {
	flags := flag.NewFlagSet("page-evidence", flag.ContinueOnError)
	database := flags.String("name", "", "owned baseline database")
	casesPath := flags.String("cases", "", "original boundary-cut cases with include_rows enabled")
	offsetArg := flags.Int("offset-arg", -1, "zero-based OFFSET argument position")
	limitArg := flags.Int("limit-arg", -1, "zero-based LIMIT argument position")
	out := flags.String("output", "", "baseline full-row oracle JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *offsetArg < 0 || *limitArg < 0 || *offsetArg == *limitArg || *out == "" {
		return errors.New("page evidence needs distinct pagination argument positions and output")
	}
	queries, err := inventory(root)
	if err != nil {
		return err
	}
	cases, err := readCases(*casesPath)
	if err != nil {
		return err
	}
	conn, err := connect(ctx, localDSN, *database, "tadoku")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	oracles := []pageOracle{}
	for _, c := range cases {
		q, err := findQuery(queries, c.Query)
		if err != nil {
			return err
		}
		if q.Write || len(c.MaskColumns) != 0 || len(c.TieColumns) == 0 ||
			*offsetArg >= len(c.Args) || *limitArg >= len(c.Args) {
			return fmt.Errorf("%s/%s: oracle needs an unmasked ordered read and pagination arguments", c.Query, c.Name)
		}
		if err := conn.Deallocate(ctx, "qc_query"); err != nil {
			return err
		}
		if _, err := conn.Prepare(ctx, "qc_query", q.SQL); err != nil {
			return err
		}
		all := c
		all.Args = slices.Clone(c.Args)
		zero, unlimited := "0", "2147483647"
		all.Args[*offsetArg], all.Args[*limitArg] = &zero, &unlimited
		var rows [][]*string
		err = scoped(ctx, conn, c.Tenant, func() error {
			rows, _, err = execute(ctx, conn, all)
			return err
		})
		if err != nil {
			return err
		}
		if len(rows) >= 2147483647 {
			return errors.New("oracle was truncated by the pagination bound")
		}
		oracles = append(oracles, pageOracle{
			Query:      c.Query,
			Case:       c.Name,
			Tenant:     c.Tenant,
			Args:       c.Args,
			TieColumns: c.TieColumns,
			OffsetArg:  *offsetArg,
			LimitArg:   *limitArg,
			SQLHash:    digest(q.SQL),
			Rows:       rows,
		})
	}
	return writeJSON(*out, oracles)
}
