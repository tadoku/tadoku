package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

type fidelityEntry struct {
	Query           string   `json:"query"`
	Pattern         string   `json:"pattern"`
	Case            string   `json:"case"`
	Status          string   `json:"status"`
	Reason          string   `json:"reason,omitempty"`
	ProductionRows  float64  `json:"production_rows"`
	SyntheticRows   float64  `json:"synthetic_rows"`
	Ratio           float64  `json:"row_ratio"`
	RowsPass        bool     `json:"rows_pass"`
	IndexesPass     bool     `json:"indexes_pass"`
	ExpectedIndexes []string `json:"expected_indexes"`
	ActualIndexes   []string `json:"actual_indexes"`
}

func fidelity(args []string) error {
	flags := flag.NewFlagSet("fidelity", flag.ContinueOnError)
	mapping := flags.String("map", "", "Phase 0 pattern-map TSV")
	measured := flags.String("metrics", "", "1x diagnostic metrics")
	if err := flags.Parse(args); err != nil {
		return err
	}
	var samples []metrics
	if err := readJSON(*measured, &samples); err != nil {
		return err
	}
	byQuery := map[string][]metrics{}
	for _, sample := range samples {
		byQuery[sample.Query] = append(byQuery[sample.Query], sample)
	}
	file, err := os.Open(*mapping)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	header, err := reader.Read()
	if err != nil {
		return err
	}
	entries := []fidelityEntry{}
	passed, failed, excluded := 0, 0, 0
	for {
		values, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		row := map[string]string{}
		for i, name := range header {
			row[name] = values[i]
		}
		share7, _ := strconv.ParseFloat(row["share_7d"], 64)
		share1, _ := strconv.ParseFloat(row["share_1d"], 64)
		if max(share7, share1) <= 0.05 {
			continue
		}
		name := row["query"]
		if strings.Contains(name, " (IN list collapsed") {
			name = strings.SplitN(name, " (", 2)[0]
		}
		entry := fidelityEntry{Query: name, Pattern: row["fingerprint"], Case: "representative", Status: "failed"}
		if strings.Contains(name, "previous version") || strings.Contains(name, "not sqlc") {
			entry.Status = "excluded"
			entry.Reason = "historical SQL variant or operation outside the current sqlc inventory"
			entries = append(entries, entry)
			excluded++
			continue
		}
		productionRows, err := strconv.ParseFloat(row["rows_read_per_call"], 64)
		if err != nil {
			return err
		}
		entry.ProductionRows = productionRows
		if productionRows == 0 {
			entry.Case = "production-no-due-jobs"
		}
		found := false
		actual := map[string]bool{}
		for _, sample := range byQuery[name] {
			for _, index := range sample.Indexes {
				actual[index] = true
			}
			if sample.Case == entry.Case {
				entry.SyntheticRows = sample.RowsRead
				found = true
			}
		}
		for index := range actual {
			entry.ActualIndexes = append(entry.ActualIndexes, index)
		}
		for _, token := range strings.Fields(row["indexes"]) {
			qualified := strings.SplitN(token, ":", 2)[0]
			parts := strings.Split(qualified, ".")
			entry.ExpectedIndexes = append(entry.ExpectedIndexes, parts[len(parts)-1])
		}
		slices.Sort(entry.ActualIndexes)
		slices.Sort(entry.ExpectedIndexes)
		entry.ExpectedIndexes = slices.Compact(entry.ExpectedIndexes)
		entry.IndexesPass = slices.Equal(entry.ActualIndexes, entry.ExpectedIndexes)
		if productionRows == 0 {
			entry.RowsPass = found && entry.SyntheticRows == 0
		} else {
			entry.Ratio = entry.SyntheticRows / productionRows
			entry.RowsPass = found && entry.Ratio >= 0.75 && entry.Ratio <= 1.25
		}
		if entry.RowsPass && entry.IndexesPass {
			entry.Status = "passed"
			passed++
		} else {
			failed++
			entry.Reason = fmt.Sprintf(
				"case found=%t, rows within tolerance=%t, index union equal=%t",
				found, entry.RowsPass, entry.IndexesPass,
			)
		}
		entries = append(entries, entry)
	}
	if err := output(map[string]any{"pass": failed == 0, "passed": passed, "failed": failed, "excluded": excluded,
		"counter":          "relation scan rows plus filtered rows and index rechecks; inferred Query Insights proxy",
		"index_comparison": "union of 1x case plans; usage frequencies are not compared", "entries": entries}); err != nil {
		return err
	}
	if failed > 0 {
		return errors.New("production fidelity gate failed for the listed queries")
	}
	return nil
}
