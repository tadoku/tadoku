package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

func generatedInterface(root string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(
		filepath.Join(root, "services/tadoku-api/generated/sqlc"),
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
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
						text, err := strconv.Unquote(literal.Value)
						if err != nil {
							return err
						}
						if namePattern.MatchString(text) {
							literal.Value = "`<query SQL>`"
						}
					}
				}
			}
			var b bytes.Buffer
			if err := format.Node(&b, fset, file); err != nil {
				return err
			}
			rel, _ := filepath.Rel(root, path)
			files[rel] = b.String()
			return nil
		},
	)
	return files, err
}
