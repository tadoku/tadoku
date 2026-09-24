package main

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"
)

const fixtureHeader = `package pages

import (
	"context"

	guuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	pagesdb "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	other "github.com/tadoku/tadoku/services/tadoku-api/internal/other"
)

type Repository struct{ db *pgxpool.Pool }

func (r *Repository) queries(ctx context.Context) (*pagesdb.Queries, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	return pagesdb.New(executor), nil
}
`

func TestAnalyze(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{
			name: "one aliased inline query with row mapping loop",
			source: `
func (r *Repository) List(ctx context.Context, executor postgres.DBTX) ([]string, error) {
	rows, err := pagesdb.New(executor).ListPages(ctx)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, row := range rows {
		ids = append(ids, row.ID.String())
	}
	other.New(executor).Unrelated(ctx)
	_ = guuid.MustParse("00000000-0000-0000-0000-000000000000")
	return ids, nil
}`,
		},
		{
			name: "two inline queries",
			source: `
func (r *Repository) Create(ctx context.Context, executor postgres.DBTX) error {
	if err := pagesdb.New(executor).CreatePage(ctx); err != nil {
		return err
	}
	return pagesdb.New(executor).CreatePageContent(ctx)
}`,
			want: []string{"single-statement"},
		},
		{
			name: "one query through local variable",
			source: `
func (r *Repository) Find(ctx context.Context, executor postgres.DBTX) error {
	q := pagesdb.New(executor)
	_, err := q.FindPage(ctx)
	return err
}`,
		},
		{
			name: "two queries through local variable",
			source: `
func (r *Repository) Find(ctx context.Context, executor postgres.DBTX) error {
	var q = pagesdb.New(executor)
	if _, err := q.FindPage(ctx); err != nil {
		return err
	}
	_, err := q.FindPageContent(ctx)
	return err
}`,
			want: []string{"single-statement"},
		},
		{
			name: "one query through helper",
			source: `
func (r *Repository) Find(ctx context.Context) error {
	q, err := r.queries(ctx)
	if err != nil {
		return err
	}
	_, err = q.FindPage(ctx)
	return err
}`,
		},
		{
			name: "two queries through helper",
			source: `
func (r *Repository) Find(ctx context.Context) error {
	q, err := r.queries(ctx)
	if err != nil {
		return err
	}
	if _, err := q.FindPage(ctx); err != nil {
		return err
	}
	return q.TouchPage(ctx)
}`,
			want: []string{"single-statement"},
		},
		{
			name: "query in loop",
			source: `
func (r *Repository) CreateAll(ctx context.Context, executor postgres.DBTX, ids []string) error {
	q := pagesdb.New(executor)
	for range ids {
		if err := q.CreatePage(ctx); err != nil {
			return err
		}
	}
	return nil
}`,
			want: []string{"single-statement"},
		},
		{
			name: "queries in exclusive branches",
			source: `
func (r *Repository) Save(ctx context.Context, executor postgres.DBTX, create bool) error {
	if create {
		return pagesdb.New(executor).CreatePage(ctx)
	}
	return pagesdb.New(executor).UpdatePage(ctx)
}`,
			want: []string{"single-statement"},
		},
		{
			name: "one direct executor statement",
			source: `
func (r *Repository) Touch(ctx context.Context, executor postgres.DBTX) error {
	_, err := executor.Exec(ctx, "select 1")
	return err
}`,
		},
		{
			name: "direct executor statement and query",
			source: `
func (r *Repository) Touch(ctx context.Context, executor postgres.DBTX) error {
	if err := executor.QueryRow(ctx, "select 1").Scan(); err != nil {
		return err
	}
	return pagesdb.New(executor).TouchPage(ctx)
}`,
			want: []string{"single-statement"},
		},
		{
			name: "ID allocation",
			source: `
func (r *Repository) Create(ctx context.Context, executor postgres.DBTX) error {
	id := guuid.New()
	contentID, err := guuid.NewV7()
	if err != nil {
		return err
	}
	return pagesdb.New(executor).CreatePage(ctx, id, contentID)
}`,
			want: []string{"id-allocation", "id-allocation"},
		},
		{
			name: "transaction control",
			source: `
func (r *Repository) Create(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := postgres.RunInTransaction(ctx, r.db, nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}`,
			want: []string{"transaction-control", "transaction-control", "transaction-control", "transaction-control"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), "pages_repository.go", fixtureHeader+tc.source, parser.ParseComments)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}

			var got []string
			for _, f := range analyze(file) {
				got = append(got, f.rule)
			}

			if !slices.Equal(got, tc.want) {
				t.Errorf("rules = %q, want %q", got, tc.want)
			}
		})
	}
}
