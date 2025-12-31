package withings

import "errors"

// ErrAPI indicates a non-success response from the Withings API.
var ErrAPI = errors.New("withings API error")
