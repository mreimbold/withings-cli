package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

type dateRange struct {
	Start string
	End   string
}

var (
	errInvalidDate       = errors.New("invalid --date (expected YYYY-MM-DD)")
	errDateRangeConflict = errors.New(
		"--date cannot be combined with --start or --end",
	)
)

func parseDateValue(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == emptyString {
		return emptyString, errInvalidDate
	}

	parsed, err := time.Parse(dateLayout, trimmed)
	if err != nil {
		return emptyString, errInvalidDate
	}

	return parsed.Format(dateLayout), nil
}

func dateFromTimeValue(raw string, errInvalid error) (string, error) {
	if raw == emptyString {
		return emptyString, nil
	}

	epoch, err := parseEpoch(raw)
	if err != nil {
		return emptyString, fmt.Errorf("%w: %w", errInvalid, err)
	}

	return time.Unix(epoch, defaultInt64).UTC().Format(dateLayout), nil
}

func resolveDateRange(
	date dateOption,
	timeRange timeRangeOptions,
	errStart error,
	errEnd error,
) (dateRange, error) {
	if date.Date != emptyString {
		return resolveDateRangeFromDate(date, timeRange)
	}

	startDate, err := dateFromTimeValue(timeRange.Start, errStart)
	if err != nil {
		return dateRange{}, err
	}

	endDate, err := dateFromTimeValue(timeRange.End, errEnd)
	if err != nil {
		return dateRange{}, err
	}

	return dateRange{Start: startDate, End: endDate}, nil
}

func resolveDateRangeFromDate(
	date dateOption,
	timeRange timeRangeOptions,
) (dateRange, error) {
	if hasTimeRange(timeRange) {
		return dateRange{}, errDateRangeConflict
	}

	parsed, err := parseDateValue(date.Date)
	if err != nil {
		return dateRange{}, err
	}

	return dateRange{Start: parsed, End: parsed}, nil
}

func hasTimeRange(timeRange timeRangeOptions) bool {
	return timeRange.Start != emptyString || timeRange.End != emptyString
}

func applyLastUpdateFilter(
	values *url.Values,
	param string,
	lastUpdate lastUpdateOption,
	date dateOption,
	timeRange timeRangeOptions,
) error {
	hasRange := date.Date != emptyString || hasTimeRange(timeRange)

	if lastUpdate.LastUpdate < defaultInt64 {
		return errInvalidLastUpdate
	}

	if lastUpdate.LastUpdate > defaultInt64 && hasRange {
		return errLastUpdateConflict
	}

	if lastUpdate.LastUpdate > defaultInt64 {
		values.Set(
			param,
			strconv.FormatInt(lastUpdate.LastUpdate, measureNumberBase10),
		)
	}

	return nil
}

func applyDateRangeParams(
	values *url.Values,
	startKey string,
	endKey string,
	rangeValues dateRange,
) {
	if rangeValues.Start != emptyString {
		values.Set(startKey, rangeValues.Start)
	}

	if rangeValues.End != emptyString {
		values.Set(endKey, rangeValues.End)
	}
}
