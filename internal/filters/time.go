// Package filters provides shared filter parsing helpers.
package filters

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mreimbold/withings-cli/internal/errs"
	"github.com/mreimbold/withings-cli/internal/params"
)

const (
	dateLayout   = "2006-01-02"
	numberBase10 = 10
	epochBitSize = 64
	defaultInt64 = 0
	emptyString  = ""
)

// DateRange represents resolved start/end dates.
type DateRange struct {
	Start string
	End   string
}

// ParseDateValue parses a YYYY-MM-DD value into a normalized date string.
func ParseDateValue(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == emptyString {
		return emptyString, errs.ErrInvalidDate
	}

	parsed, err := time.Parse(dateLayout, trimmed)
	if err != nil {
		return emptyString, errs.ErrInvalidDate
	}

	return parsed.Format(dateLayout), nil
}

// DateFromTimeValue resolves a start/end time into a date string.
func DateFromTimeValue(raw string, errInvalid error) (string, error) {
	if raw == emptyString {
		return emptyString, nil
	}

	epoch, err := ParseEpoch(raw)
	if err != nil {
		return emptyString, fmt.Errorf("%w: %w", errInvalid, err)
	}

	return time.Unix(epoch, defaultInt64).UTC().Format(dateLayout), nil
}

// ResolveDateRange derives a DateRange from date or time-range filters.
func ResolveDateRange(
	date params.Date,
	timeRange params.TimeRange,
	errStart error,
	errEnd error,
) (DateRange, error) {
	if date.Date != emptyString {
		return resolveDateRangeFromDate(date, timeRange)
	}

	startDate, err := DateFromTimeValue(timeRange.Start, errStart)
	if err != nil {
		return DateRange{}, err
	}

	endDate, err := DateFromTimeValue(timeRange.End, errEnd)
	if err != nil {
		return DateRange{}, err
	}

	return DateRange{Start: startDate, End: endDate}, nil
}

func resolveDateRangeFromDate(
	date params.Date,
	timeRange params.TimeRange,
) (DateRange, error) {
	if HasTimeRange(timeRange) {
		return DateRange{}, errs.ErrDateRangeConflict
	}

	parsed, err := ParseDateValue(date.Date)
	if err != nil {
		return DateRange{}, err
	}

	return DateRange{Start: parsed, End: parsed}, nil
}

// HasTimeRange reports whether any time-range values are set.
func HasTimeRange(timeRange params.TimeRange) bool {
	return timeRange.Start != emptyString || timeRange.End != emptyString
}

// ParseEpoch parses RFC3339, YYYY-MM-DD, or epoch timestamp strings.
func ParseEpoch(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == emptyString {
		return defaultInt64, errs.ErrEmptyTimeValue
	}

	epoch, err := strconv.ParseInt(trimmed, numberBase10, epochBitSize)
	if err == nil {
		return epoch, nil
	}

	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		parsedDate, dateErr := time.Parse(dateLayout, trimmed)
		if dateErr != nil {
			return defaultInt64, errs.ErrInvalidTimeFormat
		}

		return parsedDate.Unix(), nil
	}

	return parsed.Unix(), nil
}

// ApplyLastUpdateFilter enforces last-update and range conflicts.
func ApplyLastUpdateFilter(
	values *url.Values,
	param string,
	lastUpdate params.LastUpdate,
	date params.Date,
	timeRange params.TimeRange,
	errInvalid error,
	errConflict error,
) error {
	hasRange := date.Date != emptyString || HasTimeRange(timeRange)

	if lastUpdate.LastUpdate < defaultInt64 {
		return errInvalid
	}

	if lastUpdate.LastUpdate > defaultInt64 && hasRange {
		return errConflict
	}

	if lastUpdate.LastUpdate > defaultInt64 {
		values.Set(
			param,
			strconv.FormatInt(lastUpdate.LastUpdate, numberBase10),
		)
	}

	return nil
}

// ApplyDateRangeParams sets date range parameters when present.
func ApplyDateRangeParams(
	values *url.Values,
	startKey string,
	endKey string,
	rangeValues DateRange,
) {
	if rangeValues.Start != emptyString {
		values.Set(startKey, rangeValues.Start)
	}

	if rangeValues.End != emptyString {
		values.Set(endKey, rangeValues.End)
	}
}
