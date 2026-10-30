package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type latency struct {
	Query           string       `json:"query"`
	Case            string       `json:"case"`
	Baseline        []float64    `json:"baseline_ms"`
	Candidate       []float64    `json:"candidate_ms"`
	MedianBaseline  float64      `json:"median_baseline_ms"`
	MedianCandidate float64      `json:"median_candidate_ms"`
	Noise           float64      `json:"noise"`
	Threshold       float64      `json:"threshold"`
	Executions      int          `json:"executions_per_batch"`
	Probe           [2][]float64 `json:"probe_ms"`
	BatchNoise      []float64    `json:"batch_noise"`
	AA              bool         `json:"aa_control"`
}

func machineQuiet() error {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return err
	}
	fields := strings.Fields(string(b))
	load, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return err
	}
	ceiling := math.Max(2, float64(runtime.NumCPU())/2)
	if load > ceiling {
		return fmt.Errorf("machine load %.2f exceeds timing ceiling %.2f; retry later", load, ceiling)
	}

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}
		cmd, err := os.ReadFile("/proc/" + entry.Name() + "/cmdline")
		if err != nil || len(cmd) == 0 {
			continue
		}
		text := strings.ReplaceAll(string(cmd), "\x00", " ")
		for _, heavy := range []string{"BazelServer", "sqlc generate", "pnpm build", "compilepkg", "link -o"} {
			if strings.Contains(text, heavy) {
				return fmt.Errorf("timing blocked by process %s (%s)", entry.Name(), heavy)
			}
		}
	}
	return nil
}

func baseline(ctx context.Context, root string, args []string) error {
	flags := flag.NewFlagSet("baseline", flag.ContinueOnError)
	database := flags.String("name", "qc_schema", "baseline database")
	candidateDB := flags.String("candidate-name", "", "candidate database; default is A/A")
	candidateRoot := flags.String("candidate-root", "", "candidate SQL checkout; default is A/A")
	casesPath := flags.String("cases", "", "case definitions")
	iterations := flags.Int("iterations", 0, "executions per batch; zero chooses from baseline plan cost")
	batches := flags.Int("batches", 15, "alternating batches per side; minimum 15")
	outputPath := flags.String("output", "", "retain each completed case even if a later case fails")
	pageOracles := flags.String("page-oracles", "", "baseline legal-page oracle file for this scale")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *batches < 15 || *iterations < 0 {
		return errors.New("need at least 15 batches; iterations must be nonnegative")
	}
	if *candidateRoot == "" {
		*candidateRoot = root
	}
	if *candidateDB == "" {
		*candidateDB = *database
	}
	results, err := collectBaseline(
		ctx, root, *candidateRoot, *database, *candidateDB, *casesPath, *iterations, *batches, *outputPath, *pageOracles,
	)
	if err != nil {
		return err
	}
	return output(results)
}

