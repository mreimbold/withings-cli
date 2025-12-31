//nolint:testpackage // test unexported helpers in cmd.
package cmd

import (
	"errors"
	"testing"
	"time"
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
)

// TestSleepServiceForBase handles base URLs with and without /v2.
func TestSleepServiceForBase(t *testing.T) {
	t.Parallel()

	got := sleepServiceForBase(sleepTestBaseNoV2)
	if got != sleepServiceName {
		t.Fatalf(sleepTestServiceFmt, got, sleepServiceName)
	}

	got = sleepServiceForBase(sleepTestBaseV2)
	if got != sleepServiceShort {
		t.Fatalf(sleepTestServiceFmt, got, sleepServiceShort)
	}

	got = sleepServiceForBase(sleepTestBaseV2Sl)
	if got != sleepServiceShort {
		t.Fatalf(sleepTestServiceFmt, got, sleepServiceShort)
	}
}

// TestBuildSleepParamsDate builds date-scoped params.
func TestBuildSleepParamsDate(t *testing.T) {
	t.Parallel()

	opts := sleepGetOptions{
		TimeRange: timeRangeOptions{Start: emptyString, End: emptyString},
		Date:      dateOption{Date: sleepTestDate},
		Pagination: paginationOptions{
			Limit:  sleepTestLimit,
			Offset: sleepTestOffset,
		},
		User:       userOption{UserID: sleepTestUserID},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
		Model:      sleepTestModel,
	}

	values, err := buildSleepParams(opts)
	if err != nil {
		t.Fatalf("buildSleepParams: %v", err)
	}

	assertParam(
		t,
		values.Get(sleepStartDateParam),
		sleepTestDate,
		"startdateymd",
	)
	assertParam(
		t,
		values.Get(sleepEndDateParam),
		sleepTestDate,
		"enddateymd",
	)
	assertParam(t, values.Get(sleepLimitParam), "20", "limit")
	assertParam(t, values.Get(sleepOffsetParam), "5", "offset")
	assertParam(t, values.Get(sleepUserIDParam), sleepTestUserID, "userid")
	assertParam(t, values.Get(sleepModelParam), "2", "model")
}

// TestBuildSleepParamsTimeRange converts epoch range to dates.
func TestBuildSleepParamsTimeRange(t *testing.T) {
	t.Parallel()

	startEpoch := time.Date(
		sleepTestYear,
		time.Month(sleepTestMonth),
		sleepTestDay,
		sleepTestStartHour,
		defaultInt,
		defaultInt,
		defaultInt,
		time.UTC,
	).Unix()
	endEpoch := time.Date(
		sleepTestYear,
		time.Month(sleepTestMonth),
		sleepTestDay,
		sleepTestEndHour,
		defaultInt,
		defaultInt,
		defaultInt,
		time.UTC,
	).Unix()

	opts := sleepGetOptions{
		TimeRange: timeRangeOptions{
			Start: strconvFormatInt(startEpoch),
			End:   strconvFormatInt(endEpoch),
		},
		Date:       dateOption{Date: emptyString},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
		Model:      defaultInt,
	}

	values, err := buildSleepParams(opts)
	if err != nil {
		t.Fatalf("buildSleepParams: %v", err)
	}

	assertParam(
		t,
		values.Get(sleepStartDateParam),
		sleepTestDate,
		"startdateymd",
	)
	assertParam(
		t,
		values.Get(sleepEndDateParam),
		sleepTestDate,
		"enddateymd",
	)
}

// TestBuildSleepParamsLastUpdateConflict rejects mixed filters.
func TestBuildSleepParamsLastUpdateConflict(t *testing.T) {
	t.Parallel()

	opts := sleepGetOptions{
		TimeRange: timeRangeOptions{
			Start: sleepTestRangeValue,
			End:   emptyString,
		},
		Date:       dateOption{Date: emptyString},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: sleepTestLastUpdate},
		Model:      defaultInt,
	}

	_, err := buildSleepParams(opts)
	if err == nil {
		t.Fatal(sleepTestExpectErr)
	}

	if !errors.Is(err, errLastUpdateConflict) {
		t.Fatalf(sleepTestErrFmt, err, errLastUpdateConflict)
	}
}

// TestBuildSleepParamsDateConflict rejects date + start.
func TestBuildSleepParamsDateConflict(t *testing.T) {
	t.Parallel()

	opts := sleepGetOptions{
		TimeRange: timeRangeOptions{
			Start: sleepTestRangeValue,
			End:   emptyString,
		},
		Date:       dateOption{Date: sleepTestDate},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
		Model:      defaultInt,
	}

	_, err := buildSleepParams(opts)
	if err == nil {
		t.Fatal(sleepTestExpectErr)
	}

	if !errors.Is(err, errDateRangeConflict) {
		t.Fatalf(sleepTestErrFmt, err, errDateRangeConflict)
	}
}

// TestBuildSleepParamsInvalidDate rejects invalid date.
func TestBuildSleepParamsInvalidDate(t *testing.T) {
	t.Parallel()

	opts := sleepGetOptions{
		TimeRange:  timeRangeOptions{Start: emptyString, End: emptyString},
		Date:       dateOption{Date: sleepTestDateBad},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
		Model:      defaultInt,
	}

	_, err := buildSleepParams(opts)
	if err == nil {
		t.Fatal(sleepTestExpectErr)
	}

	if !errors.Is(err, errInvalidDate) {
		t.Fatalf(sleepTestErrFmt, err, errInvalidDate)
	}
}
