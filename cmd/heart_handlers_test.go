//nolint:testpackage // test unexported helpers in cmd.
package cmd

import (
	"errors"
	"testing"
)

const (
	testStartEpochStr = "1700000000"
	testEndEpochStr   = "1700003600"
	testLimit         = 50
	testOffset        = 10
	testUserID        = "user-123"
	testBaseV2        = "https://wbsapi.withings.net/v2"
	testBaseV2Slash   = "https://wbsapi.withings.net/v2/"
	testBaseNoV2      = "https://wbsapi.withings.net"
	testLastUpdate    = 100
	testLastInvalid   = -1
	testSeriesStart   = 111
	testSeriesStamp   = 222
	testSeriesEnd     = 333
	testSeriesID      = 42
	testSignalID      = 99
	testTimestampFmt  = "timestamp got %d want %d"
	testSignalIDFmt   = "signal id got %d want %d"
	testServiceFmt    = "service got %q want %q"
)

// TestHeartServiceForBase handles base URLs with and without /v2.
func TestHeartServiceForBase(t *testing.T) {
	t.Parallel()

	if got := heartServiceForBase(testBaseNoV2); got != heartServiceName {
		t.Fatalf(testServiceFmt, got, heartServiceName)
	}

	if got := heartServiceForBase(testBaseV2); got != heartServiceShort {
		t.Fatalf(testServiceFmt, got, heartServiceShort)
	}

	if got := heartServiceForBase(testBaseV2Slash); got != heartServiceShort {
		t.Fatalf(testServiceFmt, got, heartServiceShort)
	}
}

// TestBuildHeartParams ensures standard heart query params are built.
func TestBuildHeartParams(t *testing.T) {
	t.Parallel()

	opts := heartGetOptions{
		TimeRange: timeRangeOptions{
			Start: testStartEpochStr,
			End:   testEndEpochStr,
		},
		Pagination: paginationOptions{
			Limit:  testLimit,
			Offset: testOffset,
		},
		User:       userOption{UserID: testUserID},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
		Signal:     true,
	}

	values, err := buildHeartParams(opts)
	if err != nil {
		t.Fatalf("buildHeartParams: %v", err)
	}

	assertParam(
		t,
		values.Get(heartStartDateParam),
		testStartEpochStr,
		"startdate",
	)
	assertParam(t, values.Get(heartEndDateParam), testEndEpochStr, "enddate")
	assertParam(t, values.Get(heartLimitParam), "50", "limit")
	assertParam(t, values.Get(heartOffsetParam), "10", "offset")
	assertParam(t, values.Get(heartUserIDParam), testUserID, "userid")
	assertParam(t, values.Get(heartSignalParam), heartSignalEnabled, "signal")
}

// TestBuildHeartParamsNoSignal ensures signal param is omitted when disabled.
func TestBuildHeartParamsNoSignal(t *testing.T) {
	t.Parallel()

	opts := heartGetOptions{
		TimeRange:  timeRangeOptions{Start: emptyString, End: emptyString},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: defaultInt64},
		Signal:     false,
	}

	values, err := buildHeartParams(opts)
	if err != nil {
		t.Fatalf("buildHeartParams: %v", err)
	}

	if values.Get(heartSignalParam) != emptyString {
		t.Fatalf("signal got %q want empty", values.Get(heartSignalParam))
	}
}

// TestBuildHeartParamsLastUpdateConflict rejects mixing last-update and range.
func TestBuildHeartParamsLastUpdateConflict(t *testing.T) {
	t.Parallel()

	opts := heartGetOptions{
		TimeRange: timeRangeOptions{
			Start: testStartEpochStr,
			End:   emptyString,
		},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: testLastUpdate},
		Signal:     false,
	}

	_, err := buildHeartParams(opts)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errLastUpdateConflict) {
		t.Fatalf("err got %v want %v", err, errLastUpdateConflict)
	}
}