func collectBaseline(
	ctx context.Context,
	root, candidateRoot, database, candidateDB, casesPath string,
	iterations, batches int,
	outputPath, pageOracles string,
) ([]latency, error) {
	queries, err := inventory(root)
	if err != nil {
		return nil, err
	}
	candidates, err := inventory(candidateRoot)
	if err != nil {
		return nil, err
	}
	cases, err := readCases(casesPath)
	if err != nil {
		return nil, err
	}
	qualification, err := readQualifications(pageOracles, "", "", false)
	if err != nil {
		return nil, err
	}
	admin, err := connect(ctx, localDSN, "postgres", "postgres")
	if err != nil {
		return nil, err
	}
	defer admin.Close(ctx)

	sameInput := root == candidateRoot && database == candidateDB
	suffix := digest(fmt.Sprintf("%s/%s/%d", database, casesPath, time.Now().UnixNano()))[:16]
	probeDB := "qc_attempt_probe_" + suffix
	cleanup := []string{}
	defer func() {
		for _, name := range cleanup {
			if err := dropAttemptClone(context.WithoutCancel(ctx), name); err != nil {
				fmt.Fprintln(os.Stderr, "retained measurement clone:", err)
			}
		}
	}()
	if err := clone(ctx, database, probeDB); err != nil {
		return nil, err
	}
	cleanup = append(cleanup, probeDB)
	if database == candidateDB {
		original := database
		database, candidateDB = "qc_attempt_aa_"+suffix, "qc_attempt_ab_"+suffix
		for _, name := range []string{database, candidateDB} {
			if err := clone(ctx, original, name); err != nil {
				return nil, err
			}
			cleanup = append(cleanup, name)
		}
		for key, oracle := range qualification.pages {
			if oracle.Database == original {
				oracle.Database = database
				qualification.pages[key] = oracle
			}
		}
		fmt.Fprintf(os.Stderr, "measurement clones: source=%s baseline=%s candidate=%s probe=%s\n",
			original, database, candidateDB, probeDB)
	}
	connections := []*pgx.Conn{}
	for _, db := range []string{database, candidateDB, probeDB} {
		conn, err := connect(ctx, localDSN, db, "tadoku")
		if err != nil {
			return nil, err
		}
		defer conn.Close(ctx)
		connections = append(connections, conn)
	}
	results := []latency{}
	for _, c := range cases {
		a, err := findQuery(queries, c.Query)
		if err != nil {
			return nil, err
		}
		b, err := findQuery(candidates, c.Query)
		if err != nil {
			return nil, err
		}
		caseIterations := iterations
		queryIDs := make([]int64, 3)
		var baselineSample metrics
		for i, q := range []query{a, b, a} {
			sample, err := measure(ctx, connections[i], q, c)
			if err != nil {
				return nil, err
			}
			queryIDs[i] = sample.QueryID
			if i == 0 {
				baselineSample = sample
			}
			if i == 1 {
				key := "0/" + c.Query + "/" + c.Name
				if _, hasOracle := qualification.pages[key]; hasOracle {
					if err := qualification.page(key, baselineSample, sample); err != nil {
						return nil, fmt.Errorf("legal page before timing: %w", err)
					}
				} else if sample.Fingerprint != baselineSample.Fingerprint {
					return nil, errors.New("baseline/candidate fingerprints differ before timing")
				}
			}
			if i == 0 && caseIterations == 0 {
				var plan []map[string]any
				if err := json.Unmarshal(sample.Plan, &plan); err != nil {
					return nil, err
				}
				elapsed, _ := plan[0]["Execution Time"].(float64)
				caseIterations = max(1, min(1000, int(math.Ceil(20/math.Max(0.001, elapsed)))))
				if elapsed < 1 {
					caseIterations = 1000
				}
			}
		}
		fmt.Fprintf(os.Stderr, "case %s/%s: %d executions per batch\n", c.Query, c.Name, caseIterations)
		for warm := 0; warm < 2; warm++ {
			if _, err := timedBatch(ctx, admin, connections[2], c, caseIterations, queryIDs[2]); err != nil {
				return nil, err
			}
		}
		probe := [2][]float64{}
		for n := 0; n < 5; n++ {
			for side := 0; side < 2; side++ {
				ms, err := timedBatch(ctx, admin, connections[2], c, caseIterations, queryIDs[2])
				if err != nil {
					return nil, err
				}
				probe[side] = append(probe[side], ms)
			}
		}
		noise := pairedNoise(probe[0], probe[1])
		if noise > 0.10 {
			raw, _ := json.Marshal(probe)
			fmt.Fprintf(os.Stderr, "noise probe samples (ms): %s\n", raw)
			return nil, fmt.Errorf("%s/%s: noise probe %.2f%% exceeds 10%%; retry later", c.Query, c.Name, noise*100)
		}
		result := latency{
			Query: c.Query, Case: c.Name, Noise: noise, Threshold: math.Max(0.02, 2*noise),
			Executions: caseIterations, Probe: probe, AA: sameInput,
		}
		for n := 0; n < batches; n++ {
			for side := 0; side < 2; side++ {
				ms, batchNoise, err := probedBatch(
					ctx, admin, connections[side], connections[2], c, caseIterations, queryIDs[side], queryIDs[2],
				)
				result.BatchNoise = append(result.BatchNoise, batchNoise)
				result.Noise = math.Max(result.Noise, batchNoise)
				result.Threshold = math.Max(0.02, 2*result.Noise)
				if err != nil {
					return nil, err
				}
				if side == 0 {
					result.Baseline = append(result.Baseline, ms)
				} else {
					result.Candidate = append(result.Candidate, ms)
				}
			}
		}
		result.MedianBaseline, result.MedianCandidate = median(result.Baseline), median(result.Candidate)
		if sameInput {
			result.Noise = math.Max(result.Noise, pairedNoise(result.Baseline, result.Candidate))
			result.Threshold = math.Max(0.02, 2*result.Noise)
		}
		if result.AA && result.Noise > 0.10 {
			raw, _ := json.Marshal(result)
			fmt.Fprintf(os.Stderr, "A/A result: %s\n", raw)
			return nil, fmt.Errorf("%s/%s: A/A noise exceeds 10%%", c.Query, c.Name)
		}
		results = append(results, result)
		if outputPath != "" {
			if err := writeJSON(outputPath, results); err != nil {
				return nil, err
			}
		}
	}
	return results, nil
}

