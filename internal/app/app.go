// Package app provides shared CLI options and exit metadata.
package app

// Options holds global CLI settings.
type Options struct {
	Verbose int
	Quiet   bool
	JSON    bool
	Plain   bool
	NoColor bool
	NoInput bool
	Config  string
	Cloud   string
	BaseURL string
}

const (
	// ExitCodeSuccess indicates a successful run.
	ExitCodeSuccess = 0
	// ExitCodeFailure indicates an internal failure.
	ExitCodeFailure = 1
	// ExitCodeUsage indicates invalid CLI usage.
	ExitCodeUsage = 2
	// ExitCodeAuth indicates authentication is required or failed.
	ExitCodeAuth = 3
	// ExitCodeNetwork indicates a network failure.
	ExitCodeNetwork = 4
	// ExitCodeAPI indicates an upstream API error.
	ExitCodeAPI = 5
)

// ExitError couples an exit code with an error.
type ExitError struct {
	Code int
	Err  error
}

// NewExitError builds an ExitError with a code and cause.
func NewExitError(code int, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

// Error returns the wrapped error message.
func (e *ExitError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *ExitError) Unwrap() error {
	return e.Err
}