// TestBuildHeartParamsLastUpdateInvalid rejects negative last-update.
func TestBuildHeartParamsLastUpdateInvalid(t *testing.T) {
	t.Parallel()

	opts := heartGetOptions{
		TimeRange:  timeRangeOptions{Start: emptyString, End: emptyString},
		Pagination: paginationOptions{Limit: defaultInt, Offset: defaultInt},
		User:       userOption{UserID: emptyString},
		LastUpdate: lastUpdateOption{LastUpdate: testLastInvalid},
		Signal:     false,
	}

	_, err := buildHeartParams(opts)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errInvalidLastUpdate) {
		t.Fatalf("err got %v want %v", err, errInvalidLastUpdate)
	}
}

// TestHeartSeriesTimestampPreference chooses the best available timestamp.
func TestHeartSeriesTimestampPreference(t *testing.T) {
	t.Parallel()

	series := heartSeries{
		ID:        defaultInt64,
		SignalID:  defaultInt64,
		StartDate: testSeriesStart,
		EndDate:   testSeriesEnd,
		Timestamp: testSeriesStamp,
		DeviceID:  emptyString,
		Model:     defaultInt,
		ECG:       defaultInt,
		AFib:      defaultInt,
		HeartRate: defaultInt,
		Signal:    nil,
	}

	if got := heartSeriesTimestamp(series); got != testSeriesStart {
		t.Fatalf(testTimestampFmt, got, testSeriesStart)
	}

	series = heartSeries{
		ID:        defaultInt64,
		SignalID:  defaultInt64,
		StartDate: defaultInt64,
		EndDate:   testSeriesEnd,
		Timestamp: testSeriesStamp,
		DeviceID:  emptyString,
		Model:     defaultInt,
		ECG:       defaultInt,
		AFib:      defaultInt,
		HeartRate: defaultInt,
		Signal:    nil,
	}
	if got := heartSeriesTimestamp(series); got != testSeriesStamp {
		t.Fatalf(testTimestampFmt, got, testSeriesStamp)
	}

	series = heartSeries{
		ID:        defaultInt64,
		SignalID:  defaultInt64,
		StartDate: defaultInt64,
		EndDate:   testSeriesEnd,
		Timestamp: defaultInt64,
		DeviceID:  emptyString,
		Model:     defaultInt,
		ECG:       defaultInt,
		AFib:      defaultInt,
		HeartRate: defaultInt,
		Signal:    nil,
	}
	if got := heartSeriesTimestamp(series); got != testSeriesEnd {
		t.Fatalf(testTimestampFmt, got, testSeriesEnd)
	}
}

// TestHeartSeriesSignalIDFallback uses ID when signal ID is missing.
func TestHeartSeriesSignalIDFallback(t *testing.T) {
	t.Parallel()

	series := heartSeries{
		ID:        testSeriesID,
		SignalID:  defaultInt64,
		StartDate: defaultInt64,
		EndDate:   defaultInt64,
		Timestamp: defaultInt64,
		DeviceID:  emptyString,
		Model:     defaultInt,
		ECG:       defaultInt,
		AFib:      defaultInt,
		HeartRate: defaultInt,
		Signal:    nil,
	}
	if got := heartSeriesSignalID(series); got != testSeriesID {
		t.Fatalf(testSignalIDFmt, got, testSeriesID)
	}

	series = heartSeries{
		ID:        testSeriesID,
		SignalID:  testSignalID,
		StartDate: defaultInt64,
		EndDate:   defaultInt64,
		Timestamp: defaultInt64,
		DeviceID:  emptyString,
		Model:     defaultInt,
		ECG:       defaultInt,
		AFib:      defaultInt,
		HeartRate: defaultInt,
		Signal:    nil,
	}
	if got := heartSeriesSignalID(series); got != testSignalID {
		t.Fatalf(testSignalIDFmt, got, testSignalID)
	}
}

func assertParam(t *testing.T, got, want, name string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s got %q want %q", name, got, want)
	}
}
