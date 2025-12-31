//nolint:testpackage // test unexported helpers in cmd.
package cmd

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

const (
	testDateValue        = "2025-12-30"
	testDateInvalidValue = "2025-13-40"
	testExpectErr        = "expected error"
	testErrFmt           = "err got %v want %v"
	testRangeValue       = "1"
	testYear             = 2025
	testMonth            = 12
	testDay              = 30
	testStartHour        = 8
	testEndHour          = 20
)

// TestParseDateValueValid validates date parsing.
func TestParseDateValueValid(t *testing.T) {
	t.Parallel()

	got, err := parseDateValue(testDateValue)
	if err != nil {
		t.Fatalf("parseDateValue: %v", err)
	}

	if got != testDateValue {
		t.Fatalf("date got %q want %q", got, testDateValue)
	}
}

// TestParseDateValueInvalid rejects non-YYYY-MM-DD values.
func TestParseDateValueInvalid(t *testing.T) {
	t.Parallel()

	_, err := parseDateValue(testDateInvalidValue)
	if err == nil {
		t.Fatal(testExpectErr)
	}

	if !errors.Is(err, errInvalidDate) {
		t.Fatalf(testErrFmt, err, errInvalidDate)
	}
}

// TestResolveDateRangeConflict rejects date combined with start/end.
func TestResolveDateRangeConflict(t *testing.T) {
	t.Parallel()

	_, err := resolveDateRange(
		dateOption{Date: testDateValue},
		timeRangeOptions{Start: testRangeValue, End: emptyString},
		errInvalidStartTime,
		errInvalidEndTime,
	)
	if err == nil {
		t.Fatal(testExpectErr)
	}

	if !errors.Is(err, errDateRangeConflict) {
		t.Fatalf(testErrFmt, err, errDateRangeConflict)
	}
}

// TestResolveDateRangeDate uses date for both start/end.
func TestResolveDateRangeDate(t *testing.T) {
	t.Parallel()

	rangeValues, err := resolveDateRange(
		dateOption{Date: testDateValue},
		timeRangeOptions{Start: emptyString, End: emptyString},
		errInvalidStartTime,
		errInvalidEndTime,
	)
	if err != nil {
		t.Fatalf("resolveDateRange: %v", err)
	}

	if rangeValues.Start != testDateValue {
		t.Fatalf("start got %q want %q", rangeValues.Start, testDateValue)
	}

	if rangeValues.End != testDateValue {
		t.Fatalf("end got %q want %q", rangeValues.End, testDateValue)
	}
}

// TestResolveDateRangeTimeRange formats epoch times as dates.
func TestResolveDateRangeTimeRange(t *testing.T) {
	t.Parallel()

	startEpoch := time.Date(
		testYear,
		time.Month(testMonth),
		testDay,
		testStartHour,
		defaultInt,
		defaultInt,
		defaultInt,
		time.UTC,
	).Unix()
	endEpoch := time.Date(
		testYear,
		time.Month(testMonth),
		testDay,
		testEndHour,
		defaultInt,
		defaultInt,
		defaultInt,
		time.UTC,
	).Unix()

	rangeValues, err := resolveDateRange(
		dateOption{Date: emptyString},
		timeRangeOptions{
			Start: strconvFormatInt(startEpoch),
			End:   strconvFormatInt(endEpoch),
		},
		errInvalidStartTime,
		errInvalidEndTime,
	)
	if err != nil {
		t.Fatalf("resolveDateRange: %v", err)
	}

	if rangeValues.Start != testDateValue {
		t.Fatalf("start got %q want %q", rangeValues.Start, testDateValue)
	}

	if rangeValues.End != testDateValue {
		t.Fatalf("end got %q want %q", rangeValues.End, testDateValue)
	}
}

func strconvFormatInt(value int64) string {
	return strconv.FormatInt(value, measureNumberBase10)
}
