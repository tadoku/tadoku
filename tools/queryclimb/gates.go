package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
)

func gate(args []string) error {
	flags := flag.NewFlagSet("gate", flag.ContinueOnError)
	mode := flags.String("mode", "determinism", "determinism, guard or candidate")
	target := flags.String("target", "", "target query for candidate gate")
	outputPath := flags.String("output", "", "JSON result file; default stdout")
	baselineFiles := flags.String("baseline", "", "comma-separated baseline metrics files")
	candidateFiles := flags.String("candidate", "", "comma-separated candidate metrics files")
	latencyFiles := flags.String("latency", "", "comma-separated paired latency files")
	pageFiles := flags.String("page-oracles", "", "baseline legal-page oracles, one file per scale")
	proofFiles := flags.String("buffer-proof", "", "checked unchanged-input proofs for documented buffer variance")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *mode != "determinism" && *mode != "candidate" && *mode != "guard" {
		return errors.New("invalid gate mode")
	}
	a, err := metricFiles(*baselineFiles)
	if err != nil {
		return err
	}
	b, err := metricFiles(*candidateFiles)
	if err != nil {
		return err
	}
	qualification, err := readQualifications(*pageFiles, *proofFiles, *target)
	if err != nil {
		return err
	}
	replacements := []string{}
	failures := []string{}
	if len(a) != len(b) {
		failures = append(failures, "coverage differs")
	}
	for key, original := range a {
		candidate, ok := b[key]
		if !ok {
			failures = append(failures, key+": missing")
			continue
		}
		if _, hasOracle := qualification.pages[key]; hasOracle {
			if err := qualification.page(key, original, candidate); err != nil {
				failures = append(failures, key+": legal-page check: "+err.Error())
			} else {
				replacements = append(replacements, key+": exact legal boundary-page multiset")
			}
		} else if candidate.Fingerprint != original.Fingerprint {
			failures = append(failures, key+": fingerprint mismatch")
		}
		buffersDiffer := candidate.SharedBuffers != original.SharedBuffers || candidate.TempBuffers != original.TempBuffers
		if *mode == "determinism" {
			if candidate.RowsProcessed != original.RowsProcessed || candidate.PlanHash != original.PlanHash ||
				candidate.CustomPlans != original.CustomPlans || candidate.GenericPlans != original.GenericPlans {
				failures = append(failures, key+": deterministic metrics differ")
			}
			if buffersDiffer {
				if err := qualification.buffers(key, original, candidate); err != nil {
					failures = append(failures, key+": buffers differ: "+err.Error())
				} else {
					replacements = append(replacements, key+": unchanged-input and full estimated-plan identity")
				}
			}
		} else {
			if candidate.RowsProcessed > original.RowsProcessed {
				failures = append(failures, key+": rows processed increased")
			}
			buffersIncreased := candidate.SharedBuffers > original.SharedBuffers || candidate.TempBuffers > original.TempBuffers
			if original.Query != *target && buffersIncreased {
				if err := qualification.buffers(key, original, candidate); err != nil {
					failures = append(failures, key+": global buffers increased: "+err.Error())
				} else {
					replacements = append(replacements, key+": unchanged-input and full estimated-plan identity")
				}
			}
		}
	}
	ratio := 1.0
	if *mode == "candidate" {
		if *target == "" || *latencyFiles == "" {
			return errors.New("candidate gate needs target and paired latency files")
		}
		logRatio, count, minimumThreshold := 0.0, 0, 0.02
		seen := map[string]bool{}
		for scale, path := range strings.Split(*latencyFiles, ",") {
			var samples []latency
			if err := readJSON(path, &samples); err != nil {
				return err
			}
			for _, sample := range samples {
				if sample.Query != *target {
					return errors.New("latency includes wrong target")
				}
				key := fmt.Sprintf("%d/%s/%s", scale, sample.Query, sample.Case)
				if seen[key] {
					return fmt.Errorf("duplicate latency %s", key)
				}
				seen[key] = true
				original, exists := a[key]
				if !exists || original.Query != sample.Query {
					return fmt.Errorf("latency has no baseline: %s", key)
				}
				if len(sample.Baseline) < 15 || len(sample.Candidate) < 15 {
					return errors.New("fewer than 15 batches")
				}
				noise, err := measuredNoise(sample)
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				if noise > 0.1 {
					failures = append(failures, key+": noise exceeds ceiling")
				}
				threshold := math.Max(0.02, 2*noise)
				minimumThreshold = math.Max(minimumThreshold, threshold)
				r := median(sample.Candidate) / median(sample.Baseline)
				if r > 1+threshold {
					failures = append(failures, key+": case latency regressed")
				}
				logRatio += math.Log(r)
				count++
			}
		}
		for key, original := range a {
			if original.Query == *target && !seen[key] {
				failures = append(failures, key+": missing latency")
			}
		}
		if count == 0 {
			return errors.New("no target latency cases")
		}
		ratio = math.Exp(logRatio / float64(count))
		if ratio >= 1-minimumThreshold {
			failures = append(failures, "geometric mean improvement does not exceed noise threshold")
		}
	}
	result := map[string]any{"mode": *mode, "pass": len(failures) == 0, "cases": len(a),
		"latency_ratio": ratio, "failures": failures, "assertion_replacements": replacements}
	if *outputPath == "" {
		if err := output(result); err != nil {
			return err
		}
	} else if err := writeJSON(*outputPath, result); err != nil {
		return err
	}

	if len(failures) > 0 {
		return errors.New("gate failed")
	}
	return nil
}

func readJSON(path string, value any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, value)
}

func metricFiles(paths string) (map[string]metrics, error) {
	all := map[string]metrics{}
	for scale, path := range strings.Split(paths, ",") {
		var samples []metrics
		if err := readJSON(path, &samples); err != nil {
			return nil, err
		}
		for _, sample := range samples {
			key := fmt.Sprintf("%d/%s/%s", scale, sample.Query, sample.Case)
			if _, exists := all[key]; exists {
				return nil, fmt.Errorf("duplicate metric %s", key)
			}
			all[key] = sample
		}
	}
	if len(all) == 0 {
		return nil, errors.New("empty metrics")
	}
	return all, nil
}

func measuredNoise(sample latency) (float64, error) {
	if len(sample.Probe[0]) != 5 || len(sample.Probe[1]) != 5 ||
		len(sample.BatchNoise) != len(sample.Baseline)+len(sample.Candidate) {
		return 0, errors.New("missing raw noise samples")
	}
	for _, series := range [][]float64{sample.Probe[0], sample.Probe[1], sample.Baseline, sample.Candidate} {
		for _, value := range series {
			if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
				return 0, errors.New("invalid latency sample")
			}
		}
	}
	noise := pairedNoise(sample.Probe[0], sample.Probe[1])
	for _, value := range sample.BatchNoise {
		if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, errors.New("invalid pre-batch noise")
		}
		noise = math.Max(noise, value)
	}
	if sample.AA {
		if len(sample.Baseline) != len(sample.Candidate) {
			return 0, errors.New("unpaired A/A sample")
		}
		noise = math.Max(noise, pairedNoise(sample.Baseline, sample.Candidate))
	}
	return noise, nil
}
