//nolint:testpackage // test unexported helpers in cmd.
package cmd

import (
	"errors"
	"net/url"
	"strconv"
	"testing"
	"time"
)

const (
	measureTypeWeight       = "weight"
	measureTypeBPSys        = "bp_sys"
	measureTypeWeightID     = "1"
	measureTypeBPSysID      = "10"
	measureTypeDedup        = "bodyweight"
	measureCategoryRealText = "real"
	measureCategoryGoalText = "goal"
	measureCategoryRealID   = "1"
	measureCategoryGoalID   = "2"
	testParseCategoryErrFmt = "parseMeasureCategory: %v"
	testCategoryGotFmt      = "category got %q want %q"
	testParseTypesErrFmt    = "parseMeasureTypes: %v"
	testTypesGotFmt         = "types got %q want %q"
	testParseEpochErrFmt    = "parseEpoch: %v"
	testEpochGotFmt         = "epoch got %d want %d"
	testBuildParamsErrFmt   = "buildMeasureParams: %v"
	testParamGotFmt         = "param %s got %v want %v"
	testLastUpdateValue     = 123
	testLimitValue          = 100
	testOffsetValue         = 10
	testFirstIndex          = 0
)

// TestParseMeasureCategory accepts text and numeric values.
func TestParseMeasureCategory(t *testing.T) {
	t.Parallel()

	category, err := parseMeasureCategory(measureCategoryRealText)
	if err != nil {
		t.Fatalf(testParseCategoryErrFmt, err)
	}

	if category != measureCategoryRealID {
		t.Fatalf(testCategoryGotFmt, category, measureCategoryRealID)
	}

	category, err = parseMeasureCategory(measureCategoryGoalText)
	if err != nil {
		t.Fatalf(testParseCategoryErrFmt, err)
	}

	if category != measureCategoryGoalID {
		t.Fatalf(testCategoryGotFmt, category, measureCategoryGoalID)
	}

	category, err = parseMeasureCategory(measureCategoryGoalID)
	if err != nil {
		t.Fatalf(testParseCategoryErrFmt, err)
	}

	if category != measureCategoryGoalID {
		t.Fatalf(testCategoryGotFmt, category, measureCategoryGoalID)
	}
}

// TestParseMeasureCategoryRejectsInvalid rejects invalid values.
func TestParseMeasureCategoryRejectsInvalid(t *testing.T) {
	t.Parallel()

	_, err := parseMeasureCategory("nope")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errInvalidMeasureCategory) {
		t.Fatalf("expected errInvalidMeasureCategory, got %v", err)
	}
}

// TestParseMeasureTypesMapsNames converts aliases to numeric codes.
func TestParseMeasureTypesMapsNames(t *testing.T) {
	t.Parallel()

	types, err := parseMeasureTypes(
		measureTypeWeight + measureTypeDelimiter + measureTypeBPSys,
	)
	if err != nil {
		t.Fatalf(testParseTypesErrFmt, err)
	}

	want := measureTypeWeightID + measureTypeDelimiter + measureTypeBPSysID
	if types != want {
		t.Fatalf(testTypesGotFmt, types, want)
	}
}

// TestParseMeasureTypesDedup removes duplicates.
func TestParseMeasureTypesDedup(t *testing.T) {
	t.Parallel()

	types, err := parseMeasureTypes(
		measureTypeWeight + measureTypeDelimiter + measureTypeDedup,
	)
	if err != nil {
		t.Fatalf(testParseTypesErrFmt, err)
	}

	if types != measureTypeWeightID {
		t.Fatalf(testTypesGotFmt, types, measureTypeWeightID)
	}
}

// TestParseMeasureTypesAllowsNumeric accepts numeric types.
func TestParseMeasureTypesAllowsNumeric(t *testing.T) {
	t.Parallel()

	types, err := parseMeasureTypes(
		measureTypeWeightID + measureTypeDelimiter + measureTypeBPSysID,
	)
	if err != nil {
		t.Fatalf(testParseTypesErrFmt, err)
	}

	want := measureTypeWeightID + measureTypeDelimiter + measureTypeBPSysID
	if types != want {
		t.Fatalf(testTypesGotFmt, types, want)
	}
}

// TestParseEpochRFC3339 parses RFC3339 values.
func TestParseEpochRFC3339(t *testing.T) {
	t.Parallel()

	input := "2025-12-30T12:34:56Z"

	epoch, err := parseEpoch(input)
	if err != nil {
		t.Fatalf(testParseEpochErrFmt, err)
	}

	want := time.Date(2025, 12, 30, 12, 34, 56, 0, time.UTC).Unix()
	if epoch != want {
		t.Fatalf(testEpochGotFmt, epoch, want)
	}
}

// TestBuildMeasureParamsLastUpdateConflict rejects mixed filters.
func TestBuildMeasureParamsLastUpdateConflict(t *testing.T) {
	t.Parallel()

	opts := measuresGetOptions{
		TimeRange: timeRangeOptions{
			Start: "2025-12-30T00:00:00Z",
			End:   emptyString,
		},
		Pagination: paginationOptions{
			Limit:  defaultInt,
			Offset: defaultInt,
		},
		User: userOption{
			UserID: emptyString,
		},
		LastUpdate: lastUpdateOption{
			LastUpdate: testLastUpdateValue,
		},
		Types:    emptyString,
		Category: emptyString,
	}

	_, err := buildMeasureParams(opts)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errLastUpdateConflict) {
		t.Fatalf("expected errLastUpdateConflict, got %v", err)
	}
}

// TestBuildMeasureParamsMapsFields validates common params.
func TestBuildMeasureParamsMapsFields(t *testing.T) {
	t.Parallel()

	opts := measuresGetOptions{
		TimeRange: timeRangeOptions{
			Start: "2025-12-30T00:00:00Z",
			End:   "2025-12-30T23:59:59Z",
		},
		Pagination: paginationOptions{
			Limit:  testLimitValue,
			Offset: testOffsetValue,
		},
		User: userOption{
			UserID: "user",
		},
		LastUpdate: lastUpdateOption{
			LastUpdate: defaultInt64,
		},
		Types:    measureTypeWeight,
		Category: measureCategoryRealText,
	}

	values, err := buildMeasureParams(opts)
	if err != nil {
		t.Fatalf(testBuildParamsErrFmt, err)
	}

	startEpoch := time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC).Unix()
	endEpoch := time.Date(2025, 12, 30, 23, 59, 59, 0, time.UTC).Unix()
	startValue := strconv.FormatInt(startEpoch, measureNumberBase10)
	endValue := strconv.FormatInt(endEpoch, measureNumberBase10)

	want := url.Values{
		measureTypeParam:      {measureTypeWeightID},
		measureCategoryParam:  {measureCategoryRealID},
		measureStartDateParam: {startValue},
		measureEndDateParam:   {endValue},
		measureLimitParam:     {strconv.Itoa(testLimitValue)},
		measureOffsetParam:    {strconv.Itoa(testOffsetValue)},
		measureUserIDParam:    {"user"},
	}

	for key, wantValues := range want {
		gotValues := values[key]
		if len(gotValues) != len(wantValues) ||
			gotValues[testFirstIndex] != wantValues[testFirstIndex] {
			t.Fatalf(testParamGotFmt, key, gotValues, wantValues)
		}
	}
}
