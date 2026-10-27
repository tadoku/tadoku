package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type query struct {
	Key         string   `json:"key"`
	Source      string   `json:"source"`
	Generated   string   `json:"generated"`
	Command     string   `json:"command"`
	SQL         string   `json:"sql"`
	Fingerprint string   `json:"fingerprint"`
	Tables      []string `json:"tables"`
	Families    []string `json:"families"`
	Mutations   []string `json:"mutations"`
	Write       bool     `json:"write"`
}

var namePattern = regexp.MustCompile(`(?m)^-- name: (\w+) (:\w+)\s*\n`)
var macroPattern = regexp.MustCompile(`sqlc\.(?:narg|arg)\('?(\w+)'?\)|@(\w+)`)

func inventory(root string) ([]query, error) {
	generated := map[string]string{}
	generatedPaths := map[string]string{}
	err := filepath.WalkDir(
		filepath.Join(root, "services/tadoku-api/generated/sqlc"),
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.CONST {
					continue
				}
				for _, spec := range gen.Specs {
					for _, value := range spec.(*ast.ValueSpec).Values {
						literal, ok := value.(*ast.BasicLit)
						if !ok || literal.Kind != token.STRING {
							continue
						}
						sql, err := strconv.Unquote(literal.Value)
						if err != nil {
							return err
						}
						match := namePattern.FindStringSubmatch(sql)
						if match == nil {
							continue
						}
						rel, _ := filepath.Rel(filepath.Join(root, "services/tadoku-api/generated/sqlc"), filepath.Dir(path))
						if rel == filepath.Join("identity", "reads") {
							rel = "identityreads"
						}
						key := rel + "." + match[1]
						if _, exists := generated[key]; exists {
							return fmt.Errorf("duplicate constant %s", key)
						}
						generated[key] = sql
						generatedPaths[key], _ = filepath.Rel(root, path)
					}
				}
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	var queries []query
	err = filepath.WalkDir(
		filepath.Join(root, "services/tadoku-api/sql"),
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".sql") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := string(content)
			matches := namePattern.FindAllStringSubmatchIndex(text, -1)
			for i, match := range matches {
				name, command := text[match[2]:match[3]], text[match[4]:match[5]]
				key := filepath.Base(filepath.Dir(path)) + "." + name
				end := len(text)
				if i+1 < len(matches) {
					end = matches[i+1][0]
				}
				sourceSQL := strings.TrimSpace(text[match[1]:end])
				sql, exists := generated[key]
				if !exists {
					return fmt.Errorf("no generated SQL for %s", key)
				}
				if err := validateSource(sourceSQL); err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				fingerprint, err := pg_query.Fingerprint(sql)
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				tree, err := pg_query.ParseToJSON(sql)
				if err != nil {
					return err
				}
				var parsed any
				if err := json.Unmarshal([]byte(tree), &parsed); err != nil {
					return err
				}
				tables, ctes, locks := map[string]bool{}, map[string]bool{}, map[string]bool{}
				write := false
				mutations := map[string]bool{}
				walkJSON(parsed, func(m map[string]any) {
					if locking, ok := m["LockingClause"]; ok {
						walkJSON(locking, func(node map[string]any) {
							if name, ok := node["relname"].(string); ok {
								locks[name] = true
							}
						})
					}
					if table, ok := m["relname"].(string); ok {
						tables[table] = true
					}
					if c, ok := m["CommonTableExpr"].(map[string]any); ok {
						ctes[c["ctename"].(string)] = true
					}
					for _, kind := range []string{"InsertStmt", "UpdateStmt", "DeleteStmt", "MergeStmt"} {
						if statement, ok := m[kind].(map[string]any); ok {
							write = true
							if relation, ok := statement["relation"].(map[string]any); ok {
								if table, ok := relation["relname"].(string); ok {
									mutations[table] = true
								}
							}
						}
					}
				})
				tableNames := []string{}
				for t := range tables {
					if !ctes[t] && !locks[t] {
						tableNames = append(tableNames, t)
					}
				}
				slices.Sort(tableNames)
				mutationNames := []string{}
				for name := range mutations {
					mutationNames = append(mutationNames, name)
				}
				slices.Sort(mutationNames)
				relative, _ := filepath.Rel(root, path)
				queries = append(queries, query{
					Key: key, Source: relative, Generated: generatedPaths[key], Command: command,
					SQL: sql, Fingerprint: fingerprint, Tables: tableNames,
					Families: families(parsed), Write: write, Mutations: mutationNames,
				})
				delete(generated, key)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	if len(generated) > 0 {
		return nil, fmt.Errorf("generated queries without source: %v", generated)
	}
	slices.SortFunc(queries, func(a, b query) int { return strings.Compare(a.Key, b.Key) })
	return queries, nil
}

func validateSource(source string) error {
	names := map[string]int{}
	lowered := macroPattern.ReplaceAllStringFunc(source, func(macro string) string {
		match := macroPattern.FindStringSubmatch(macro)
		name := match[1]
		if name == "" {
			name = match[2]
		}
		if names[name] == 0 {
			names[name] = len(names) + 1
		}
		return "$" + strconv.Itoa(names[name])
	},
	)
	_, err := pg_query.Parse(lowered)
	return err
}

func walkJSON(value any, visit func(map[string]any)) {
	switch x := value.(type) {
	case map[string]any:
		visit(x)
		for _, v := range x {
			walkJSON(v, visit)
		}
	case []any:
		for _, v := range x {
			walkJSON(v, visit)
		}
	}
}

func families(parsed any) []string {
	matches := map[string]bool{}
	count, limit, tags := false, false, false
	walkJSON(parsed, func(m map[string]any) {
		if m["relname"] == "log_tags" {
			tags = true
		}
		if m["limitCount"] != nil {
			limit = true
		}
		if function, ok := m["FuncCall"].(map[string]any); ok {
			name := ""
			walkJSON(function["funcname"], func(part map[string]any) {
				if value, ok := part["sval"].(string); ok {
					name = value
				}
			})
			switch name {
			case "count":
				count = true
			case "array_agg":
				matches["array-aggregate"] = true
			case "rank":
				matches["ranking"] = function["over"] != nil
			case "sum":
				walkJSON(function["args"], func(part map[string]any) {
					if part["CoalesceExpr"] != nil {
						matches["score-sum"] = true
					}
				})
			}
		}
		if boolean, ok := m["BoolExpr"].(map[string]any); ok && boolean["boolop"] == "OR_EXPR" {
			parameter, nullTest := false, false
			walkJSON(boolean["args"], func(part map[string]any) {
				parameter = parameter || part["ParamRef"] != nil
				nullTest = nullTest || part["NullTest"] != nil
			})
			matches["optional-filter"] = matches["optional-filter"] || (parameter && nullTest)
		}
		if locking, ok := m["LockingClause"].(map[string]any); ok && locking["waitPolicy"] == "LockWaitSkip" {
			matches["job-claim"] = true
		}
	},
	)
	matches["correlated-tags"] = tags && matches["array-aggregate"]
	delete(matches, "array-aggregate")
	matches["count-and-page"] = count && limit
	names := []string{}
	for name, matched := range matches {
		if matched {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		names = append(names, "other")
	}
	slices.Sort(names)
	return names
}
