//nolint:testpackage // test unexported helpers.
package activity

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/mreimbold/withings-cli/internal/errs"
	"github.com/mreimbold/withings-cli/internal/params"
)

const (
	activityTestDate       = "2025-12-30"
	activityTestDateBad    = "2025-13-01"
	activityTestUserID     = "user-123"
	activityTestLimit      = 25
	activityTestOffset     = 10
	activityTestLastUpdate = 100
	activityTestYear       = 2025
	activityTestMonth      = 12
	activityTestDay        = 30
	activityTestStartHour  = 4
	activityTestEndHour    = 12
	activityTestBaseNoV2   = "https://wbsapi.withings.net"
	activityTestBaseV2     = "https://wbsapi.withings.net/v2"
	activityTestBaseV2Sl   = "https://wbsapi.withings.net/v2/"
	activityTestServiceFmt = "service got %q want %q"
	activityTestErrFmt     = "err got %v want %v"
	activityTestExpectErr  = "expected error"
	activityTestRangeValue = "1"
	activityTestEmpty      = ""
	activityTestDefaultInt = 0
	activityTestBase10     = 10
)

// TestActivityServiceForBase handles base URLs with and without /v2.
func TestActivityServiceForBase(t *testing.T) {
	t.Parallel()

	got := serviceForBase(activityTestBaseNoV2)
	if got != serviceName {
		t.Fatalf(activityTestServiceFmt, got, serviceName)
	}

	got = serviceForBase(activityTestBaseV2)
	if got != serviceShort {
		t.Fatalf(activityTestServiceFmt, got, serviceShort)
	}

	got = serviceForBase(activityTestBaseV2Sl)
	if got != serviceShort {
		t.Fatalf(activityTestServiceFmt, got, serviceShort)
	}
}

// TestBuildParamsDate builds date-scoped params.
func TestBuildParamsDate(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: activityTestEmpty,
			End:   activityTestEmpty,
		},
		Date: params.Date{Date: activityTestDate},
		Pagination: params.Pagination{
			Limit:  activityTestLimit,
			Offset: activityTestOffset,
		},
		User:       params.User{UserID: activityTestUserID},
		LastUpdate: params.LastUpdate{LastUpdate: activityTestDefaultInt},
	}

	values, err := buildParams(opts)
	if err != nil {
		t.Fatalf("buildParams: %v", err)
	}

	assertParam(
		t,
		values.Get(startDateParam),
		activityTestDate,
		"startdateymd",
	)
	assertParam(
		t,
		values.Get(endDateParam),
		activityTestDate,
		"enddateymd",
	)
	assertParam(t, values.Get(limitParam), "25", "limit")
	assertParam(t, values.Get(offsetParam), "10", "offset")
	assertParam(
		t,
		values.Get(userIDParam),
		activityTestUserID,
		"userid",
	)
}

// TestBuildParamsTimeRange converts epoch range to dates.
func TestBuildParamsTimeRange(t *testing.T) {
	t.Parallel()

	startEpoch := time.Date(
		activityTestYear,
		time.Month(activityTestMonth),
		activityTestDay,
		activityTestStartHour,
		activityTestDefaultInt,
		activityTestDefaultInt,
		activityTestDefaultInt,
		time.UTC,
	).Unix()
	endEpoch := time.Date(
		activityTestYear,
		time.Month(activityTestMonth),
		activityTestDay,
		activityTestEndHour,
		activityTestDefaultInt,
		activityTestDefaultInt,
		activityTestDefaultInt,
		time.UTC,
	).Unix()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: strconv.FormatInt(startEpoch, activityTestBase10),
			End:   strconv.FormatInt(endEpoch, activityTestBase10),
		},
		Date: params.Date{Date: activityTestEmpty},
		Pagination: params.Pagination{
			Limit:  activityTestDefaultInt,
			Offset: activityTestDefaultInt,
		},
		User:       params.User{UserID: activityTestEmpty},
		LastUpdate: params.LastUpdate{LastUpdate: activityTestDefaultInt},
	}

	values, err := buildParams(opts)
	if err != nil {
		t.Fatalf("buildParams: %v", err)
	}

	assertParam(
		t,
		values.Get(startDateParam),
		activityTestDate,
		"startdateymd",
	)
	assertParam(
		t,
		values.Get(endDateParam),
		activityTestDate,
		"enddateymd",
	)
}

// TestBuildParamsLastUpdateConflict rejects mixed filters.
func TestBuildParamsLastUpdateConflict(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: activityTestRangeValue,
			End:   activityTestEmpty,
		},
		Date: params.Date{Date: activityTestEmpty},
		Pagination: params.Pagination{
			Limit:  activityTestDefaultInt,
			Offset: activityTestDefaultInt,
		},
		User:       params.User{UserID: activityTestEmpty},
		LastUpdate: params.LastUpdate{LastUpdate: activityTestLastUpdate},
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal(activityTestExpectErr)
	}

	if !errors.Is(err, errs.ErrLastUpdateConflict) {
		t.Fatalf(activityTestErrFmt, err, errs.ErrLastUpdateConflict)
	}
}

// TestBuildParamsDateConflict rejects date + start.
func TestBuildParamsDateConflict(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: activityTestRangeValue,
			End:   activityTestEmpty,
		},
		Date: params.Date{Date: activityTestDate},
		Pagination: params.Pagination{
			Limit:  activityTestDefaultInt,
			Offset: activityTestDefaultInt,
		},
		User:       params.User{UserID: activityTestEmpty},
		LastUpdate: params.LastUpdate{LastUpdate: activityTestDefaultInt},
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal(activityTestExpectErr)
	}

	if !errors.Is(err, errs.ErrDateRangeConflict) {
		t.Fatalf(activityTestErrFmt, err, errs.ErrDateRangeConflict)
	}
}

// TestBuildParamsInvalidDate rejects invalid date.
func TestBuildParamsInvalidDate(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: activityTestEmpty,
			End:   activityTestEmpty,
		},
		Date: params.Date{Date: activityTestDateBad},
		Pagination: params.Pagination{
			Limit:  activityTestDefaultInt,
			Offset: activityTestDefaultInt,
		},
		User:       params.User{UserID: activityTestEmpty},
		LastUpdate: params.LastUpdate{LastUpdate: activityTestDefaultInt},
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal(activityTestExpectErr)
	}

	if !errors.Is(err, errs.ErrInvalidDate) {
		t.Fatalf(activityTestErrFmt, err, errs.ErrInvalidDate)
	}
}

func assertParam(t *testing.T, got, want, label string) {
	t.Helper()

	if got != want {
		t.Fatalf("param %s got %q want %q", label, got, want)
	}
}
