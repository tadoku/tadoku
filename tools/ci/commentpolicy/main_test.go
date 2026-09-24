package main

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestAnalyze(t *testing.T) {
	source := `// Package sample merely repeats its name.
package sample

//go:generate echo keep

// Use starts a task. Call Close after it finishes.
func Use() { _ = "// not a comment" }

//nolint:example
func Suppressed() {}

func test() {
	// A narration comment.

	// Test safety: only use disposable state.
	println("ready") // An inline comment.
}
`
	file, err := parser.ParseFile(token.NewFileSet(), "sample.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(analyze(file, true)); got != 4 {
		t.Fatalf("test code violations=%d, want 4", got)
	}
	if got := len(analyze(file, false)); got != 5 {
		t.Fatalf("application code violations=%d, want 5", got)
	}
}
