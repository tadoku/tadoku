package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//go:embed dashboard.html
var dashboardTemplate string

type dashboardQuery struct {
	Query      string   `json:"query"`
	Tier       string   `json:"tier"`
	Family     string   `json:"family"`
	Status     string   `json:"status"`
	Reason     string   `json:"reason,omitempty"`
	Baseline1  *float64 `json:"baseline_1"`
	Current1   *float64 `json:"current_1"`
	Baseline10 *float64 `json:"baseline_10"`
	Current10  *float64 `json:"current_10"`
	Ratio      *float64 `json:"ratio"`
	RowsChange *float64 `json:"rows_change"`
	Share      float64  `json:"share"`
	Kept       int      `json:"kept"`
	PR         string   `json:"pr"`
}

type dashboardStatus struct {
	Phase   string          `json:"phase"`
	Query   string          `json:"query"`
	Family  string          `json:"family"`
	State   string          `json:"state"`
	Reason  string          `json:"reason"`
	Noise   *float64        `json:"noise_threshold"`
	Reports []dashboardLink `json:"reports"`
}

type dashboardLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}
type dashboardPR struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	State string `json:"state"`
}
type dashboardEntry struct {
	Signature string `json:"signature"`
	Change    string `json:"change"`
	Evidence  string `json:"evidence"`
	AppliesIf string `json:"applies_if,omitempty"`
	FailedOn  string `json:"failed_on,omitempty"`
}
type dashboardDecision struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}
type dashboardTier struct {
	Tier                                      string
	Done, Progress, Stalled, Skipped, Pending int
}
type chartPoint struct {
	X, Y    float64
	Kept    bool
	Attempt string
	Value   float64
}
type queryBar struct {
	Query, Scale            string
	Before, After           *float64
	BeforeWidth, AfterWidth float64
}

type dashboardData struct {
	Status         dashboardStatus
	Updated        string
	Queries        []dashboardQuery
	PRs            []dashboardPR
	Playbook       []dashboardEntry
	Decisions      []dashboardDecision
	Recent         []map[string]string
	Tiers          []dashboardTier
	Bars           []queryBar
	Points         []chartPoint
	Polyline       string
	Attempts, Kept int
	Predicted      string
	Noise          string
}

func dashboard(args []string) error {
	flags := flag.NewFlagSet("dashboard", flag.ContinueOnError)
	state := flags.String("state", "", "directory containing status, triage, results, playbook, PRs and decisions")
	out := flags.String("output", "", "self-contained HTML output file")
	publish := flags.Bool("publish", false, "publish exact-base revision to Draft and record its version")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *state == "" {
		return errors.New("dashboard requires --state")
	}
	if *out == "" {
		*out = filepath.Join(*state, "dashboard.html")
	}
	data, err := loadDashboard(*state)
	if err != nil {
		return err
	}
	funcs := template.FuncMap{
		"number": func(n *float64) string {
			if n == nil {
				return "—"
			}
			return fmt.Sprintf("%.3f", *n)
		},
		"percent": func(n *float64) string {
			if n == nil {
				return "—"
			}
			return fmt.Sprintf("%+.1f%%", *n)
		},
	}
	rendered, err := template.New("dashboard").Funcs(funcs).Parse(dashboardTemplate)
	if err != nil {
		return err
	}
	var html bytes.Buffer
	if err := rendered.Execute(&html, data); err != nil {
		return err
	}
	if err := os.WriteFile(*out, html.Bytes(), 0644); err != nil {
		return err
	}
	if !*publish {
		return output(map[string]string{"html": *out})
	}

	versionPath := filepath.Join(*state, "dashboard-revision.json")
	var current struct {
		Revision struct {
			Version   int    `json:"version"`
			PinnedURL string `json:"pinned_url"`
		} `json:"revision"`
	}
	if err := readJSON(versionPath, &current); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	publishArgs := []string{"publish", *out, "--slug", "tadoku-query-climb-dashboard", "--title", "Query climb dashboard",
		"--collection", "Tadoku query climb", "--client-label", "query-climb", "--json"}
	if current.Revision.Version > 0 {
		publishArgs = append(publishArgs, "--base-version", strconv.Itoa(current.Revision.Version))
	}
	result, err := exec.Command("draft", publishArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Draft publish failed; fetch and reconcile on stale_base: %w: %s", err, result)
	}
	if err := json.Unmarshal(result, &current); err != nil {
		return fmt.Errorf("publish response must be reconciled before retry: %w", err)
	}
	if current.Revision.Version < 1 {
		return errors.New("publish did not return a revision; reconcile before retry")
	}
	if err := os.WriteFile(versionPath, result, 0644); err != nil {
		return err
	}
	if current.Revision.Version == 1 {
		result, err := exec.Command(
			"draft", "collection", "move", "tadoku-query-climb", "tadoku-query-climb-dashboard", "--first",
		).CombinedOutput()
		if err != nil {
			return fmt.Errorf("dashboard published but collection pin failed: %w: %s", err, result)
		}
	}
	return output(map[string]any{"version": current.Revision.Version, "url": current.Revision.PinnedURL})
}

