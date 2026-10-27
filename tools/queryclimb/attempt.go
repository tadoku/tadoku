package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

func attempt(ctx context.Context, root string, args []string) error {
	flags := flag.NewFlagSet("attempt", flag.ContinueOnError)
	candidate := flags.String("candidate-root", "", "candidate checkout with regenerated sqlc")
	templates := flags.String("templates", "", "comma-separated frozen 1x and 10x template names")
	caseFiles := flags.String("cases", "", "comma-separated frozen 1x and 10x case files")
	target := flags.String("target", "", "one query to judge")
	dir := flags.String("output", "", "new attempt evidence directory")
	iterations := flags.Int("iterations", 0, "executions per latency batch; zero chooses automatically")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *candidate == "" || *target == "" || *dir == "" {
		return errors.New("attempt requires candidate-root, target and output")
	}
	templateNames, cases := strings.Split(*templates, ","), strings.Split(*caseFiles, ",")
	if len(templateNames) != 2 || len(cases) != 2 {
		return errors.New("attempt requires 1x and 10x templates and cases")
	}
	if err := os.Mkdir(*dir, 0755); err != nil {
		return err
	}
	if err := sqlcCheck(ctx, *candidate); err != nil {
		return err
	}

	for side, checkout := range map[string]string{"baseline": root, "candidate": *candidate} {
		command := exec.CommandContext(ctx, "git", "diff", "HEAD", "--",
			"services/tadoku-api/sql", "services/tadoku-api/generated/sqlc", "services/tadoku-api/migrations")
		command.Dir = checkout
		diff, err := command.Output()
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(*dir, side+".diff"), diff, 0644); err != nil {
			return err
		}
	}
	caseHashes := []string{}
	for _, path := range cases {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		caseHashes = append(caseHashes, digest(string(data)))
	}
	beforeSHA, err := revision(root)
	if err != nil {
		return err
	}
	afterSHA, err := revision(*candidate)
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*dir, "manifest.json"), map[string]any{
		"baseline_revision": beforeSHA, "candidate_revision": afterSHA,
		"case_sha256": caseHashes, "harness_sha256": digest(string(binary)),
		"templates": templateNames, "cases": cases, "target": *target,
	}); err != nil {
		return err
	}
	a, err := generatedInterface(root)
	if err != nil {
		return err
	}
	b, err := generatedInterface(*candidate)
	if err != nil {
		return err
	}
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	if string(aJSON) != string(bJSON) {
		return errors.New("candidate generated Go interface differs outside SQL strings")
	}
	queries, err := inventory(root)
	if err != nil {
		return err
	}
	candidates, err := inventory(*candidate)
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*dir, "candidate-queries.json"), candidates); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(*dir, "baseline-queries.json"), queries); err != nil {
		return err
	}
	if _, err := findQuery(queries, *target); err != nil {
		return err
	}
	if len(queries) != len(candidates) {
		return errors.New("candidate query inventory differs")
	}

	for _, original := range queries {
		changed, err := findQuery(candidates, original.Key)
		if err != nil {
			return err
		}
		if original.Key != *target && original.SQL != changed.SQL {
			return fmt.Errorf("one attempt may change only the target SQL; %s also changed", original.Key)
		}
	}
	for side, checkout := range map[string]string{"baseline": root, "candidate": *candidate} {
		paths, err := filepath.Glob(filepath.Join(checkout, "services/tadoku-api/migrations/*.up.sql"))
		if err != nil {
			return err
		}
		migrations := map[string]string{}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			migrations[filepath.Base(path)] = string(data)
		}
		if err := writeJSON(filepath.Join(*dir, side+"-migrations.json"), migrations); err != nil {
			return err
		}
	}
	baselinePaths, candidatePaths, latencyPaths := []string{}, []string{}, []string{}
	for scale, templateName := range templateNames {
		baselineDB := "qc_attempt_" + digest(*dir)[:12] + "_a" + strconv.Itoa(scale)
		candidateDB := "qc_attempt_" + digest(*dir)[:12] + "_b" + strconv.Itoa(scale)
		baselinePath := filepath.Join(*dir, fmt.Sprintf("baseline-%d.json", scale))
		candidatePath := filepath.Join(*dir, fmt.Sprintf("candidate-%d.json", scale))
		latencyPath := filepath.Join(*dir, fmt.Sprintf("latency-%d.json", scale))
		baselinePaths = append(baselinePaths, baselinePath)
		candidatePaths = append(candidatePaths, candidatePath)
		latencyPaths = append(latencyPaths, latencyPath)
		if err := clone(ctx, templateName, baselineDB); err != nil {
			return err
		}
		if err := clone(ctx, templateName, candidateDB); err != nil {
			return err
		}
		defer func() {
			if err := dropAttemptClone(context.WithoutCancel(ctx), baselineDB); err != nil {
				fmt.Fprintln(os.Stderr, "retained baseline clone:", err)
			}
			if err := dropAttemptClone(context.WithoutCancel(ctx), candidateDB); err != nil {
				fmt.Fprintln(os.Stderr, "retained candidate clone:", err)
			}
		}()
		if err := applyNewIndexes(ctx, root, baselineDB); err != nil {
			return err
		}
		if err := applyNewIndexes(ctx, *candidate, candidateDB); err != nil {
			return err
		}
		specs, err := readCases(cases[scale])
		if err != nil {
			return err
		}
		for i := range specs {
			before, err := findQuery(queries, specs[i].Query)
			if err != nil {
				return err
			}
			after, err := findQuery(candidates, specs[i].Query)
			if err != nil {
				return err
			}
			tables := append(slices.Clone(before.Mutations), after.Mutations...)
			slices.Sort(tables)
			specs[i].StateTables = slices.Compact(tables)
		}
		for side, db := range []string{baselineDB, candidateDB} {
			source, path := queries, baselinePath
			if side == 1 {
				source, path = candidates, candidatePath
			}
			samples, err := inspectCases(ctx, db, source, specs)
			if err != nil {
				return err
			}
			if err := writeJSON(path, samples); err != nil {
				return err
			}
		}
		if err := gate([]string{
			"--mode", "guard", "--target", *target,
			"--baseline", strings.Join(baselinePaths, ","),
			"--candidate", strings.Join(candidatePaths, ","),
			"--output", filepath.Join(*dir, fmt.Sprintf("guard-%d.json", scale)),
		}); err != nil {
			return err
		}
		targets := []caseSpec{}
		for _, c := range specs {
			if c.Query == *target {
				targets = append(targets, c)
			}
		}
		if len(targets) == 0 {
			return errors.New("target has no frozen cases")
		}
		targetPath := filepath.Join(*dir, fmt.Sprintf("target-cases-%d.json", scale))
		if err := writeJSON(targetPath, targets); err != nil {
			return err
		}
		samples, err := collectBaseline(
			ctx, root, *candidate, baselineDB, candidateDB, targetPath, *iterations, 15, latencyPath,
		)
		if err != nil {
			return err
		}
		if err := writeJSON(latencyPath, samples); err != nil {
			return err
		}
	}
	return gate([]string{"--mode", "candidate", "--target", *target,
		"--baseline", strings.Join(baselinePaths, ","), "--candidate", strings.Join(candidatePaths, ","),
		"--latency", strings.Join(latencyPaths, ",")})
}

