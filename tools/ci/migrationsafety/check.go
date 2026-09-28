package migrationsafety

import (
	"fmt"
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
	var findings []Finding
	for _, raw := range parsed.Stmts {
		start := min(int(raw.StmtLocation), len(sql))
		for start < len(sql) && strings.ContainsRune(" \r\n\t", rune(sql[start])) {
			start++
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
		for _, node := range cmd.GetDef().GetColumnDef().GetConstraints() {
			constraint := node.GetConstraint()
			if constraint.GetContype() == pg_query.ConstrType_CONSTR_DEFAULT && !literal(constraint.GetRawExpr()) {
				add("add-column-expression-default", "ADD COLUMN with an expression default can rewrite existing rows")
			}
		}
	case pg_query.AlterTableType_AT_AlterColumnType:
		add("alter-column-type", "ALTER COLUMN TYPE can rewrite the table")
	case pg_query.AlterTableType_AT_AddConstraint:
		constraint := cmd.GetDef().GetConstraint()
		switch constraint.GetContype() {
		case pg_query.ConstrType_CONSTR_FOREIGN:
			if !constraint.GetSkipValidation() {
				add("foreign-key-validates-immediately", "adding a foreign key without NOT VALID scans existing rows")
			}
		case pg_query.ConstrType_CONSTR_UNIQUE:
			if constraint.GetIndexname() == "" {
				add("unique-constraint", "adding UNIQUE directly builds a backing index")
			}
		}
	case pg_query.AlterTableType_AT_SetNotNull:
		add("set-not-null", "SET NOT NULL can scan the table")
	case pg_query.AlterTableType_AT_DropColumn:
		add("drop-column", "DROP COLUMN can break code using the old column")
	}
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
	if stmt == nil {
		return false
	}
	for _, node := range stmt.GetOptions() {
		option := node.GetDefElem()
		if option.GetDefname() == "full" {
			value := option.GetArg().GetAConst()
			return value == nil || value.GetBoolval().GetBoolval()
		}
	}
	return false
}
