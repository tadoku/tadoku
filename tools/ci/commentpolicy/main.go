package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := check(os.Getenv("BUILD_WORKSPACE_DIRECTORY"), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func check(root string, out io.Writer) error {
	if root == "" {
		return fmt.Errorf("run with bazel run //tools/ci/commentpolicy")
	}
	var files, violations int
	for _, dir := range []string{"services", "tools"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if entry.Name() == "generated" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(name, ".go") {
				return nil
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			if ast.IsGenerated(file) {
				return nil
			}
			files++
			for _, comment := range analyze(file, strings.HasSuffix(name, "_test.go") || strings.Contains(name, "/internal/test")) {
				violations++
				rel, err := filepath.Rel(root, name)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s:%d: Go comment must document declaration usage or begin // Test safety: in test code\n", rel, fset.Position(comment.Pos()).Line)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	fmt.Fprintf(out, "Commentpolicy checked %d handwritten Go files; %d violations\n", files, violations)
	if violations != 0 {
		return fmt.Errorf("Go comment policy failed")
	}
	return nil
}

func analyze(file *ast.File, testCode bool) []*ast.Comment {
	docs := map[*ast.CommentGroup]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		switch decl := node.(type) {
		case *ast.FuncDecl:
			docs[decl.Doc] = true
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				docs[decl.Doc] = true
			}
		case *ast.TypeSpec:
			docs[decl.Doc] = true
		case *ast.Field:
			docs[decl.Doc] = true
		}
		return true
	})
	var violations []*ast.Comment
	for _, group := range file.Comments {
		allowed := docs[group] || testCode && strings.HasPrefix(group.List[0].Text, "// Test safety:")
		for _, comment := range group.List {
			suppression := strings.HasPrefix(comment.Text, "//nolint") || strings.HasPrefix(comment.Text, "//lint:ignore")
			if suppression || !allowed && !strings.HasPrefix(comment.Text, "//go:") {
				violations = append(violations, comment)
			}
		}
	}
	return violations
}