func loadDashboard(dir string) (dashboardData, error) {
	data := dashboardData{
		Updated:   time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		Predicted: "0.00% (no kept SQL changes)", Noise: "Not measured",
	}
	for _, item := range []struct {
		name  string
		value any
	}{
		{"status.json", &data.Status},
		{"triage.json", &data.Queries},
		{"prs.json", &data.PRs},
		{"playbook.json", &data.Playbook},
		{"decisions.json", &data.Decisions},
	} {
		if err := readJSON(filepath.Join(dir, item.name), item.value); err != nil {
			return data, err
		}
	}
	file, err := os.Open(filepath.Join(dir, "results.tsv"))
	if err != nil {
		return data, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	header, err := reader.Read()
	if err != nil {
		return data, err
	}
	var attempts []map[string]string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return data, err
		}
		record := map[string]string{}
		for i, name := range header {
			record[name] = row[i]
		}
		attempts = append(attempts, record)
		if record["kept"] == "true" {
			data.Kept++
		}
	}
	if data.Kept > 0 {
		data.Predicted = "Not measured"
	}
	data.Attempts = len(attempts)
	data.Recent = attempts[max(0, len(attempts)-20):]
	if data.Status.Noise != nil {
		data.Noise = fmt.Sprintf("%.2f%%", *data.Status.Noise*100)
	}
	predicted, measured := 100.0, false
	for _, q := range data.Queries {
		if q.Baseline1 != nil && q.Current1 != nil && *q.Baseline1 > 0 {
			predicted += q.Share * (*q.Current1 / *q.Baseline1 - 1)
			measured = true
		}
	}
	if measured {
		data.Predicted = fmt.Sprintf("%+.2f%% (baseline = 100)", predicted-100)
	}
	for _, tier := range []string{"1", "2", "3", "skip", "unassigned"} {
		summary := dashboardTier{Tier: tier}
		for _, q := range data.Queries {
			if q.Tier != tier {
				continue
			}
			switch q.Status {
			case "done":
				summary.Done++
			case "in progress":
				summary.Progress++
			case "stalled", "out of budget":
				summary.Stalled++
			case "skipped":
				summary.Skipped++
			default:
				summary.Pending++
			}
			if tier == "1" {
				for _, bar := range []queryBar{{Query: q.Query, Scale: "1×", Before: q.Baseline1, After: q.Current1},
					{Query: q.Query, Scale: "10×", Before: q.Baseline10, After: q.Current10}} {
					data.Bars = append(data.Bars, bar)
				}
			}
		}
		data.Tiers = append(data.Tiers, summary)
	}
	maximum := 0.0
	for _, bar := range data.Bars {
		if bar.Before != nil {
			maximum = math.Max(maximum, *bar.Before)
		}
		if bar.After != nil {
			maximum = math.Max(maximum, *bar.After)
		}
	}
	if maximum > 0 {
		for i := range data.Bars {
			if v := data.Bars[i].Before; v != nil {
				data.Bars[i].BeforeWidth = *v / maximum * 520
			}
			if v := data.Bars[i].After; v != nil {
				data.Bars[i].AfterWidth = *v / maximum * 520
			}
		}
	}
	for _, attempt := range attempts {
		value, err := strconv.ParseFloat(attempt["predicted_time"], 64)
		if err != nil {
			continue
		}
		data.Points = append(data.Points, chartPoint{
			Value: value, Kept: attempt["kept"] == "true", Attempt: attempt["attempt"],
		})
	}
	maxY := 100.0
	for _, point := range data.Points {
		maxY = math.Max(maxY, point.Value)
	}
	coordinates := []string{}
	for i := range data.Points {
		point := &data.Points[i]
		point.X = 45 + float64(i)/float64(max(1, len(data.Points)-1))*650
		point.Y = 210 - point.Value/maxY*180
		coordinates = append(coordinates, fmt.Sprintf("%.2f,%.2f", point.X, point.Y))
	}
	data.Polyline = strings.Join(coordinates, " ")
	return data, nil
}
