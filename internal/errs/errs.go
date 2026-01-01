// Package errs defines shared CLI validation errors.
package errs

import "errors"

var (
	// ErrInvalidStartTime indicates an invalid start time argument.
	ErrInvalidStartTime = errors.New(
		"invalid --start (expected RFC3339, YYYY-MM-DD, or epoch)",
	)
	// ErrInvalidEndTime indicates an invalid end time argument.
	ErrInvalidEndTime = errors.New(
		"invalid --end (expected RFC3339, YYYY-MM-DD, or epoch)",
	)
	// ErrInvalidLastUpdate indicates an invalid last-update argument.
	ErrInvalidLastUpdate = errors.New(
		"invalid --last-update (expected epoch)",
	)
	// ErrLastUpdateConflict indicates last-update used with date range flags.
	ErrLastUpdateConflict = errors.New(
		"--last-update cannot be combined with --start or --end",
	)
	// ErrInvalidDate indicates an invalid date argument.
	ErrInvalidDate = errors.New("invalid --date (expected YYYY-MM-DD)")
	// ErrInvalidTimeFormat indicates a time parse failure.
	ErrInvalidTimeFormat = errors.New("expected RFC3339 or YYYY-MM-DD")
	// ErrDateRangeConflict indicates --date used with --start or --end.
	ErrDateRangeConflict = errors.New(
		"--date cannot be combined with --start or --end",
	)
	// ErrEmptyTimeValue indicates a required time value is empty.
	ErrEmptyTimeValue = errors.New("empty time value")
)