func timedBatch(
	ctx context.Context,
	admin, conn *pgx.Conn,
	c caseSpec,
	iterations int,
	queryID int64,
) (float64, error) {
	if err := machineQuiet(); err != nil {
		return 0, err
	}
	if _, err := admin.Exec(ctx, "select pg_stat_statements_reset()"); err != nil {
		return 0, err
	}
	for n := 0; n < iterations; n++ {
		err := scoped(ctx, conn, c.Tenant, func() error { _, _, err := execute(ctx, conn, c); return err })
		if err != nil {
			return 0, err
		}
	}
	var elapsed float64
	var calls int64
	err := admin.QueryRow(ctx, `select coalesce(sum(total_exec_time), 0), coalesce(sum(calls), 0)::bigint
      from pg_stat_statements
      where userid = (select oid from pg_roles where rolname = 'tadoku')
        and dbid = (select oid from pg_database where datname = $1)
        and queryid = $2 and toplevel`, conn.Config().Database, queryID).Scan(&elapsed, &calls)
	if err != nil {
		return 0, err
	}
	if calls != int64(iterations) {
		return 0, fmt.Errorf("pg_stat_statements calls: got %d, want %d", calls, iterations)
	}
	return elapsed / float64(calls), nil
}

func pairedNoise(a, b []float64) float64 {
	differences := make([]float64, len(a))
	for i := range a {
		differences[i] = math.Abs(b[i]/a[i] - 1)
	}
	slices.Sort(differences)
	return differences[int(math.Ceil(float64(len(a))*0.9))-1]
}

func median(values []float64) float64 {
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[middle-1] + sorted[middle]) / 2
	}
	return sorted[middle]
}

func sqlcCheck(ctx context.Context, root string) error {
	before, err := generatedInterface(root)
	if err != nil {
		return err
	}
	if err := shell(ctx, root, "./scripts/generate-sqlc.sh"); err != nil {
		return err
	}
	after, err := generatedInterface(root)
	if err != nil {
		return err
	}
	a, _ := json.Marshal(before)
	b, _ := json.Marshal(after)
	if string(a) != string(b) {
		return errors.New("sqlc changed generated Go outside named query SQL literals")
	}
	cmd := exec.CommandContext(ctx, "git", "diff", "--stat", "--", "services/tadoku-api/generated/sqlc")
	cmd.Dir = root
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	return cmd.Run()
}

func probedBatch(
	ctx context.Context,
	admin, target, probe *pgx.Conn,
	c caseSpec,
	iterations int,
	targetID, probeID int64,
) (float64, float64, error) {
	if target.Config().Database == probe.Config().Database {
		return 0, 0, errors.New("noise probes must use an independent database")
	}
	a, err := timedBatch(ctx, admin, probe, c, iterations, probeID)
	if err != nil {
		return 0, 0, err
	}
	b, err := timedBatch(ctx, admin, probe, c, iterations, probeID)
	if err != nil {
		return 0, 0, err
	}
	noise := math.Abs(b/a - 1)
	if noise > 0.1 {
		return 0, noise, fmt.Errorf("pre-batch noise %.2f%% exceeds 10%%; retry later", noise*100)
	}
	ms, err := timedBatch(ctx, admin, target, c, iterations, targetID)
	return ms, noise, err
}
