//nolint:testpackage // test unexported helpers.
package heart

import (
	"errors"
	"testing"

	"github.com/mreimbold/withings-cli/internal/errs"
	"github.com/mreimbold/withings-cli/internal/params"
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
	testEmptyString   = ""
	testDefaultInt    = 0
	testDefaultInt64  = 0
)

// TestHeartServiceForBase handles base URLs with and without /v2.
func TestHeartServiceForBase(t *testing.T) {
	t.Parallel()

	if got := serviceForBase(testBaseNoV2); got != serviceName {
		t.Fatalf(testServiceFmt, got, serviceName)
	}

	if got := serviceForBase(testBaseV2); got != serviceShort {
		t.Fatalf(testServiceFmt, got, serviceShort)
	}

	if got := serviceForBase(testBaseV2Slash); got != serviceShort {
		t.Fatalf(testServiceFmt, got, serviceShort)
	}
}

// TestBuildParams ensures standard heart query params are built.
func TestBuildParams(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: testStartEpochStr,
			End:   testEndEpochStr,
		},
		Pagination: params.Pagination{
			Limit:  testLimit,
			Offset: testOffset,
		},
		User:       params.User{UserID: testUserID},
		LastUpdate: params.LastUpdate{LastUpdate: testDefaultInt64},
		Signal:     true,
	}

	values, err := buildParams(opts)
	if err != nil {
		t.Fatalf("buildParams: %v", err)
	}

	assertParam(t, values.Get(startDateParam), testStartEpochStr, "startdate")
	assertParam(t, values.Get(endDateParam), testEndEpochStr, "enddate")
	assertParam(t, values.Get(limitParam), "50", "limit")
	assertParam(t, values.Get(offsetParam), "10", "offset")
	assertParam(t, values.Get(userIDParam), testUserID, "userid")
	assertParam(t, values.Get(signalParam), signalEnabled, "signal")
}

// TestBuildParamsNoSignal ensures signal param is omitted when disabled.
func TestBuildParamsNoSignal(t *testing.T) {
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
		Signal:     false,
	}

	values, err := buildParams(opts)
	if err != nil {
		t.Fatalf("buildParams: %v", err)
	}

	if values.Get(signalParam) != testEmptyString {
		t.Fatalf("signal got %q want empty", values.Get(signalParam))
	}
}

// TestBuildParamsLastUpdateConflict rejects mixing last-update and range.
func TestBuildParamsLastUpdateConflict(t *testing.T) {
	t.Parallel()

	opts := Options{
		TimeRange: params.TimeRange{
			Start: testStartEpochStr,
			End:   testEmptyString,
		},
		Pagination: params.Pagination{
			Limit:  testDefaultInt,
			Offset: testDefaultInt,
		},
		User:       params.User{UserID: testEmptyString},
		LastUpdate: params.LastUpdate{LastUpdate: testLastUpdate},
		Signal:     false,
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errs.ErrLastUpdateConflict) {
		t.Fatalf("err got %v want %v", err, errs.ErrLastUpdateConflict)
	}
}

// TestBuildParamsLastUpdateInvalid rejects negative last-update.
func TestBuildParamsLastUpdateInvalid(t *testing.T) {
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
		LastUpdate: params.LastUpdate{LastUpdate: testLastInvalid},
		Signal:     false,
	}

	_, err := buildParams(opts)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errs.ErrInvalidLastUpdate) {
		t.Fatalf("err got %v want %v", err, errs.ErrInvalidLastUpdate)
	}
}

// TestSeriesTimestampPreference chooses the best available timestamp.
func TestSeriesTimestampPreference(t *testing.T) {
	t.Parallel()

	seriesValue := series{
		ID:        testDefaultInt64,
		SignalID:  testDefaultInt64,
		StartDate: testSeriesStart,
		EndDate:   testSeriesEnd,
		Timestamp: testSeriesStamp,
		DeviceID:  testEmptyString,
		Model:     testDefaultInt,
		ECG:       testDefaultInt,
		AFib:      testDefaultInt,
		HeartRate: testDefaultInt,
		Signal:    nil,
	}

	if got := seriesTimestamp(seriesValue); got != testSeriesStart {
		t.Fatalf(testTimestampFmt, got, testSeriesStart)
	}

	seriesValue = series{
		ID:        testDefaultInt64,
		SignalID:  testDefaultInt64,
		StartDate: testDefaultInt64,
		EndDate:   testSeriesEnd,
		Timestamp: testSeriesStamp,
		DeviceID:  testEmptyString,
		Model:     testDefaultInt,
		ECG:       testDefaultInt,
		AFib:      testDefaultInt,
		HeartRate: testDefaultInt,
		Signal:    nil,
	}
	if got := seriesTimestamp(seriesValue); got != testSeriesStamp {
		t.Fatalf(testTimestampFmt, got, testSeriesStamp)
	}

	seriesValue = series{
		ID:        testDefaultInt64,
		SignalID:  testDefaultInt64,
		StartDate: testDefaultInt64,
		EndDate:   testSeriesEnd,
		Timestamp: testDefaultInt64,
		DeviceID:  testEmptyString,
		Model:     testDefaultInt,
		ECG:       testDefaultInt,
		AFib:      testDefaultInt,
		HeartRate: testDefaultInt,
		Signal:    nil,
	}
	if got := seriesTimestamp(seriesValue); got != testSeriesEnd {
		t.Fatalf(testTimestampFmt, got, testSeriesEnd)
	}
}

// TestSeriesSignalIDFallback uses ID when signal ID is missing.
func TestSeriesSignalIDFallback(t *testing.T) {
	t.Parallel()

	seriesValue := series{
		ID:        testSeriesID,
		SignalID:  testDefaultInt64,
		StartDate: testDefaultInt64,
		EndDate:   testDefaultInt64,
		Timestamp: testDefaultInt64,
		DeviceID:  testEmptyString,
		Model:     testDefaultInt,
		ECG:       testDefaultInt,
		AFib:      testDefaultInt,
		HeartRate: testDefaultInt,
		Signal:    nil,
	}
	if got := seriesSignalID(seriesValue); got != testSeriesID {
		t.Fatalf(testSignalIDFmt, got, testSeriesID)
	}

	seriesValue = series{
		ID:        testSeriesID,
		SignalID:  testSignalID,
		StartDate: testDefaultInt64,
		EndDate:   testDefaultInt64,
		Timestamp: testDefaultInt64,
		DeviceID:  testEmptyString,
		Model:     testDefaultInt,
		ECG:       testDefaultInt,
		AFib:      testDefaultInt,
		HeartRate: testDefaultInt,
		Signal:    nil,
	}
	if got := seriesSignalID(seriesValue); got != testSignalID {
		t.Fatalf(testSignalIDFmt, got, testSignalID)
	}
}

func assertParam(t *testing.T, got, want, name string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s got %q want %q", name, got, want)
	}
}
