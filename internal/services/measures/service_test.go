//nolint:testpackage // test unexported helpers.
package measures

import (
	"errors"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/mreimbold/withings-cli/internal/errs"
	"github.com/mreimbold/withings-cli/internal/params"
)

const (
	measureTypeWeight       = "weight"
	measureTypeBPSys        = "bp_sys"
	measureTypeWeightID     = "1"
	measureTypeBPSysID      = "10"
	measureTypeDedup        = "bodyweight"
	measureCategoryRealID   = "1"
	measureCategoryGoalID   = "2"
	testParseCategoryErrFmt = "parseCategory: %v"
	testCategoryGotFmt      = "category got %q want %q"
	testParseTypesErrFmt    = "parseTypes: %v"
	testTypesGotFmt         = "types got %q want %q"
	testBuildParamsErrFmt   = "buildParams: %v"
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
	testEmptyString         = ""
	testDefaultInt          = 0
	testDefaultInt64        = int64(0)
	testImperialWeightType  = 1
	testImperialWeightValue = int64(7500)
	testImperialWeightUnit  = -2
	testImperialWeightLb    = "165.35"
	testImperialLbUnit      = "lb"
	testRowsGotFmt          = "rows got %d want %d"
)

// TestParseCategory accepts text and numeric values.
func TestParseCategory(t *testing.T) {
	t.Parallel()

	category, err := parseCategory(categoryRealText)
	if err != nil {
		t.Fatalf(testParseCategoryErrFmt, err)
	}

	if category != measureCategoryRealID {
		t.Fatalf(testCategoryGotFmt, category, measureCategoryRealID)
	}

	category, err = parseCategory(categoryGoalText)
	if err != nil {
		t.Fatalf(testParseCategoryErrFmt, err)
	}

	if category != measureCategoryGoalID {
		t.Fatalf(testCategoryGotFmt, category, measureCategoryGoalID)
	}

	category, err = parseCategory(measureCategoryGoalID)
	if err != nil {
		t.Fatalf(testParseCategoryErrFmt, err)
	}

	if category != measureCategoryGoalID {
		t.Fatalf(testCategoryGotFmt, category, measureCategoryGoalID)
	}
}

// TestParseCategoryRejectsInvalid rejects invalid values.
func TestParseCategoryRejectsInvalid(t *testing.T) {
	t.Parallel()

	_, err := parseCategory("nope")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errInvalidMeasureCategory) {
		t.Fatalf("expected errInvalidMeasureCategory, got %v", err)
	}
}

// TestParseTypesMapsNames converts aliases to numeric codes.
func TestParseTypesMapsNames(t *testing.T) {
	t.Parallel()

	types, err := parseTypes(
		measureTypeWeight + typeDelimiter + measureTypeBPSys,
	)
	if err != nil {
		t.Fatalf(testParseTypesErrFmt, err)
	}

	want := measureTypeWeightID + typeDelimiter + measureTypeBPSysID
	if types != want {
		t.Fatalf(testTypesGotFmt, types, want)
	}
}

// TestParseTypesDedup removes duplicates.
func TestParseTypesDedup(t *testing.T) {
	t.Parallel()

	types, err := parseTypes(
		measureTypeWeight + typeDelimiter + measureTypeDedup,
	)
	if err != nil {
		t.Fatalf(testParseTypesErrFmt, err)
	}

	if types != measureTypeWeightID {
		t.Fatalf(testTypesGotFmt, types, measureTypeWeightID)
	}
}

// TestParseTypesAllowsNumeric accepts numeric types.
func TestParseTypesAllowsNumeric(t *testing.T) {
	t.Parallel()

	types, err := parseTypes(
		measureTypeWeightID + typeDelimiter + measureTypeBPSysID,
	)
	if err != nil {
		t.Fatalf(testParseTypesErrFmt, err)
	}

	want := measureTypeWeightID + typeDelimiter + measureTypeBPSysID
	if types != want {
		t.Fatalf(testTypesGotFmt, types, want)
	}
}

// TestBuildParamsLastUpdateConflict rejects mixed filters.
func TestBuildParamsLastUpdateConflict(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: "2025-12-30T00:00:00Z",
			End:   testEmptyString,
		},
		Pagination: params.Pagination{
			Limit:  testDefaultInt,
			Offset: testDefaultInt,
		},
		User: params.User{
			UserID: testEmptyString,
		},
		LastUpdate: params.LastUpdate{
			LastUpdate: testLastUpdateValue,
		},
		Types:    testEmptyString,
		Category: testEmptyString,
		Units:    testEmptyString,
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errs.ErrLastUpdateConflict) {
		t.Fatalf("expected errLastUpdateConflict, got %v", err)
	}
}

