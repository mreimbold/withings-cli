// Package params defines shared CLI parameter structs.
package params

// TimeRange captures start/end time filters.
type TimeRange struct {
	Start string
	End   string
}

// Date captures a single date filter.
type Date struct {
	Date string
}

// Pagination captures limit/offset paging.
type Pagination struct {
	Limit  int
	Offset int
}

// User captures a Withings user ID.
type User struct {
	UserID string
}

// LastUpdate captures a last-update epoch filter.
type LastUpdate struct {
	LastUpdate int64
}
