package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

var mainBaselinePaths = []string{
	"services/tadoku-api/migrations",
	"services/tadoku-api/sql",
	"services/tadoku-api/generated/sqlc",
}

type migrationFile struct {
	name string
	body []byte
}

type migrationPlan struct {
	candidateOnly []migrationFile
}

// planAttemptMigrations also writes the plan to dir/migrations.json.
func planAttemptMigrations(
	ctx context.Context,
	root, candidate string,
	templates []string,
	dir string,
) (migrationPlan, error) {
	if err := requireMainBaseline(ctx, root); err != nil {
		return migrationPlan{}, err
	}

	baseline, err := readMigrations(root)
	if err != nil {
		return migrationPlan{}, err
	}
	expected := map[string]string{}
	for name, body := range baseline {
		if strings.HasSuffix(name, ".up.sql") {
			expected[name] = digest(string(body))
		}
	}
	for _, template := range templates {
		if err := requireTemplateManifest(ctx, template, expected); err != nil {
			return migrationPlan{}, err
		}
	}

	candidateFiles, err := readMigrations(candidate)
	if err != nil {
		return migrationPlan{}, err
	}
	candidateOnly, err := candidateOnlyMigrations(baseline, candidateFiles)
	if err != nil {
		return migrationPlan{}, err
	}

	baselineRevision, err := revision(root)
	if err != nil {
		return migrationPlan{}, err
	}
	candidateRevision, err := revision(candidate)
	if err != nil {
		return migrationPlan{}, err
	}
	manifest, err := json.Marshal(expected)
	if err != nil {
		return migrationPlan{}, err
	}
	names := []string{}
	for _, file := range candidateOnly {
		names = append(names, file.name)
	}
	err = writeJSON(filepath.Join(dir, "migrations.json"), map[string]any{
		"baseline_revision":        baselineRevision,
		"candidate_revision":       candidateRevision,
		"template_manifest_sha256": digest(string(manifest)),
		"candidate_only":           names,
	})
	return migrationPlan{candidateOnly: candidateOnly}, err
}

func (plan migrationPlan) applyToCandidate(ctx context.Context, database string) error {
	if len(plan.candidateOnly) == 0 {
		return nil
	}
	owner, err := connect(ctx, localDSN, database, "tadoku_owner")
	if err != nil {
		return err
	}
	defer owner.Close(ctx)

	for _, file := range plan.candidateOnly {
		if _, err := owner.Exec(ctx, string(file.body)); err != nil {
			return fmt.Errorf("migration %s: %w", file.name, err)
		}
	}
	return nil
}

// requireMainBaselineFor checks root against origin/main only when candidate is a different checkout.
func requireMainBaselineFor(ctx context.Context, root, candidate string) error {
	a, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	b, err := filepath.Abs(candidate)
	if err != nil {
		return err
	}
	if a == b {
		return nil
	}
	return requireMainBaseline(ctx, a)
}

func requireMainBaseline(ctx context.Context, root string) error {
	if _, err := git(ctx, root, "fetch", "--quiet", "origin", "main"); err != nil {
		return fmt.Errorf("baseline must match origin/main: fetch failed: %w", err)
	}

	for _, path := range mainBaselinePaths {
		head, err := git(ctx, root, "rev-parse", "HEAD:"+path)
		if err != nil {
			return fmt.Errorf("baseline must match origin/main: %s at HEAD: %w", path, err)
		}
		main, err := git(ctx, root, "rev-parse", "origin/main:"+path)
		if err != nil {
			return fmt.Errorf("baseline must match origin/main: %s at origin/main: %w", path, err)
		}
		if head != main {
			return fmt.Errorf("baseline must match origin/main: %s differs", path)
		}
	}

	status, err := git(ctx, root, append([]string{"status", "--porcelain", "--untracked-files=all", "--"},
		mainBaselinePaths...)...)
	if err != nil {
		return fmt.Errorf("baseline must match origin/main: status failed: %w", err)
	}
	if status != "" {
		return fmt.Errorf("baseline must match origin/main: uncommitted changes: %s", strings.Split(status, "\n")[0])
	}
	return nil
}