func inspectCases(ctx context.Context, database string, queries []query, cases []caseSpec) ([]metrics, error) {
	conn, err := connect(ctx, localDSN, database, "tadoku")
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)
	samples := make([]metrics, 0, len(cases))
	for _, c := range cases {
		q, err := findQuery(queries, c.Query)
		if err != nil {
			return nil, err
		}
		sample, err := measure(ctx, conn, q, c)
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", q.Key, c.Name, err)
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func applyNewIndexes(ctx context.Context, root, database string) error {
	paths, err := filepath.Glob(filepath.Join(root, "services/tadoku-api/migrations/*.up.sql"))
	if err != nil {
		return err
	}
	slices.Sort(paths)
	owner, err := connect(ctx, localDSN, database, "tadoku_owner")
	if err != nil {
		return err
	}
	defer owner.Close(ctx)
	for _, path := range paths {
		number, err := strconv.Atoi(strings.SplitN(filepath.Base(path), "_", 2)[0])
		if err != nil {
			return err
		}
		if number <= 58 {
			continue
		}
		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		tree, err := pg_query.Parse(string(text))
		if err != nil {
			return err
		}
		if len(tree.Stmts) != 1 {
			return errors.New("new index migration must contain exactly one statement")
		}
		index := tree.Stmts[0].Stmt.GetIndexStmt()
		if index == nil || !index.Concurrent {
			return errors.New("only standalone concurrent index additions are allowed")
		}
		if _, err := owner.Exec(ctx, string(text)); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}

func recordAttempt(args []string) error {
	flags := flag.NewFlagSet("record", flag.ContinueOnError)
	state := flags.String("state", "", "dashboard state directory")
	input := flags.String("input", "", "attempt JSON; kept result must name a passing gate file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	var row map[string]string
	if err := readJSON(*input, &row); err != nil {
		return err
	}
	if row["kept"] == "true" {
		var judged struct {
			Pass  bool    `json:"pass"`
			Mode  string  `json:"mode"`
			Ratio float64 `json:"latency_ratio"`
		}
		if err := readJSON(row["gate_file"], &judged); err != nil {
			return err
		}
		if !judged.Pass || judged.Mode != "candidate" {
			return errors.New("cannot record a keep without a passing harness gate")
		}
		row["latency_ratio"] = strconv.FormatFloat(judged.Ratio, 'g', -1, 64)
	}
	path := filepath.Join(*state, "results.tsv")
	source, err := os.Open(path)
	if err != nil {
		return err
	}
	reader := csv.NewReader(source)
	reader.Comma = '\t'
	header, err := reader.Read()
	source.Close()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	values := make([]string, len(header))
	for i, name := range header {
		values[i] = row[name]
	}
	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	if err := writer.Write(values); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

func revision(root string) (string, error) {
	command := exec.Command("git", "rev-parse", "HEAD")
	command.Dir = root
	b, err := command.Output()
	return strings.TrimSpace(string(b)), err
}

func dropAttemptClone(ctx context.Context, name string) error {
	if !strings.HasPrefix(name, "qc_attempt_") {
		return errors.New("only disposable attempt clones may be removed")
	}
	admin, err := connect(ctx, localDSN, "postgres", "postgres")
	if err != nil {
		return err
	}
	defer admin.Close(ctx)
	_, err = admin.Exec(ctx, "drop database "+pgx.Identifier{name}.Sanitize())
	return err
}
