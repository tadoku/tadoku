package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tadoku/tadoku/tools/ci/migrationsafety"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrationsafety <migration.up.sql>...")
		os.Exit(2)
	}
	failed := false
	for _, path := range os.Args[1:] {
		input := path
		if root := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); root != "" && !filepath.IsAbs(path) {
			input = filepath.Join(root, path)
		}
		sql, err := os.ReadFile(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			failed = true
			continue
		}
		findings, err := migrationsafety.Check(string(sql))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			failed = true
			continue
		}
		for _, finding := range findings {
			fmt.Fprintf(os.Stderr, "%s:%d: %s: %s\n", path, finding.Line, finding.Rule, finding.Message)
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}
