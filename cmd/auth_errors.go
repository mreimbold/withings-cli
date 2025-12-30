package cmd

import "errors"

var (
	errAuthTimedOut             = errors.New("authorization timed out")
	errAuthorizationFailed      = errors.New("authorization failed")
	errAuthRequired             = errors.New("authentication required")
	errClientCredentialsMissing = errors.New("missing client ID or secret")
	errInputRequired            = errors.New(
		"input required but prompting disabled",
	)
	errMissingAuthCode    = errors.New("missing code")
	errInvalidOpenMode    = errors.New("invalid open mode")
	errStateMismatch      = errors.New("state mismatch")
	errTokenRequestFailed = errors.New("token request failed")
	errWithingsAPI        = errors.New("withings API error")
)
