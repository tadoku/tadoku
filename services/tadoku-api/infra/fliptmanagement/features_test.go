package fliptmanagement_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
)

func TestParseCurrentFeatures(t *testing.T) {
	path, err := runfiles.Rlocation("tadoku/k8s/dev/base/flipt/features.yaml")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	resources, err := fliptmanagement.ParseFeatures(file)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(resources)
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("testdata/features.json")
	if err != nil {
		t.Fatal(err)
	}
	var got, want any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(golden, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("converted resources differ:\n%s\nwant:\n%s", encoded, golden)
	}
}

func TestParseFeaturesRejectsInvalidInput(t *testing.T) {
	valid := "version: \"1.6\"\nnamespace:\n  key: default\n  name: Default\n"
	cases := map[string]string{
		"unknown field": valid + "unknown: true\n",
		"nested unknown field": valid + `flags:
- key: example
  type: BOOLEAN_FLAG_TYPE
  unknown: true
`,
		"duplicate document":  valid + "---\n" + valid,
		"duplicate key":       valid + "segments:\n- key: one\n- key: one\n",
		"missing version":     "namespace: {key: default}\n",
		"unsupported version": strings.Replace(valid, "1.6", "1.5", 1),
		"two rollout types": valid + `flags:
- key: flag
  type: BOOLEAN_FLAG_TYPE
  rollouts:
  - threshold: {percentage: 10, value: true}
    segment: {keys: [one], operator: OR_SEGMENT_OPERATOR, value: true}
`,
		"no rollout type": valid + `flags:
- key: flag
  type: BOOLEAN_FLAG_TYPE
  rollouts:
  - description: missing
`,
		"invalid threshold": valid + `flags:
- key: flag
  type: BOOLEAN_FLAG_TYPE
  rollouts:
  - threshold: {percentage: 101, value: true}
`,
		"invalid operator": valid + `flags:
- key: flag
  type: BOOLEAN_FLAG_TYPE
  rollouts:
  - segment: {keys: [one], operator: UNKNOWN, value: true}
`,
		"mapping constraint value": valid + `segments:
- key: one
  constraints:
  - {type: STRING_COMPARISON_TYPE, property: user, operator: eq, value: {bad: value}}
`,
		"unsupported flag type": valid + "flags:\n- key: flag\n  type: NOT_A_FLAG\n",
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := fliptmanagement.ParseFeatures(strings.NewReader(input)); err == nil {
				t.Fatal("invalid features accepted")
			}
		})
	}
}

func TestParseFeaturesMapsThresholdAndConstraintValues(t *testing.T) {
	input := `version: "1.6"
namespace: {key: default}
flags:
- key: threshold
  type: BOOLEAN_FLAG_TYPE
  enabled: true
  rollouts:
  - threshold: {percentage: 25.5, value: false}
segments:
- key: audience
  match_type: ANY_MATCH_TYPE
  constraints:
  - {type: STRING_COMPARISON_TYPE, property: country, operator: isoneof, value: [BE, NL]}
  - {type: NUMBER_COMPARISON_TYPE, property: age, operator: gt, value: "18"}
`
	resources, err := fliptmanagement.ParseFeatures(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(resources)
	if err != nil {
		t.Fatal(err)
	}
	var got []struct {
		Payload struct {
			Constraints []struct{ Value string }
			Rollouts    []struct {
				Type      string
				Threshold struct {
					Percentage float64
					Value      bool
				}
			}
		}
	}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Payload.Constraints[0].Value != `["BE","NL"]` ||
		got[0].Payload.Constraints[1].Value != "18" {
		t.Fatalf("constraint conversion: %s", encoded)
	}
	rollout := got[1].Payload.Rollouts[0]
	if rollout.Type != "THRESHOLD_ROLLOUT_TYPE" || rollout.Threshold.Percentage != 25.5 || rollout.Threshold.Value {
		t.Fatalf("threshold conversion: %s", encoded)
	}
}
