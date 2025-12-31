//nolint:testpackage // test unexported helpers.
package sleep

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/mreimbold/withings-cli/internal/errs"
	"github.com/mreimbold/withings-cli/internal/params"
)

const (
	sleepTestDate       = "2025-12-30"
	sleepTestDateBad    = "2025-13-01"
	sleepTestUserID     = "user-123"
	sleepTestLimit      = 20
	sleepTestOffset     = 5
	sleepTestModel      = 2
	sleepTestLastUpdate = 100
	sleepTestYear       = 2025
	sleepTestMonth      = 12
	sleepTestDay        = 30
	sleepTestStartHour  = 1
	sleepTestEndHour    = 8
	sleepTestBaseNoV2   = "https://wbsapi.withings.net"
	sleepTestBaseV2     = "https://wbsapi.withings.net/v2"
	sleepTestBaseV2Sl   = "https://wbsapi.withings.net/v2/"
	sleepTestServiceFmt = "service got %q want %q"
	sleepTestErrFmt     = "err got %v want %v"
	sleepTestExpectErr  = "expected error"
	sleepTestRangeValue = "1"
	sleepTestEmpty      = ""
	sleepTestDefaultInt = 0
	sleepTestBase10     = 10
)

// TestSleepServiceForBase handles base URLs with and without /v2.
func TestSleepServiceForBase(t *testing.T) {
	t.Parallel()

	got := serviceForBase(sleepTestBaseNoV2)
	if got != serviceName {
		t.Fatalf(sleepTestServiceFmt, got, serviceName)
	}

	got = serviceForBase(sleepTestBaseV2)
	if got != serviceShort {
		t.Fatalf(sleepTestServiceFmt, got, serviceShort)
	}

	got = serviceForBase(sleepTestBaseV2Sl)
	if got != serviceShort {
		t.Fatalf(sleepTestServiceFmt, got, serviceShort)
	}
}

// TestBuildParamsDate builds date-scoped params.
func TestBuildParamsDate(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{Start: sleepTestEmpty, End: sleepTestEmpty},
		Date:      params.Date{Date: sleepTestDate},
		Pagination: params.Pagination{
			Limit:  sleepTestLimit,
			Offset: sleepTestOffset,
		},
		User:       params.User{UserID: sleepTestUserID},
		LastUpdate: params.LastUpdate{LastUpdate: sleepTestDefaultInt},
		Model:      sleepTestModel,
	}

	values, err := buildParams(opts)
	if err != nil {
		t.Fatalf("buildParams: %v", err)
	}

	assertParam(
		t,
		values.Get(startDateParam),
		sleepTestDate,
		"startdateymd",
	)
	assertParam(
		t,
		values.Get(endDateParam),
		sleepTestDate,
		"enddateymd",
	)
	assertParam(t, values.Get(limitParam), "20", "limit")
	assertParam(t, values.Get(offsetParam), "5", "offset")
	assertParam(t, values.Get(userIDParam), sleepTestUserID, "userid")
	assertParam(t, values.Get(modelParam), "2", "model")
}

// TestBuildParamsTimeRange converts epoch range to dates.
func TestBuildParamsTimeRange(t *testing.T) {
	t.Parallel()

	startEpoch := time.Date(
		sleepTestYear,
		time.Month(sleepTestMonth),
		sleepTestDay,
		sleepTestStartHour,
		sleepTestDefaultInt,
		sleepTestDefaultInt,
		sleepTestDefaultInt,
		time.UTC,
	).Unix()
	endEpoch := time.Date(
		sleepTestYear,
		time.Month(sleepTestMonth),
		sleepTestDay,
		sleepTestEndHour,
		sleepTestDefaultInt,
		sleepTestDefaultInt,
		sleepTestDefaultInt,
		time.UTC,
	).Unix()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: strconv.FormatInt(startEpoch, sleepTestBase10),
			End:   strconv.FormatInt(endEpoch, sleepTestBase10),
		},
		Date: params.Date{Date: sleepTestEmpty},
		Pagination: params.Pagination{
			Limit:  sleepTestDefaultInt,
			Offset: sleepTestDefaultInt,
		},
		User:       params.User{UserID: sleepTestEmpty},
		LastUpdate: params.LastUpdate{LastUpdate: sleepTestDefaultInt},
		Model:      sleepTestDefaultInt,
	}

	values, err := buildParams(opts)
	if err != nil {
		t.Fatalf("buildParams: %v", err)
	}

	assertParam(
		t,
		values.Get(startDateParam),
		sleepTestDate,
		"startdateymd",
	)
	assertParam(
		t,
		values.Get(endDateParam),
		sleepTestDate,
		"enddateymd",
	)
}

// TestBuildParamsLastUpdateConflict rejects mixed filters.
func TestBuildParamsLastUpdateConflict(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: sleepTestRangeValue,
			End:   sleepTestEmpty,
		},
		Date: params.Date{Date: sleepTestEmpty},
		Pagination: params.Pagination{
			Limit:  sleepTestDefaultInt,
			Offset: sleepTestDefaultInt,
		},
		User:       params.User{UserID: sleepTestEmpty},
		LastUpdate: params.LastUpdate{LastUpdate: sleepTestLastUpdate},
		Model:      sleepTestDefaultInt,
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal(sleepTestExpectErr)
	}

	if !errors.Is(err, errs.ErrLastUpdateConflict) {
		t.Fatalf(sleepTestErrFmt, err, errs.ErrLastUpdateConflict)
	}
}

// TestBuildParamsDateConflict rejects date + start.
func TestBuildParamsDateConflict(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: sleepTestRangeValue,
			End:   sleepTestEmpty,
		},
		Date: params.Date{Date: sleepTestDate},
		Pagination: params.Pagination{
			Limit:  sleepTestDefaultInt,
			Offset: sleepTestDefaultInt,
		},
		User:       params.User{UserID: sleepTestEmpty},
		LastUpdate: params.LastUpdate{LastUpdate: sleepTestDefaultInt},
		Model:      sleepTestDefaultInt,
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal(sleepTestExpectErr)
	}

	if !errors.Is(err, errs.ErrDateRangeConflict) {
		t.Fatalf(sleepTestErrFmt, err, errs.ErrDateRangeConflict)
	}
}

// TestBuildParamsInvalidDate rejects invalid date.
func TestBuildParamsInvalidDate(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: sleepTestEmpty,
			End:   sleepTestEmpty,
		},
		Date: params.Date{Date: sleepTestDateBad},
		Pagination: params.Pagination{
			Limit:  sleepTestDefaultInt,
			Offset: sleepTestDefaultInt,
		},
		User:       params.User{UserID: sleepTestEmpty},
		LastUpdate: params.LastUpdate{LastUpdate: sleepTestDefaultInt},
		Model:      sleepTestDefaultInt,
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal(sleepTestExpectErr)
	}

	if !errors.Is(err, errs.ErrInvalidDate) {
		t.Fatalf(sleepTestErrFmt, err, errs.ErrInvalidDate)
	}
}

func assertParam(t *testing.T, got, want, label string) {
	t.Helper()

	if got != want {
		t.Fatalf("param %s got %q want %q", label, got, want)
	}
}
