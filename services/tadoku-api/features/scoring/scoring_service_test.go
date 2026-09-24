package scoring

import (
	"math"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestErrRuleSetNotFoundIsNotFound(t *testing.T) {
	if got := errx.KindOf(ErrRuleSetNotFound); got != errx.NotFound {
		t.Errorf("errx.KindOf(ErrRuleSetNotFound) = %v, want %v", got, errx.NotFound)
	}
}

func TestEvaluateSelectsAndCombinesRules(t *testing.T) {
	baseFirst := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	baseLater := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	modifierOne := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	modifierTwo := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	ruleSetID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	rules := []Rule{
		{
			ID:           modifierTwo,
			Priority:     40,
			Stackable:    true,
			ActivityID:   1,
			UnitKey:      "reading_page",
			LanguageCode: "jpn",
			Tag:          "book",
			Source:       SourceAmount,
			Rate:         3,
		},
		{
			ID:         baseLater,
			Priority:   30,
			ActivityID: 1,
			Source:     SourceAmount,
			Rate:       9,
		},
		{
			ID:         baseFirst,
			Priority:   10,
			ActivityID: 1,
			Source:     SourceAmount,
			Rate:       2,
		},
		{
			ID:         modifierOne,
			Priority:   5,
			Stackable:  true,
			ActivityID: 1,
			Source:     SourceAmount,
			Rate:       .5,
		},
		{
			Priority:     50,
			Stackable:    true,
			ActivityID:   1,
			LanguageCode: "eng",
			Source:       SourceAmount,
			Rate:         100,
		},
		{
			Priority:   60,
			Stackable:  true,
			ActivityID: 2,
			Source:     SourceAmount,
			Rate:       100,
		},
		{
			Priority:   70,
			Stackable:  true,
			ActivityID: 1,
			UnitKey:    "reading_sentence",
			Source:     SourceAmount,
			Rate:       100,
		},
		{
			Priority:   80,
			Stackable:  true,
			ActivityID: 1,
			Tag:        "video",
			Source:     SourceAmount,
			Rate:       100,
		},
	}
	original := append([]Rule(nil), rules...)
	amount := float32(10)

	estimate, matched, err := evaluate(scoringInput{
		activityID:   1,
		unitKey:      "reading_page",
		languageCode: "jpn",
		tags:         []string{"book"},
		amount:       &amount,
	}, RuleSet{
		ID:    ruleSetID,
		Rules: rules,
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if !matched {
		t.Fatal("evaluate did not report a match")
	}
	if estimate.Score != 30 {
		t.Errorf("score = %v, want 30", estimate.Score)
	}
	if estimate.RuleSetID == nil || *estimate.RuleSetID != ruleSetID {
		t.Errorf("rule set ID = %v, want %v", estimate.RuleSetID, ruleSetID)
	}
	wantRules := []AppliedRule{
		{
			RuleID: baseFirst,
			Rate:   2,
		},
		{
			RuleID: modifierOne,
			Rate:   .5,
		},
		{
			RuleID: modifierTwo,
			Rate:   3,
		},
	}
	if !reflect.DeepEqual(estimate.Rules, wantRules) {
		t.Errorf("applied rules = %#v, want %#v", estimate.Rules, wantRules)
	}
	if !reflect.DeepEqual(rules, original) {
		t.Errorf("evaluate mutated source rules: got %#v, want %#v", rules, original)
	}
}

func TestEvaluateSourceValuesAndUnmatchedResult(t *testing.T) {
	amount := float32(12)
	duration := int32(90)
	tests := []struct {
		name       string
		input      scoringInput
		rule       Rule
		wantScore  float32
		wantSource Source
		matched    bool
	}{
		{
			name: "amount takes precedence over duration",
			input: scoringInput{
				activityID:      1,
				languageCode:    "jpn",
				amount:          &amount,
				durationSeconds: &duration,
			},
			rule: Rule{
				ActivityID: 1,
				Source:     SourceAmount,
				Rate:       2,
			},
			wantScore:  24,
			wantSource: SourceAmount,
			matched:    true,
		},
		{
			name: "duration converts seconds to minutes",
			input: scoringInput{
				activityID:      2,
				languageCode:    "jpn",
				durationSeconds: &duration,
			},
			rule: Rule{
				ActivityID: 2,
				Source:     SourceDurationMinutes,
				Rate:       2,
			},
			wantScore:  3,
			wantSource: SourceDurationMinutes,
			matched:    true,
		},
		{
			name: "modifier without base has zero score and nil provenance",
			input: scoringInput{
				activityID:   1,
				languageCode: "jpn",
				amount:       &amount,
			},
			rule: Rule{
				Stackable:  true,
				ActivityID: 1,
				Source:     SourceAmount,
				Rate:       2,
			},
			wantSource: SourceAmount,
		},
		{
			name: "zero rate is valid",
			input: scoringInput{
				activityID:   1,
				languageCode: "jpn",
				amount:       &amount,
			},
			rule: Rule{
				ActivityID: 1,
				Source:     SourceAmount,
				Rate:       0,
			},
			wantSource: SourceAmount,
			matched:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimate, matched, err := evaluate(tt.input, RuleSet{Rules: []Rule{tt.rule}})
			if err != nil {
				t.Fatalf("evaluate: %v", err)
			}
			if matched != tt.matched {
				t.Errorf("matched = %v, want %v", matched, tt.matched)
			}
			if estimate.Score != tt.wantScore || estimate.Source != tt.wantSource {
				t.Errorf("estimate = %#v, want score %v source %q", estimate, tt.wantScore, tt.wantSource)
			}
			if !tt.matched && (estimate.RuleSetID != nil || len(estimate.Rules) != 0) {
				t.Errorf("unmatched provenance = %#v, want nil rule set and no rules", estimate)
			}
		})
	}
}

func TestEvaluateRejectsInvalidRuleSets(t *testing.T) {
	amount := float32(1)
	valid := Rule{
		Priority:   1,
		ActivityID: 1,
		Source:     SourceAmount,
		Rate:       1,
	}
	tests := []struct {
		name  string
		rules []Rule
	}{
		{
			name:  "duplicate priority",
			rules: []Rule{valid, valid},
		},
		{
			name: "invalid activity",
			rules: []Rule{{
				Priority:   1,
				ActivityID: 0,
				Source:     SourceAmount,
				Rate:       1,
			}},
		},
		{
			name: "invalid source",
			rules: []Rule{{
				Priority:   1,
				ActivityID: 1,
				Source:     "other",
				Rate:       1,
			}},
		},
		{
			name: "negative rate",
			rules: []Rule{{
				Priority:   1,
				ActivityID: 1,
				Source:     SourceAmount,
				Rate:       -1,
			}},
		},
		{
			name: "non-finite rate",
			rules: []Rule{{
				Priority:   1,
				ActivityID: 1,
				Source:     SourceAmount,
				Rate:       float32(math.NaN()),
			}},
		},
		{
			name: "invalid nonmatching rule",
			rules: []Rule{
				{
					Priority:   1,
					ActivityID: 2,
					Source:     SourceAmount,
					Rate:       1,
				},
				{
					Priority:   2,
					ActivityID: 1,
					Source:     "other",
					Rate:       1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := evaluate(scoringInput{
				activityID:   1,
				languageCode: "jpn",
				amount:       &amount,
			}, RuleSet{Rules: tt.rules})
			if err == nil {
				t.Fatal("evaluate returned nil error")
			}
		})
	}
}

func TestEvaluateRejectsInvalidInput(t *testing.T) {
	zero := float32(0)
	negativeDuration := int32(-1)
	tests := []struct {
		name  string
		input scoringInput
	}{
		{
			name: "invalid activity",
			input: scoringInput{
				activityID:   0,
				languageCode: "jpn",
				amount:       floatPointerForTest(1),
			},
		},
		{
			name: "missing language",
			input: scoringInput{
				activityID: 1,
				amount:     floatPointerForTest(1),
			},
		},
		{
			name: "invalid amount",
			input: scoringInput{
				activityID:   1,
				languageCode: "jpn",
				amount:       &zero,
			},
		},
		{
			name: "invalid duration",
			input: scoringInput{
				activityID:      1,
				languageCode:    "jpn",
				durationSeconds: &negativeDuration,
			},
		},
		{
			name: "missing value",
			input: scoringInput{
				activityID:   1,
				languageCode: "jpn",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := evaluate(tt.input, RuleSet{})
			if err == nil {
				t.Fatal("evaluate returned nil error")
			}
		})
	}
}

func floatPointerForTest(value float32) *float32 { return &value }
