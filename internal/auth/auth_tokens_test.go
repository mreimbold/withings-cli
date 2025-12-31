//nolint:testpackage // test unexported helpers.
package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mreimbold/withings-cli/internal/app"
)

const (
	testSourceProject       = "project"
	testSourceUser          = "user"
	testTokenProject        = "project-token"
	testTokenProjectRefresh = "project-refresh"
	testTokenUserAccess     = "user-access"
	testTokenUserRefresh    = "user-refresh"
	testTokenUser           = "usertoken"
	testGotWantFormat       = "got %q want %q"
	testExitErrFormat       = "expected exitError, got %T"
	testExitCodeFormat      = "exit code got %d want %d"
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
		AccessSource:  testSourceProject,
		RefreshToken:  emptyString,
		RefreshSource: emptyString,
		ExpiresAt:     future,
	}

	got := usableAccessToken(state)
	if got != emptyString {
		t.Fatalf(testGotWantFormat, got, emptyString)
	}
}

// TestUsableAccessTokenProject respects expiry for project tokens.
func TestUsableAccessTokenProject(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)
	valid := tokenState{
		AccessToken:   testTokenProject,
		AccessSource:  testSourceProject,
		RefreshToken:  emptyString,
		RefreshSource: emptyString,
		ExpiresAt:     future,
	}
	expired := tokenState{
		AccessToken:   testTokenProject,
		AccessSource:  testSourceProject,
		RefreshToken:  emptyString,
		RefreshSource: emptyString,
		ExpiresAt:     past,
	}

	if got := usableAccessToken(valid); got != testTokenProject {
		t.Fatalf(testGotWantFormat, got, testTokenProject)
	}

	if got := usableAccessToken(expired); got != emptyString {
		t.Fatalf(testGotWantFormat, got, emptyString)
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

// TestBuildTokenStatePrefersProject verifies project precedence.
func TestBuildTokenStatePrefersProject(t *testing.T) {
	t.Parallel()

	projectConfig := testConfigFile(map[string]string{
		configKeyAccessToken:  testTokenProject,
		configKeyRefreshToken: testTokenProjectRefresh,
	})
	userConfig := testConfigFile(map[string]string{
		configKeyAccessToken:  testTokenUserAccess,
		configKeyRefreshToken: testTokenUserRefresh,
		configKeyTokenExpiresAt: time.Now().
			Add(2 * time.Hour).
			UTC().
			Format(time.RFC3339),
	})

	state := buildTokenState(projectConfig, userConfig)
	if state.AccessToken != testTokenProject {
		t.Fatalf("access got %q want %q", state.AccessToken, testTokenProject)
	}

	if state.RefreshToken != testTokenProjectRefresh {
		t.Fatalf(
			"refresh got %q want %q",
			state.RefreshToken,
			testTokenProjectRefresh,
		)
	}

	if state.AccessSource != testSourceProject ||
		state.RefreshSource != testSourceProject {
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

// TestEnsureAccessTokenConfig returns stored tokens without refresh.
func TestEnsureAccessTokenConfig(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.toml")

	err := writeConfigFile(configPath, map[string]string{
		configKeyAccessToken: testTokenUser,
	})
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts := testAppOptions(configPath)

	token, err := EnsureAccessToken(context.Background(), opts)
	if err != nil {
		t.Fatalf("ensureAccessToken: %v", err)
	}

	if token != testTokenUser {
		t.Fatalf(testGotWantFormat, token, testTokenUser)
	}
}

// TestEnsureAccessTokenRequiresAuth fails without stored tokens.
func TestEnsureAccessTokenRequiresAuth(t *testing.T) {
	t.Parallel()

	opts := testAppOptions(filepath.Join(t.TempDir(), "config.toml"))

	_, err := EnsureAccessToken(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error")
	}

	var exitErr *app.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf(testExitErrFormat, err)
	}

	if exitErr.Code != app.ExitCodeAuth {
		t.Fatalf(testExitCodeFormat, exitErr.Code, app.ExitCodeAuth)
	}

	if !errors.Is(exitErr.Err, errAuthRequired) {
		t.Fatalf("expected errAuthRequired, got %v", exitErr.Err)
	}
}

// TestClassifyRefreshError maps network errors to network exits.
func TestClassifyRefreshError(t *testing.T) {
	t.Parallel()

	err := classifyRefreshError(networkError{err: errAuthRequired})

	var exitErr *app.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf(testExitErrFormat, err)
	}

	if exitErr.Code != app.ExitCodeNetwork {
		t.Fatalf(testExitCodeFormat, exitErr.Code, app.ExitCodeNetwork)
	}

	err = classifyRefreshError(errAuthRequired)
	if !errors.As(err, &exitErr) {
		t.Fatalf(testExitErrFormat, err)
	}

	if exitErr.Code != app.ExitCodeAuth {
		t.Fatalf(testExitCodeFormat, exitErr.Code, app.ExitCodeAuth)
	}
}

func testAppOptions(configPath string) app.Options {
	return app.Options{
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

func writeConfigFile(path string, values map[string]string) error {
	lines := make([]string, defaultInt, len(values))
	for key, value := range values {
		lines = append(lines, key+` = "`+value+`"`)
	}

	data := []byte(strings.Join(lines, "\n") + "\n")

	err := os.WriteFile(path, data, configFileMode)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
