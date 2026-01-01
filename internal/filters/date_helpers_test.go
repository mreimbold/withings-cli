//nolint:testpackage // test unexported helpers.
package filters

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/mreimbold/withings-cli/internal/errs"
	"github.com/mreimbold/withings-cli/internal/params"
)

const (
	testDateValue        = "2025-12-30"
	testDateInvalidValue = "2025-13-40"
	testExpectErr        = "expected error"
	testErrFmt           = "err got %v want %v"
	testRangeValue       = "1"
	testNumberBase10     = 10
	testDefaultInt       = 0
	testEmptyString      = ""
	testYear             = 2025
	testMonth            = 12
	testDay              = 30
	testStartHour        = 8
	testEndHour          = 20
	testEpochRFC3339     = "2025-12-30T12:34:56Z"
)

// TestParseDateValueValid validates date parsing.
func TestParseDateValueValid(t *testing.T) {
	t.Parallel()

	got, err := ParseDateValue(testDateValue)
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

	_, err := ParseDateValue(testDateInvalidValue)
	if err == nil {
		t.Fatal(testExpectErr)
	}

	if !errors.Is(err, errs.ErrInvalidDate) {
		t.Fatalf(testErrFmt, err, errs.ErrInvalidDate)
	}
}

// TestResolveDateRangeConflict rejects date combined with start/end.
func TestResolveDateRangeConflict(t *testing.T) {
	t.Parallel()

	_, err := ResolveDateRange(
		params.Date{Date: testDateValue},
		params.TimeRange{Start: testRangeValue, End: testEmptyString},
		errs.ErrInvalidStartTime,
		errs.ErrInvalidEndTime,
	)
	if err == nil {
		t.Fatal(testExpectErr)
	}

	if !errors.Is(err, errs.ErrDateRangeConflict) {
		t.Fatalf(testErrFmt, err, errs.ErrDateRangeConflict)
	}
}

// TestResolveDateRangeDate uses date for both start/end.
func TestResolveDateRangeDate(t *testing.T) {
	t.Parallel()

	rangeValues, err := ResolveDateRange(
		params.Date{Date: testDateValue},
		params.TimeRange{Start: testEmptyString, End: testEmptyString},
		errs.ErrInvalidStartTime,
		errs.ErrInvalidEndTime,
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
		testDefaultInt,
		testDefaultInt,
		testDefaultInt,
		time.UTC,
	).Unix()
	endEpoch := time.Date(
		testYear,
		time.Month(testMonth),
		testDay,
		testEndHour,
		testDefaultInt,
		testDefaultInt,
		testDefaultInt,
		time.UTC,
	).Unix()

	rangeValues, err := ResolveDateRange(
		params.Date{Date: testEmptyString},
		params.TimeRange{
			Start: strconvFormatInt(startEpoch),
			End:   strconvFormatInt(endEpoch),
		},
		errs.ErrInvalidStartTime,
		errs.ErrInvalidEndTime,
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

// TestParseEpochRFC3339 parses RFC3339 values.
func TestParseEpochRFC3339(t *testing.T) {
	t.Parallel()

	epoch, err := ParseEpoch(testEpochRFC3339)
	if err != nil {
		t.Fatalf("parseEpoch: %v", err)
	}

	want := time.Date(2025, 12, 30, 12, 34, 56, 0, time.UTC).Unix()
	if epoch != want {
		t.Fatalf("epoch got %d want %d", epoch, want)
	}
}

// TestParseEpochDate parses YYYY-MM-DD values.
func TestParseEpochDate(t *testing.T) {
	t.Parallel()

	epoch, err := ParseEpoch(testDateValue)
	if err != nil {
		t.Fatalf("parseEpoch: %v", err)
	}

	want := time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC).Unix()
	if epoch != want {
		t.Fatalf("epoch got %d want %d", epoch, want)
	}
}

func strconvFormatInt(value int64) string {
	return strconv.FormatInt(value, testNumberBase10)
}