func git(ctx context.Context, root string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "git", args...)
	command.Dir = root
	out, err := command.Output()
	if exit, ok := err.(*exec.ExitError); ok {
		return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(string(exit.Stderr)))
	}
	return strings.TrimSpace(string(out)), err
}

func readMigrations(root string) (map[string][]byte, error) {
	paths, err := filepath.Glob(filepath.Join(root, "services/tadoku-api/migrations/*.sql"))
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		files[filepath.Base(path)] = body
	}
	return files, nil
}

func requireTemplateManifest(ctx context.Context, template string, expected map[string]string) error {
	owner, err := connect(ctx, localDSN, template, "tadoku_owner")
	if err != nil {
		return err
	}
	defer owner.Close(ctx)

	var present bool
	err = owner.QueryRow(ctx, "select to_regclass('queryclimb.template_migrations') is not null").Scan(&present)
	if err != nil {
		return err
	}
	if !present {
		return fmt.Errorf("template has no migration manifest: %s; rebuild it with init from the baseline", template)
	}
	rows, err := owner.Query(ctx, "select name, sha256 from queryclimb.template_migrations")
	if err != nil {
		return err
	}
	recorded, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) ([2]string, error) {
		var entry [2]string
		err := row.Scan(&entry[0], &entry[1])
		return entry, err
	})
	if err != nil {
		return err
	}

	actual := map[string]string{}
	for _, entry := range recorded {
		actual[entry[0]] = entry[1]
	}
	names := []string{}
	for name := range expected {
		names = append(names, name)
	}
	for name := range actual {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		want, inBaseline := expected[name]
		got, inTemplate := actual[name]
		if inBaseline != inTemplate || want != got {
			return fmt.Errorf("template migrations differ from baseline at %s in %s", name, template)
		}
	}
	return nil
}

func candidateOnlyMigrations(baseline, candidate map[string][]byte) ([]migrationFile, error) {
	highest := 0
	for _, name := range slices.Sorted(maps.Keys(baseline)) {
		body, ok := candidate[name]
		if !ok {
			return nil, fmt.Errorf("candidate removed migration %s", name)
		}
		if !bytes.Equal(body, baseline[name]) {
			return nil, fmt.Errorf("candidate changed existing migration %s", name)
		}
		number, err := migrationNumber(name)
		if err != nil {
			return nil, err
		}
		highest = max(highest, number)
	}

	added := []migrationFile{}
	for _, name := range slices.Sorted(maps.Keys(candidate)) {
		if _, ok := baseline[name]; ok {
			continue
		}
		number, err := migrationNumber(name)
		if err != nil {
			return nil, err
		}
		if number <= highest {
			return nil, fmt.Errorf("new migration must be numbered after %04d: %s", highest, name)
		}
		if strings.HasSuffix(name, ".down.sql") {
			continue
		}
		if !strings.HasSuffix(name, ".up.sql") || !concurrentIndex(candidate[name]) {
			return nil, fmt.Errorf("new migration must be one create index concurrently statement: %s", name)
		}
		added = append(added, migrationFile{name: name, body: candidate[name]})
	}
	return added, nil
}

func migrationNumber(name string) (int, error) {
	number, err := strconv.Atoi(strings.SplitN(name, "_", 2)[0])
	if err != nil {
		return 0, errors.New("migration has no numeric prefix: " + name)
	}
	return number, nil
}

func concurrentIndex(body []byte) bool {
	tree, err := pg_query.Parse(string(body))
	if err != nil || len(tree.Stmts) != 1 {
		return false
	}
	index := tree.Stmts[0].Stmt.GetIndexStmt()
	return index != nil && index.Concurrent
}

func recordTemplateMigrations(ctx context.Context, owner *pgx.Conn, applied map[string]string) error {
	_, err := owner.Exec(ctx, `create schema queryclimb;
	create table queryclimb.template_migrations (name text primary key, sha256 text not null)`)
	if err != nil {
		return err
	}

	names := slices.Sorted(maps.Keys(applied))
	hashes := []string{}
	for _, name := range names {
		hashes = append(hashes, applied[name])
	}
	_, err = owner.Exec(ctx, `insert into queryclimb.template_migrations (name, sha256)
	select * from unnest($1::text[], $2::text[])`, names, hashes)
	return err
}
