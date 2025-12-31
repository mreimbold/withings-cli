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
	testScaleNoValue        = int64(120)
	testScaleNoUnit         = 0
	testScaleNoWant         = "120"
	testScalePositiveValue  = int64(123)
	testScalePositiveUnit   = 2
	testScalePositiveWant   = "12300"
	testScaleNegativeValue  = int64(84500)
	testScaleNegativeUnit   = -3
	testScaleNegativeWant   = "84.5"
	testScaleSmallValue     = int64(5)
	testScaleSmallUnit      = -3
	testScaleSmallWant      = "0.005"
	testScaleTrimValue      = int64(1000)
	testScaleTrimUnit       = -3
	testScaleTrimWant       = "1"
	testScaleNegValue       = int64(-123)
	testScaleNegUnit        = -2
	testScaleNegWant        = "-1.23"
	testMeasureRowCount     = 1
	testMeasureCategory     = 1
	testMeasureType         = 10
	testMeasureValue        = int64(1200)
	testMeasureUnit         = -1
	testMeasureExpectedTime = "2025-12-30T00:00:00Z"
	testMeasureExpectedUnit = "mmHg"
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

// TestFormatScaledValue covers scaling math.
func TestFormatScaledValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value int64
		unit  int
		want  string
	}{
		{
			name:  "no-scale",
			value: testScaleNoValue,
			unit:  testScaleNoUnit,
			want:  testScaleNoWant,
		},
		{
			name:  "positive-scale",
			value: testScalePositiveValue,
			unit:  testScalePositiveUnit,
			want:  testScalePositiveWant,
		},
		{
			name:  "negative-scale",
			value: testScaleNegativeValue,
			unit:  testScaleNegativeUnit,
			want:  testScaleNegativeWant,
		},
		{
			name:  "small-negative",
			value: testScaleSmallValue,
			unit:  testScaleSmallUnit,
			want:  testScaleSmallWant,
		},
		{
			name:  "trim-zeros",
			value: testScaleTrimValue,
			unit:  testScaleTrimUnit,
			want:  testScaleTrimWant,
		},
		{
			name:  "negative-value",
			value: testScaleNegValue,
			unit:  testScaleNegUnit,
			want:  testScaleNegWant,
		},
	}

	for _, test := range cases {
		if got := formatScaledValue(test.value, test.unit); got != test.want {
			t.Fatalf("%s got %q want %q", test.name, got, test.want)
		}
	}
}

// TestBuildMeasureRows builds a row per measure entry.
func TestBuildMeasureRows(t *testing.T) {
	t.Parallel()

	rows := buildMeasureRows(testMeasureBody())
	assertSingleMeasureRow(t, rows)
}

func testMeasureBody() measuresBody {
	epoch := time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC).Unix()

	return measuresBody{
		UpdateTime: defaultInt64,
		Timezone:   "UTC",
		MeasureGroups: []measureGroup{
			{
				GroupID:  defaultInt64,
				Attrib:   defaultInt,
				Date:     epoch,
				Category: testMeasureCategory,
				Measures: []measureItem{
					{
						Type:  testMeasureType,
						Value: testMeasureValue,
						Unit:  testMeasureUnit,
					},
				},
			},
		},
	}
}

func assertSingleMeasureRow(t *testing.T, rows []measureRow) {
	t.Helper()

	if len(rows) != testMeasureRowCount {
		t.Fatalf("rows got %d want %d", len(rows), testMeasureRowCount)
	}

	row := rows[testFirstIndex]
	assertMeasureValue(t, "time", row.Time, testMeasureExpectedTime)
	assertMeasureValue(t, "type", row.Type, measureTypeBPSys)
	assertMeasureValue(t, "value", row.Value, testScaleNoWant)
	assertMeasureValue(t, "unit", row.Unit, testMeasureExpectedUnit)
	assertMeasureValue(t, "category", row.Category, measureCategoryRealText)
}

func assertMeasureValue(t *testing.T, label, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s got %q want %q", label, got, want)
	}
}
