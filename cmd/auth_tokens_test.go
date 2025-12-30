//nolint:testpackage // test unexported helpers in cmd.
package cmd

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

const (
	testSourceEnv       = "env"
	testSourceUser      = "user"
	testTokenEnv        = "env-token"
	testTokenEnvAccess  = "env-access"
	testTokenEnvRefresh = "env-refresh"
	testTokenUser       = "usertoken"
	testGotWantFormat   = "got %q want %q"
	testExitErrFormat   = "expected exitError, got %T"
	testExitCodeFormat  = "exit code got %d want %d"
)

// TestShouldRefresh validates expiry decisions.
func TestShouldRefresh(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(1 * time.Hour)
	if shouldRefresh(future) {
		t.Fatal("expected future expiry to not refresh")
	}

	past := time.Now().Add(-1 * time.Hour)
	if !shouldRefresh(past) {
		t.Fatal("expected past expiry to refresh")
	}

	if shouldRefresh(time.Time{}) {
		t.Fatal("expected zero expiry to not refresh")
	}
}

// TestUsableAccessTokenEmpty returns empty when no token is set.
func TestUsableAccessTokenEmpty(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(1 * time.Hour)
	state := tokenState{
		AccessToken:   emptyString,
		AccessSource:  testSourceEnv,
		RefreshToken:  emptyString,
		RefreshSource: emptyString,
		ExpiresAt:     future,
	}

	got := usableAccessToken(state)
	if got != emptyString {
		t.Fatalf(testGotWantFormat, got, emptyString)
	}
}

// TestUsableAccessTokenEnv keeps env tokens even if expired.
func TestUsableAccessTokenEnv(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-1 * time.Hour)
	state := tokenState{
		AccessToken:   testTokenEnv,
		AccessSource:  testSourceEnv,
		RefreshToken:  emptyString,
		RefreshSource: emptyString,
		ExpiresAt:     past,
	}

	got := usableAccessToken(state)
	if got != testTokenEnv {
		t.Fatalf(testGotWantFormat, got, testTokenEnv)
	}
}

// TestUsableAccessTokenUser respects expiry for user tokens.
func TestUsableAccessTokenUser(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)
	valid := tokenState{
		AccessToken:   testTokenUser,
		AccessSource:  testSourceUser,
		RefreshToken:  emptyString,
		RefreshSource: emptyString,
		ExpiresAt:     future,
	}
	expired := tokenState{
		AccessToken:   testTokenUser,
		AccessSource:  testSourceUser,
		RefreshToken:  emptyString,
		RefreshSource: emptyString,
		ExpiresAt:     past,
	}

	if got := usableAccessToken(valid); got != testTokenUser {
		t.Fatalf(testGotWantFormat, got, testTokenUser)
	}

	if got := usableAccessToken(expired); got != emptyString {
		t.Fatalf(testGotWantFormat, got, emptyString)
	}
}

// TestBuildTokenStatePrefersEnv verifies env precedence.
func TestBuildTokenStatePrefersEnv(t *testing.T) {
	t.Setenv(envAccessToken, testTokenEnvAccess)
	t.Setenv(envRefreshToken, testTokenEnvRefresh)

	projectConfig := testConfigFile(map[string]string{
		configKeyAccessToken:  "project-access",
		configKeyRefreshToken: "project-refresh",
	})
	userConfig := testConfigFile(map[string]string{
		configKeyAccessToken:  "user-access",
		configKeyRefreshToken: "user-refresh",
		configKeyTokenExpiresAt: time.Now().
			Add(2 * time.Hour).
			UTC().
			Format(time.RFC3339),
	})

	state := buildTokenState(projectConfig, userConfig)
	if state.AccessToken != testTokenEnvAccess {
		t.Fatalf("access got %q want %q", state.AccessToken, testTokenEnvAccess)
	}

	if state.RefreshToken != testTokenEnvRefresh {
		t.Fatalf(
			"refresh got %q want %q",
			state.RefreshToken,
			testTokenEnvRefresh,
		)
	}

	if state.AccessSource != testSourceEnv ||
		state.RefreshSource != testSourceEnv {
		t.Fatalf(
			"unexpected sources: %s/%s",
			state.AccessSource,
			state.RefreshSource,
		)
	}

	if state.ExpiresAt.IsZero() {
		t.Fatal("expected expiresAt to be parsed")
	}
}

// TestEnsureAccessTokenEnv returns env tokens without refresh.
func TestEnsureAccessTokenEnv(t *testing.T) {
	t.Setenv(envAccessToken, testTokenEnv)

	opts := testGlobalOptions(filepath.Join(t.TempDir(), "config.toml"))

	token, err := ensureAccessToken(context.Background(), opts)
	if err != nil {
		t.Fatalf("ensureAccessToken: %v", err)
	}

	if token != testTokenEnv {
		t.Fatalf(testGotWantFormat, token, testTokenEnv)
	}
}

// TestEnsureAccessTokenRequiresAuth fails without stored tokens.
//
//nolint:paralleltest // Reads env and must not race with Setenv tests.
func TestEnsureAccessTokenRequiresAuth(t *testing.T) {
	opts := testGlobalOptions(filepath.Join(t.TempDir(), "config.toml"))

	_, err := ensureAccessToken(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error")
	}

	var exitErr *exitError
	if !errors.As(err, &exitErr) {
		t.Fatalf(testExitErrFormat, err)
	}

	if exitErr.code != exitCodeAuth {
		t.Fatalf(testExitCodeFormat, exitErr.code, exitCodeAuth)
	}

	if !errors.Is(exitErr.err, errAuthRequired) {
		t.Fatalf("expected errAuthRequired, got %v", exitErr.err)
	}
}

// TestClassifyRefreshError maps network errors to network exits.
func TestClassifyRefreshError(t *testing.T) {
	t.Parallel()

	err := classifyRefreshError(networkError{err: errAuthRequired})

	var exitErr *exitError
	if !errors.As(err, &exitErr) {
		t.Fatalf(testExitErrFormat, err)
	}

	if exitErr.code != exitCodeNetwork {
		t.Fatalf(testExitCodeFormat, exitErr.code, exitCodeNetwork)
	}

	err = classifyRefreshError(errAuthRequired)
	if !errors.As(err, &exitErr) {
		t.Fatalf(testExitErrFormat, err)
	}

	if exitErr.code != exitCodeAuth {
		t.Fatalf(testExitCodeFormat, exitErr.code, exitCodeAuth)
	}
}

func testGlobalOptions(configPath string) globalOptions {
	return globalOptions{
		Verbose: defaultInt,
		Quiet:   false,
		JSON:    false,
		Plain:   false,
		NoColor: false,
		NoInput: false,
		Config:  configPath,
		Cloud:   emptyString,
		BaseURL: emptyString,
	}
}

func testConfigFile(values map[string]string) *configFile {
	return &configFile{
		Path:     emptyString,
		Lines:    nil,
		Values:   values,
		KeyIndex: map[string]int{},
		Exists:   false,
	}
}
