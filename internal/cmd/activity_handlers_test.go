//nolint:testpackage // test unexported helpers in cmd.
package cmd

import (
	"errors"
	"testing"
	"time"
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
)

// TestActivityServiceForBase handles base URLs with and without /v2.
func TestActivityServiceForBase(t *testing.T) {
	t.Parallel()

	got := activityServiceForBase(activityTestBaseNoV2)
	if got != activityServiceName {
		t.Fatalf(activityTestServiceFmt, got, activityServiceName)
	}

	got = activityServiceForBase(activityTestBaseV2)
	if got != activityServiceShort {
		t.Fatalf(activityTestServiceFmt, got, activityServiceShort)
	}

	got = activityServiceForBase(activityTestBaseV2Sl)
	if got != activityServiceShort {
		t.Fatalf(activityTestServiceFmt, got, activityServiceShort)
	}
}

// TestBuildActivityParamsDate builds date-scoped params.
func TestBuildActivityParamsDate(t *testing.T) {
	t.Parallel()

	opts := activityGetOptions{
		TimeRange: timeRangeOptions{Start: emptyString, End: emptyString},
		Date:      dateOption{Date: activityTestDate},
		Pagination: paginationOptions{
			Limit:  activityTestLimit,
			Offset: activityTestOffset,
		},
		User:       userOption{UserID: activityTestUserID},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
	}

	values, err := buildActivityParams(opts)
	if err != nil {
		t.Fatalf("buildActivityParams: %v", err)
	}

	assertParam(
		t,
		values.Get(activityStartDateParam),
		activityTestDate,
		"startdateymd",
	)
	assertParam(
		t,
		values.Get(activityEndDateParam),
		activityTestDate,
		"enddateymd",
	)
	assertParam(t, values.Get(activityLimitParam), "25", "limit")
	assertParam(t, values.Get(activityOffsetParam), "10", "offset")
	assertParam(
		t,
		values.Get(activityUserIDParam),
		activityTestUserID,
		"userid",
	)
}

// TestBuildActivityParamsTimeRange converts epoch range to dates.
func TestBuildActivityParamsTimeRange(t *testing.T) {
	t.Parallel()

	startEpoch := time.Date(
		activityTestYear,
		time.Month(activityTestMonth),
		activityTestDay,
		activityTestStartHour,
		defaultInt,
		defaultInt,
		defaultInt,
		time.UTC,
	).Unix()
	endEpoch := time.Date(
		activityTestYear,
		time.Month(activityTestMonth),
		activityTestDay,
		activityTestEndHour,
		defaultInt,
		defaultInt,
		defaultInt,
		time.UTC,
	).Unix()

	opts := activityGetOptions{
		TimeRange: timeRangeOptions{
			Start: strconvFormatInt(startEpoch),
			End:   strconvFormatInt(endEpoch),
		},
		Date:       dateOption{Date: emptyString},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
	}

	values, err := buildActivityParams(opts)
	if err != nil {
		t.Fatalf("buildActivityParams: %v", err)
	}

	assertParam(
		t,
		values.Get(activityStartDateParam),
		activityTestDate,
		"startdateymd",
	)
	assertParam(
		t,
		values.Get(activityEndDateParam),
		activityTestDate,
		"enddateymd",
	)
}

// TestBuildActivityParamsLastUpdateConflict rejects mixed filters.
func TestBuildActivityParamsLastUpdateConflict(t *testing.T) {
	t.Parallel()

	opts := activityGetOptions{
		TimeRange: timeRangeOptions{
			Start: activityTestRangeValue,
			End:   emptyString,
		},
		Date:       dateOption{Date: emptyString},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: activityTestLastUpdate},
	}

	_, err := buildActivityParams(opts)
	if err == nil {
		t.Fatal(activityTestExpectErr)
	}

	if !errors.Is(err, errLastUpdateConflict) {
		t.Fatalf(activityTestErrFmt, err, errLastUpdateConflict)
	}
}

// TestBuildActivityParamsDateConflict rejects date + start.
func TestBuildActivityParamsDateConflict(t *testing.T) {
	t.Parallel()

	opts := activityGetOptions{
		TimeRange: timeRangeOptions{
			Start: activityTestRangeValue,
			End:   emptyString,
		},
		Date:       dateOption{Date: activityTestDate},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
	}

	_, err := buildActivityParams(opts)
	if err == nil {
		t.Fatal(activityTestExpectErr)
	}

	if !errors.Is(err, errDateRangeConflict) {
		t.Fatalf(activityTestErrFmt, err, errDateRangeConflict)
	}
}

// TestBuildActivityParamsInvalidDate rejects invalid date.
func TestBuildActivityParamsInvalidDate(t *testing.T) {
	t.Parallel()

	opts := activityGetOptions{
		TimeRange:  timeRangeOptions{Start: emptyString, End: emptyString},
		Date:       dateOption{Date: activityTestDateBad},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
	}

	_, err := buildActivityParams(opts)
	if err == nil {
		t.Fatal(activityTestExpectErr)
	}

	if !errors.Is(err, errInvalidDate) {
		t.Fatalf(activityTestErrFmt, err, errInvalidDate)
	}
}
