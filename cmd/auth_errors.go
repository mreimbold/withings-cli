package cmd

import "errors"

var (
	errAuthCodeRequired         = errors.New("authorization code required")
	errAuthTimedOut             = errors.New("authorization timed out")
	errAuthorizationFailed      = errors.New("authorization failed")
	errClientIDRequired         = errors.New("client ID required")
	errClientIDSecretRequired   = errors.New("client ID and secret required")
	errClientSecretRequired     = errors.New("client secret required")
	errClientCredentialsMissing = errors.New("missing client ID or secret")
	errClientConfigMissing      = errors.New(
		"client ID, secret, and redirect URI required",
	)
	errInputRequired = errors.New(
		"input required but prompting disabled",
	)
	errMissingAuthCode      = errors.New("missing code")
	errMissingClientID      = errors.New("missing client ID")
	errMissingRedirectURI   = errors.New("missing redirect URI")
	errInvalidOpenMode      = errors.New("invalid open mode")
	errRefreshTokenRequired = errors.New("refresh token required")
	errStateMismatch        = errors.New("state mismatch")
	errTokenRequestFailed   = errors.New("token request failed")
	errWithingsAPI          = errors.New("withings API error")
)
