// Command repopolicy enforces the Tadoku API repository conventions on every
// handwritten *_repository.go file: at most one database statement per
// function, no ID allocation and no transaction control. It checks syntax only;
// compilation/type checking remains the responsibility of the normal Bazel build.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	sqlcImportPrefix = "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/"
	convention       = `see AGENTS.md "Keep each repository method to one SQL statement"`
)

var (
	executorMethods = map[string]bool{
		"CopyFrom":  true,
		"Exec":      true,
		"Query":     true,
		"QueryRow":  true,
		"SendBatch": true,
	}
	transactionCalls = map[string]bool{
		"Begin":            true,
		"BeginFunc":        true,
		"BeginTx":          true,
		"BeginTxFunc":      true,
		"Commit":           true,
		"Rollback":         true,
		"RunInTransaction": true,
	}
)

type finding struct {
	pos     token.Pos
	rule    string
	message string
}

func main() {
	if err := check(os.Getenv("BUILD_WORKSPACE_DIRECTORY"), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func check(root string, out io.Writer) error {
	if root == "" {
		return fmt.Errorf("run with bazel run //tools/ci/repopolicy")
	}

	var files, violations int
	err := filepath.WalkDir(filepath.Join(root, "services/tadoku-api"), func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(filePath, "_repository.go") {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		if ast.IsGenerated(file) {
			return nil
		}

		rel, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		for _, f := range analyze(file) {
			violations++
			fmt.Fprintf(out, "%s:%d: %s: %s\n", rel, fset.Position(f.pos).Line, f.rule, f.message)
		}
		files++
		return nil
	})
	if err != nil {
		return err
	}

	if files == 0 {
		return fmt.Errorf("expected Tadoku API repository files under %s", root)
	}
	fmt.Fprintf(out, "Repopolicy checked %d repository files; %d violations\n", files, violations)
	if violations != 0 {
		return fmt.Errorf("Tadoku API repository policy failed")
	}
	return nil
}

// analyze reports repository policy violations in one parsed file.
func analyze(file *ast.File) []finding {
	a := fileAnalysis{
		sqlc:         map[string]bool{},
		uuid:         map[string]bool{},
		constructors: map[string]bool{},
	}
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		name := path.Base(importPath)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		switch {
		case strings.HasPrefix(importPath, sqlcImportPrefix):
			a.sqlc[name] = true
		case path.Base(importPath) == "uuid":
			a.uuid[name] = true
		}
	}

	// Local helpers that return generated queries, such as r.queries(ctx).
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Type.Results != nil && len(fn.Type.Results.List) > 0 && a.isQueriesType(fn.Type.Results.List[0].Type) {
			a.constructors[fn.Name.Name] = true
		}
	}

	var findings []finding
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			findings = append(findings, a.analyzeFunc(fn)...)
		}
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].pos < findings[j].pos })
	return findings
}

type fileAnalysis struct {
	sqlc         map[string]bool
	uuid         map[string]bool
	constructors map[string]bool
	queryVars    map[string]bool
}

func (a *fileAnalysis) analyzeFunc(fn *ast.FuncDecl) []finding {
	a.queryVars = map[string]bool{}
	for _, field := range fn.Type.Params.List {
		if a.isQueriesType(field.Type) {
			for _, name := range field.Names {
				a.queryVars[name.Name] = true
			}
		}
	}

	var findings []finding
	var statements []token.Pos
	var stack []ast.Node
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		stack = append(stack, n)

		switch n := n.(type) {
		case *ast.AssignStmt:
			a.bind(n.Lhs, n.Rhs)
		case *ast.ValueSpec:
			names := make([]ast.Expr, len(n.Names))
			for i, name := range n.Names {
				names[i] = name
			}
			a.bind(names, n.Values)
			if n.Type != nil && a.isQueriesType(n.Type) {
				for _, name := range n.Names {
					a.queryVars[name.Name] = true
				}
			}
		case *ast.CallExpr:
			sel, ok := ast.Unparen(n.Fun).(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch {
			case transactionCalls[sel.Sel.Name]:
				findings = append(findings, finding{
					pos:     n.Pos(),
					rule:    "transaction-control",
					message: fmt.Sprintf("%s must not call %s; repositories receive an executor and the application owns transactions (%s)", fn.Name.Name, sel.Sel.Name, convention),
				})
			case isPackage(sel.X, a.uuid) && strings.HasPrefix(sel.Sel.Name, "New"):
				findings = append(findings, finding{
					pos:     n.Pos(),
					rule:    "id-allocation",
					message: fmt.Sprintf("%s must not allocate IDs; the feature service allocates them and passes them in (%s)", fn.Name.Name, convention),
				})
			case executorMethods[sel.Sel.Name] || (sel.Sel.Name != "WithTx" && a.isQueries(sel.X)):
				statements = append(statements, n.Pos())
				if inLoop(stack) {
					statements = append(statements, n.Pos())
				}
			}
		}
		return true
	})

	if len(statements) > 1 {
		findings = append(findings, finding{
			pos:     statements[1],
			rule:    "single-statement",
			message: fmt.Sprintf("%s issues more than one database statement (statements in loops count repeatedly); compose statements in the feature service or application layer (%s)", fn.Name.Name, convention),
		})
	}
	return findings
}

// bind tracks local variables holding generated queries.
func (a *fileAnalysis) bind(lhs, rhs []ast.Expr) {
	for i, target := range lhs {
		ident, ok := target.(*ast.Ident)
		if !ok {
			continue
		}
		var value ast.Expr
		switch {
		case len(lhs) == len(rhs):
			value = rhs[i]
		case len(rhs) == 1 && i == 0:
			value = rhs[0]
		default:
			continue
		}
		if a.isQueries(value) {
			a.queryVars[ident.Name] = true
		} else {
			delete(a.queryVars, ident.Name)
		}
	}
}

// isQueries reports whether expr evaluates to generated sqlc queries.
func (a *fileAnalysis) isQueries(expr ast.Expr) bool {
	switch e := ast.Unparen(expr).(type) {
	case *ast.Ident:
		return a.queryVars[e.Name]
	case *ast.CallExpr:
		switch fun := ast.Unparen(e.Fun).(type) {
		case *ast.Ident:
			return a.constructors[fun.Name]
		case *ast.SelectorExpr:
			switch {
			case isPackage(fun.X, a.sqlc):
				return fun.Sel.Name == "New"
			case fun.Sel.Name == "WithTx":
				return a.isQueries(fun.X)
			default:
				return a.constructors[fun.Sel.Name]
			}
		}
	}
	return false
}

func (a *fileAnalysis) isQueriesType(expr ast.Expr) bool {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	sel, ok := expr.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Queries" && isPackage(sel.X, a.sqlc)
}

func isPackage(expr ast.Expr, names map[string]bool) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && names[ident.Name]
}

// inLoop reports whether the innermost node of stack runs repeatedly in a loop.
func inLoop(stack []ast.Node) bool {
	node := stack[len(stack)-1]
	for _, ancestor := range stack[:len(stack)-1] {
		switch loop := ancestor.(type) {
		case *ast.ForStmt:
			if contains(loop.Cond, node) || contains(loop.Post, node) || contains(loop.Body, node) {
				return true
			}
		case *ast.RangeStmt:
			if contains(loop.Body, node) {
				return true
			}
		}
	}
	return false
}

func contains(outer, inner ast.Node) bool {
	return outer != nil && outer.Pos() <= inner.Pos() && inner.End() <= outer.End()
}