// TestBuildParamsMapsFields validates common params.
func TestBuildParamsMapsFields(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: "2025-12-30T00:00:00Z",
			End:   "2025-12-30T23:59:59Z",
		},
		Pagination: params.Pagination{
			Limit:  testLimitValue,
			Offset: testOffsetValue,
		},
		User: params.User{
			UserID: "user",
		},
		LastUpdate: params.LastUpdate{
			LastUpdate: testDefaultInt64,
		},
		Types:    measureTypeWeight,
		Category: categoryRealText,
		Units:    testEmptyString,
	}

	values, err := buildParams(opts)
	if err != nil {
		t.Fatalf(testBuildParamsErrFmt, err)
	}

	startEpochSec := time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC).Unix()
	endEpochSec := time.Date(2025, 12, 30, 23, 59, 59, 0, time.UTC).Unix()
	startValue := strconv.FormatInt(startEpochSec, numberBase10)
	endValue := strconv.FormatInt(endEpochSec, numberBase10)

	want := url.Values{
		typeParam:      {measureTypeWeightID},
		categoryParam:  {measureCategoryRealID},
		startDateParam: {startValue},
		endDateParam:   {endValue},
		limitParam:     {strconv.Itoa(testLimitValue)},
		offsetParam:    {strconv.Itoa(testOffsetValue)},
		userIDParam:    {"user"},
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

// TestBuildRows builds a row per measure entry.
func TestBuildRows(t *testing.T) {
	t.Parallel()

	rows := buildRows(Options{
		TimeRange: params.TimeRange{
			Start: testEmptyString,
			End:   testEmptyString,
		},
		Pagination: params.Pagination{
			Limit:  testDefaultInt,
			Offset: testDefaultInt,
		},
		User:       params.User{UserID: testEmptyString},
		LastUpdate: params.LastUpdate{LastUpdate: testDefaultInt64},
		Types:      testEmptyString,
		Category:   testEmptyString,
		Units:      testEmptyString,
	}, testBody())
	assertSingleMeasureRow(t, rows)
}

// TestBuildRowsImperial converts weight to lbs.
//
//nolint:dupl // intentionally parallel to NonWeight; tests opposite behaviour
func TestBuildRowsImperial(t *testing.T) {
	t.Parallel()

	body := testBody()
	m := &body.MeasureGroups[testFirstIndex].Measures[testFirstIndex]
	m.Type = testImperialWeightType
	m.Value = testImperialWeightValue
	m.Unit = testImperialWeightUnit

	opts := Options{
		TimeRange: params.TimeRange{
			Start: testEmptyString,
			End:   testEmptyString,
		},
		Pagination: params.Pagination{
			Limit:  testDefaultInt,
			Offset: testDefaultInt,
		},
		User:       params.User{UserID: testEmptyString},
		LastUpdate: params.LastUpdate{LastUpdate: testDefaultInt64},
		Types:      testEmptyString,
		Category:   testEmptyString,
		Units:      unitImperial,
	}
	rows := buildRows(opts, body)

	if len(rows) != testMeasureRowCount {
		t.Fatalf(testRowsGotFmt, len(rows), testMeasureRowCount)
	}

	row := rows[testFirstIndex]
	if row.Value != testImperialWeightLb {
		t.Fatalf("value got %q want %q", row.Value, testImperialWeightLb)
	}

	if row.Unit != testImperialLbUnit {
		t.Fatalf("unit got %q want %q", row.Unit, testImperialLbUnit)
	}
}

// TestBuildRowsImperial_NonWeight does not convert non-weight units.
//
//nolint:dupl // intentionally parallel to Imperial; tests opposite behaviour
func TestBuildRowsImperial_NonWeight(t *testing.T) {
	t.Parallel()

	body := testBody()
	m := &body.MeasureGroups[testFirstIndex].Measures[testFirstIndex]
	m.Type = testMeasureType
	m.Value = testScaleNoValue
	m.Unit = testScaleNoUnit

	opts := Options{
		TimeRange: params.TimeRange{
			Start: testEmptyString,
			End:   testEmptyString,
		},
		Pagination: params.Pagination{
			Limit:  testDefaultInt,
			Offset: testDefaultInt,
		},
		User:       params.User{UserID: testEmptyString},
		LastUpdate: params.LastUpdate{LastUpdate: testDefaultInt64},
		Types:      testEmptyString,
		Category:   testEmptyString,
		Units:      unitImperial,
	}
	rows := buildRows(opts, body)

	if len(rows) != testMeasureRowCount {
		t.Fatalf(testRowsGotFmt, len(rows), testMeasureRowCount)
	}

	row := rows[testFirstIndex]
	if row.Value != testScaleNoWant {
		t.Fatalf("value got %q want %q", row.Value, testScaleNoWant)
	}

	if row.Unit != testMeasureExpectedUnit {
		t.Fatalf("unit got %q want %q", row.Unit, testMeasureExpectedUnit)
	}
}

// TestBuildParams_InvalidUnits rejects invalid unit systems.
func TestBuildParams_InvalidUnits(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: testEmptyString,
			End:   testEmptyString,
		},
		Pagination: params.Pagination{
			Limit:  testDefaultInt,
			Offset: testDefaultInt,
		},
		User:       params.User{UserID: testEmptyString},
		LastUpdate: params.LastUpdate{LastUpdate: testDefaultInt64},
		Types:      testEmptyString,
		Category:   testEmptyString,
		Units:      "invalid",
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal("buildParams should have failed with invalid units")
	}

	if !errors.Is(err, errInvalidUnits) {
		t.Fatalf("error got %v want errInvalidUnits", err)
	}
}

func testBody() body {
	epochSec := time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC).Unix()

	return body{
		UpdateTime: testDefaultInt64,
		Timezone:   "UTC",
		MeasureGroups: []group{
			{
				GroupID:  testDefaultInt64,
				Attrib:   testDefaultInt,
				Date:     epochSec,
				Category: testMeasureCategory,
				Measures: []item{
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

func assertSingleMeasureRow(t *testing.T, rows []row) {
	t.Helper()

	if len(rows) != testMeasureRowCount {
		t.Fatalf("rows got %d want %d", len(rows), testMeasureRowCount)
	}

	row := rows[testFirstIndex]
	assertMeasureValue(t, "time", row.Time, testMeasureExpectedTime)
	assertMeasureValue(t, "type", row.Type, measureTypeBPSys)
	assertMeasureValue(t, "value", row.Value, testScaleNoWant)
	assertMeasureValue(t, "unit", row.Unit, testMeasureExpectedUnit)
	assertMeasureValue(t, "category", row.Category, categoryRealText)
}

func assertMeasureValue(t *testing.T, label, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s got %q want %q", label, got, want)
	}
}
