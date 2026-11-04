package scoring

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestConvertLegacyUnit(t *testing.T) {
	unitID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	pageID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	page := &logUnit{ID: pageID, Key: "reading_page", Modifier: 1}
	readingPage := "reading_page"
	pages := "pages"
	tests := []struct {
		name     string
		input    logscore.Input
		unit     logUnit
		page     *logUnit
		want     logscore.Input
		wantKind string
	}{
		{
			name:  "current key",
			input: logscore.Input{ActivityID: 1, UnitKey: &readingPage, Amount: floatPointerForTest(10)},
			unit:  logUnit{ID: unitID, Key: "reading_page"},
			want:  logscore.Input{ActivityID: 1, UnitKey: &readingPage, Amount: floatPointerForTest(10)},
		},
		{
			name:  "current unit id",
			input: logscore.Input{ActivityID: 3, UnitID: &unitID, Amount: floatPointerForTest(2)},
			unit:  logUnit{ID: unitID, Key: "writing_page"},
			want:  logscore.Input{ActivityID: 3, UnitID: &unitID, Amount: floatPointerForTest(2)},
		},
		{
			name:  "plural key before the rename",
			input: logscore.Input{ActivityID: 1, UnitKey: &pages, Amount: floatPointerForTest(10)},
			unit:  logUnit{ID: unitID, Key: "reading_page"},
			want:  logscore.Input{ActivityID: 1, UnitKey: &readingPage, Amount: floatPointerForTest(10)},
		},
		{
			name: "legacy key after the rename",
			input: logscore.Input{
				ActivityID: 3,
				UnitKey:    stringPointerForTest("writing_character"),
				Amount:     floatPointerForTest(10),
			},
			unit: logUnit{ID: unitID, Key: "characters"},
			want: logscore.Input{
				ActivityID: 3,
				UnitKey:    stringPointerForTest("characters"),
				Amount:     floatPointerForTest(10),
			},
			wantKind: "legacy_key_spelling",
		},
		{
			name: "minute amount",
			input: logscore.Input{
				ActivityID: 2,
				UnitKey:    stringPointerForTest("listening_minute"),
				Amount:     floatPointerForTest(1.51),
			},
			unit:     logUnit{ID: unitID, Key: "listening_minute"},
			want:     logscore.Input{ActivityID: 2, DurationSeconds: int32PointerForTest(91)},
			wantKind: "minute_amount",
		},
		{
			name: "dense minute amount keeps a positive duration",
			input: logscore.Input{
				ActivityID:      4,
				UnitID:          &unitID,
				Amount:          floatPointerForTest(20),
				DurationSeconds: int32PointerForTest(1000),
				Tags:            []string{"Podcast", "dense"},
			},
			unit: logUnit{ID: unitID, Key: "speaking_dense_minutes"},
			want: logscore.Input{
				ActivityID:      4,
				DurationSeconds: int32PointerForTest(1000),
				Tags:            []string{"podcast", "dense"},
			},
			wantKind: "minute_amount",
		},
		{
			name: "minute amount with a non-positive duration",
			input: logscore.Input{
				ActivityID:      5,
				UnitKey:         stringPointerForTest("study_minute"),
				Amount:          floatPointerForTest(20),
				DurationSeconds: int32PointerForTest(0),
			},
			unit: logUnit{ID: unitID, Key: "study_minute"},
			want: logscore.Input{
				ActivityID:      5,
				UnitKey:         stringPointerForTest("study_minute"),
				Amount:          floatPointerForTest(20),
				DurationSeconds: int32PointerForTest(0),
			},
		},
		{
			name: "retired unit",
			input: logscore.Input{
				ActivityID: 1,
				UnitKey:    stringPointerForTest("reading_comic_page"),
				Amount:     floatPointerForTest(10),
			},
			unit: logUnit{ID: unitID, Key: "reading_comic_page"},
			page: page,
			want: logscore.Input{
				ActivityID: 1,
				UnitID:     &pageID,
				UnitKey:    stringPointerForTest("reading_page"),
				Amount:     floatPointerForTest(10),
				Tags:       []string{"comic"},
			},
			wantKind: "retired_unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, kind, err := convertLegacyUnit(tt.input, tt.unit, tt.page)

			if err != nil {
				t.Fatalf("convertLegacyUnit: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) || kind != tt.wantKind {
				t.Errorf("convertLegacyUnit = %+v, %q, want %+v, %q", got, kind, tt.want, tt.wantKind)
			}
		})
	}
}

func TestConvertLegacyUnitRejectsInvalidResults(t *testing.T) {
	unitID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	page := &logUnit{ID: uuid.MustParse("22222222-2222-4222-8222-222222222222"), Key: "pages"}
	tests := []struct {
		name  string
		input logscore.Input
		unit  logUnit
		page  *logUnit
		kind  errx.Kind
	}{
		{
			name:  "minute amount below one second",
			input: logscore.Input{ActivityID: 2, UnitID: &unitID, Amount: floatPointerForTest(0.001)},
			unit:  logUnit{ID: unitID, Key: "listening_minute"},
			kind:  errx.InvalidInput,
		},
		{
			name:  "minute amount beyond the duration range",
			input: logscore.Input{ActivityID: 2, UnitID: &unitID, Amount: floatPointerForTest(1e8)},
			unit:  logUnit{ID: unitID, Key: "listening_minute"},
			kind:  errx.InvalidInput,
		},
		{
			name: "added tag beyond the limit",
			input: logscore.Input{
				ActivityID: 1,
				UnitID:     &unitID,
				Amount:     floatPointerForTest(10),
				Tags:       []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			},
			unit: logUnit{ID: unitID, Key: "reading_two_column_page"},
			page: page,
			kind: errx.InvalidInput,
		},
		{
			name:  "retired unit without a page unit",
			input: logscore.Input{ActivityID: 1, UnitID: &unitID, Amount: floatPointerForTest(10)},
			unit:  logUnit{ID: unitID, Key: "reading_comic_page"},
			kind:  errx.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := convertLegacyUnit(tt.input, tt.unit, tt.page)

			if errx.KindOf(err) != tt.kind {
				t.Errorf("error = %v (%v), want %v", err, errx.KindOf(err), tt.kind)
			}
		})
	}
}

func stringPointerForTest(value string) *string { return &value }

func int32PointerForTest(value int32) *int32 { return &value }
