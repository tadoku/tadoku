// Command depolicy runs the upstream import analyzer over the native API tree.
// It intentionally parses every Go file, including tests and inactive build tags;
// compilation/type checking remains the responsibility of the normal Bazel build.
package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/satorunooshie/depolicy"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Set by Bazel. Depolicy uses go/build to distinguish stdlib from external imports.
var sdkRootFile string

func main() {
	if err := check(os.Getenv("BUILD_WORKSPACE_DIRECTORY"), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func check(root string, out io.Writer) error {
	if root == "" {
		return fmt.Errorf("run with bazel run //tools/ci/depolicy")
	}
	sdk, err := runfiles.Rlocation(sdkRootFile)
	if err != nil {
		return fmt.Errorf("locate Bazel Go SDK: %w", err)
	}
	build.Default.GOROOT = filepath.Dir(sdk)
	if !depolicy.IsStandardPackage("fmt") {
		return fmt.Errorf("Bazel Go SDK is missing standard-library sources")
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	configPath := filepath.Join(root, depolicy.ConfigFileName)
	// The upstream analyzer skips missing configuration; this gate must not.
	config, err := depolicy.LoadProjectConfig(configPath)
	if err != nil {
		return err
	}
	if config.Module.RootDir != root || config.Module.Path != "github.com/tadoku/tadoku" {
		return fmt.Errorf("expected Tadoku go.mod next to %s", configPath)
	}
	var files, tests, violations int
	err = filepath.WalkDir(filepath.Join(root, "services/tadoku-api"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("unexpected symlink in native source tree: %s", path)
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Reject nested configuration that could change the analyzer's scope.
		found, err := depolicy.FindConfigFromFiles(path)
		if err != nil {
			return err
		}
		if found != configPath {
			return fmt.Errorf("expected %s, found nested configuration %s", configPath, found)
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		pkgPath := config.Module.Path + "/" + filepath.ToSlash(rel)
		if strings.HasSuffix(path, "_test.go") {
			tests++
			if strings.HasSuffix(file.Name.Name, "_test") {
				// Keep external tests disjoint from feature-name captures. They get
				// a named /_test policy, not a blanket test exemption.
				pkgPath += "/_test"
			}
		}
		pass := &analysis.Pass{
			Analyzer: depolicy.Analyzer,
			Fset:     fset,
			Files:    []*ast.File{file},
			Pkg:      types.NewPackage(pkgPath, file.Name.Name),
			ResultOf: make(map[*analysis.Analyzer]any),
			Report: func(d analysis.Diagnostic) {
				violations++
				fmt.Fprintf(out, "%s: %s: %s\n", fset.Position(d.Pos), d.Category, d.Message)
			},
		}
		pass.ResultOf[inspect.Analyzer], err = inspect.Analyzer.Run(pass)
		if err != nil {
			return err
		}
		if _, err := depolicy.Analyzer.Run(pass); err != nil {
			return err
		}
		files++
		return nil
	})
	if err != nil {
		return err
	}
	if files == tests || tests == 0 {
		return fmt.Errorf("expected native source and test files; checked %d files, %d tests", files, tests)
	}
	fmt.Fprintf(out, "Depolicy checked %d Go files (%d test files); %d violations\n", files, tests, violations)
	if violations != 0 {
		return fmt.Errorf("native import policy failed")
	}
	return nil
}
