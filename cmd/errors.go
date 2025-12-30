package cmd

type staticError string

// Error returns the static error string.
func (e staticError) Error() string {
	return string(e)
}

const (
	errJSONPlainConflict staticError = "--json and --plain are " +
		"mutually exclusive"
	errQuietVerboseConflict staticError = "--quiet and --verbose cannot be " +
		"combined"
	errInvalidCloud   staticError = "invalid --cloud (expected eu or us)"
	errNotImplemented staticError = "not implemented"
)
