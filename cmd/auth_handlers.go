package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

const (
	authCallbackTimeout   = 2 * time.Minute
	authChannelBufferSize = 1
	authReadHeaderTimeout = 5 * time.Second
	authShutdownTimeout   = 5 * time.Second
	authStateSizeBytes    = 16
	authNumberBase10      = 10
)

type authOpenMode int

const (
	authOpenBrowser authOpenMode = iota
	authPrintURL
)

type authClientConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func runAuthLogin(cmd *cobra.Command, opts authLoginOptions) error {
	globalOpts, err := readGlobalOptions(cmd)
	if err != nil {
		return err
	}

	sources, err := loadConfigSources(globalOpts.Config)
	if err != nil {
		return err
	}

	userConfig := sources.User

	authConfig := resolveAuthConfig(opts.RedirectURI)

	err = requireClientCredentials(authConfig, errClientCredentialsMissing)
	if err != nil {
		return err
	}

	if authConfig.RedirectURI == emptyString {
		authConfig.RedirectURI = buildLocalRedirectURI(opts.Listen)
	}

	return executeAuthLogin(
		cmd.Context(),
		globalOpts,
		opts,
		authConfig,
		userConfig,
	)
}

func executeAuthLogin(
	ctx context.Context,
	globalOpts globalOptions,
	opts authLoginOptions,
	authConfig authClientConfig,
	userConfig *configFile,
) error {
	state := randomState()

	authorizeURL, err := buildAuthorizeURL(
		accountBaseURL(globalOpts.Cloud),
		authConfig.ClientID,
		authConfig.RedirectURI,
		emptyString,
		state,
	)
	if err != nil {
		return err
	}

	openMode := authOpenBrowser
	if opts.NoOpen {
		openMode = authPrintURL
	}

	code, err := waitForAuthCode(
		ctx,
		authConfig.RedirectURI,
		opts.Listen,
		state,
		authorizeURL,
		openMode,
	)
	if err != nil {
		return err
	}

	apiURL := apiBaseURL(globalOpts.BaseURL, globalOpts.Cloud)
	tokenURL := tokenEndpoint(apiURL)

	token, err := exchangeToken(
		ctx,
		tokenURL,
		authConfig.ClientID,
		authConfig.ClientSecret,
		code,
		authConfig.RedirectURI,
	)
	if err != nil {
		return classifyTokenError(err)
	}

	err = persistTokens(userConfig, token)
	if err != nil {
		return err
	}

	return writeOutput(globalOpts, "Authentication successful. Tokens saved.")
}

func runAuthStatus(cmd *cobra.Command, _ []string) error {
	globalOpts, err := readGlobalOptions(cmd)
	if err != nil {
		return err
	}

	sources, err := loadConfigSources(globalOpts.Config)
	if err != nil {
		return err
	}

	projectConfig := sources.Project
	userConfig := sources.User

	status := buildAuthStatus(projectConfig, userConfig)

	if globalOpts.JSON {
		return writeOutput(globalOpts, status.toMap())
	}

	return writeOutput(globalOpts, status.toLines())
}

func runAuthLogout(cmd *cobra.Command, opts authLogoutOptions) error {
	globalOpts, err := readGlobalOptions(cmd)
	if err != nil {
		return err
	}

	sources, err := loadConfigSources(globalOpts.Config)
	if err != nil {
		return err
	}

	userConfig := sources.User

	proceed, err := confirmLogout(opts, globalOpts)
	if err != nil {
		return err
	}

	if !proceed {
		return nil
	}

	removeTokenKeys(userConfig)

	err = userConfig.Save()
	if err != nil {
		return err
	}

	return writeOutput(globalOpts, "Tokens removed.")
}

func waitForAuthCode(
	ctx context.Context,
	redirectURI string,
	listenAddr string,
	state string,
	authorizeURL string,
	openMode authOpenMode,
) (string, error) {
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		return emptyString, fmt.Errorf("invalid redirect URI: %w", err)
	}

	path := parsed.Path
	if path == emptyString {
		path = "/"
	}

	server := startAuthServer(listenAddr, path, state)

	err = handleAuthOpen(ctx, openMode, authorizeURL)
	if err != nil {
		shutdownErr := shutdownAuthServer(ctx, server.server)

		return emptyString, errors.Join(err, shutdownErr)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, authCallbackTimeout)
	defer cancel()

	code, err := awaitAuthCode(timeoutCtx, server.codeCh, server.errCh)
	shutdownErr := shutdownAuthServer(ctx, server.server)

	if err != nil {
		return emptyString, errors.Join(err, shutdownErr)
	}

	if shutdownErr != nil {
		return emptyString, shutdownErr
	}

	return code, nil
}

func authCallbackHandler(
	state string,
	codeCh chan<- string,
	errCh chan<- error,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		if query.Get("state") != state {
			errCh <- errStateMismatch

			http.Error(
				writer,
				errStateMismatch.Error(),
				http.StatusBadRequest,
			)

			return
		}

		if errText := query.Get("error"); errText != emptyString {
			errCh <- fmt.Errorf("%w: %s", errAuthorizationFailed, errText)

			http.Error(writer, errText, http.StatusBadRequest)

			return
		}

		code := query.Get("code")
		if code == emptyString {
			errCh <- errMissingAuthCode

			http.Error(
				writer,
				errMissingAuthCode.Error(),
				http.StatusBadRequest,
			)

			return
		}

		codeCh <- code

		_, _ = fmt.Fprintln(writer, "Auth complete. You can close this tab.")
	}
}

func awaitAuthCode(
	ctx context.Context,
	codeCh <-chan string,
	errCh <-chan error,
) (string, error) {
	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return emptyString, err
	case <-ctx.Done():
		return emptyString, errAuthTimedOut
	}
}

func shutdownAuthServer(ctx context.Context, server *http.Server) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, authShutdownTimeout)
	defer cancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		return fmt.Errorf("shutdown auth server: %w", err)
	}

	return nil
}

func persistTokens(config *configFile, token tokenBody) error {
	obtainedAt := time.Now().UTC()
	expiresAt := obtainedAt.Add(time.Duration(token.ExpiresIn) * time.Second)

	config.Set(configKeyAccessToken, token.AccessToken)

	if token.RefreshToken != emptyString {
		config.Set(configKeyRefreshToken, token.RefreshToken)
	}

	config.Set(configKeyTokenType, token.TokenType)
	config.Set(configKeyScope, token.Scope)
	config.Set(configKeyUserID, token.UserID)
	config.Set(configKeyTokenExpiresAt, expiresAt.Format(time.RFC3339))
	config.Set(configKeyTokenObtained, obtainedAt.Format(time.RFC3339))

	return config.Save()
}

func removeTokenKeys(config *configFile) {
	config.Unset(configKeyAccessToken)
	config.Unset(configKeyRefreshToken)
	config.Unset(configKeyTokenType)
	config.Unset(configKeyScope)
	config.Unset(configKeyUserID)
	config.Unset(configKeyTokenExpiresAt)
	config.Unset(configKeyTokenObtained)
}

type authStatus struct {
	AccessToken   string
	AccessSource  string
	RefreshToken  string
	RefreshSource string
	Scope         string
	TokenType     string
	UserID        string
	ExpiresAt     time.Time
	Expired       bool
}

func buildAuthStatus(projectConfig, userConfig *configFile) authStatus {
	accessToken := resolveValueSource(
		os.Getenv(envAccessToken),
		projectConfig.Value(configKeyAccessToken),
		userConfig.Value(configKeyAccessToken),
	)

	refreshToken := resolveValueSource(
		os.Getenv(envRefreshToken),
		projectConfig.Value(configKeyRefreshToken),
		userConfig.Value(configKeyRefreshToken),
	)

	scope := resolveValue(
		emptyString,
		emptyString,
		projectConfig.Value(configKeyScope),
		userConfig.Value(configKeyScope),
	)

	tokenType := resolveValue(
		emptyString,
		emptyString,
		projectConfig.Value(configKeyTokenType),
		userConfig.Value(configKeyTokenType),
	)

	userID := resolveValue(
		emptyString,
		emptyString,
		projectConfig.Value(configKeyUserID),
		userConfig.Value(configKeyUserID),
	)

	expiresAt := parseTime(userConfig.Value(configKeyTokenExpiresAt))

	return authStatus{
		AccessToken:   accessToken.Value,
		AccessSource:  accessToken.Source,
		RefreshToken:  refreshToken.Value,
		RefreshSource: refreshToken.Source,
		Scope:         scope,
		TokenType:     tokenType,
		UserID:        userID,
		ExpiresAt:     expiresAt,
		Expired:       isExpired(expiresAt),
	}
}

func (status authStatus) toMap() map[string]any {
	return map[string]any{
		"access_token_present":  status.AccessToken != emptyString,
		"refresh_token_present": status.RefreshToken != emptyString,
		"access_token_source":   status.AccessSource,
		"refresh_token_source":  status.RefreshSource,
		"scope":                 status.Scope,
		"token_type":            status.TokenType,
		"user_id":               status.UserID,
		"token_expires_at":      formatExpiry(status.ExpiresAt),
		"expired":               status.Expired,
	}
}

func (status authStatus) toLines() []string {
	accessLine := "Access token: " +
		presentLabel(status.AccessToken) + " (" +
		status.AccessSource + ")"
	refreshLine := "Refresh token: " +
		presentLabel(status.RefreshToken) + " (" +
		status.RefreshSource + ")"
	scopeLine := "Scope: " +
		defaultIfEmpty(status.Scope, statusUnknownText)
	tokenTypeLine := "Token type: " +
		defaultIfEmpty(status.TokenType, statusUnknownText)
	userLine := "User ID: " +
		defaultIfEmpty(status.UserID, statusUnknownText)
	expiresLine := "Expires at: " + formatExpiry(status.ExpiresAt)
	expiredLine := "Expired: " + strconv.FormatBool(status.Expired)

	return []string{
		accessLine,
		refreshLine,
		scopeLine,
		tokenTypeLine,
		userLine,
		expiresLine,
		expiredLine,
	}
}

func resolveAuthConfig(redirectOverride string) authClientConfig {
	return authClientConfig{
		ClientID:     os.Getenv(envClientID),
		ClientSecret: os.Getenv(envClientSecret),
		RedirectURI: resolveValue(
			redirectOverride,
			os.Getenv(envRedirectURI),
			emptyString,
			emptyString,
		),
	}
}

func requireClientCredentials(config authClientConfig, missingErr error) error {
	if config.ClientID == emptyString || config.ClientSecret == emptyString {
		return newExitError(exitCodeUsage, missingErr)
	}

	return nil
}

func buildLocalRedirectURI(listenAddr string) string {
	return "http://" + listenAddr + "/callback"
}

func confirmLogout(
	opts authLogoutOptions,
	globalOpts globalOptions,
) (bool, error) {
	if opts.Force {
		return true, nil
	}

	ok, err := confirm("Delete stored tokens? [y/N]: ", globalOpts)
	if err != nil {
		return false, newExitError(exitCodeUsage, err)
	}

	return ok, nil
}

func handleAuthOpen(
	ctx context.Context,
	mode authOpenMode,
	authorizeURL string,
) error {
	switch mode {
	case authOpenBrowser:
		err := openBrowser(ctx, authorizeURL)
		if err != nil {
			return writeBrowserOpenError(err)
		}

		return nil
	case authPrintURL:
		return writeAuthURL(authorizeURL)
	default:
		return fmt.Errorf("%w: %d", errInvalidOpenMode, mode)
	}
}

func writeBrowserOpenError(openErr error) error {
	_, err := fmt.Fprintf(os.Stderr, "Failed to open browser: %v\n", openErr)
	if err != nil {
		return fmt.Errorf("write browser error: %w", err)
	}

	return nil
}

func writeAuthURL(authorizeURL string) error {
	_, err := fmt.Fprintf(os.Stderr, "Open this URL:\n%s\n", authorizeURL)
	if err != nil {
		return fmt.Errorf("write auth URL: %w", err)
	}

	return nil
}

type authServer struct {
	server *http.Server
	codeCh chan string
	errCh  chan error
}

func startAuthServer(listenAddr, path, state string) authServer {
	codeCh := make(chan string, authChannelBufferSize)
	errCh := make(chan error, authChannelBufferSize)

	mux := http.NewServeMux()
	mux.HandleFunc(path, authCallbackHandler(state, codeCh, errCh))

	//nolint:exhaustruct // Optional server fields are omitted.
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: authReadHeaderTimeout,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	return authServer{
		server: server,
		codeCh: codeCh,
		errCh:  errCh,
	}
}

func resolveValue(flagValue, envValue, projectValue, userValue string) string {
	if flagValue != emptyString {
		return flagValue
	}

	if envValue != emptyString {
		return envValue
	}

	if projectValue != emptyString {
		return projectValue
	}

	return userValue
}

type resolvedValue struct {
	Value  string
	Source string
}

func resolveValueSource(
	envValue string,
	projectValue string,
	userValue string,
) resolvedValue {
	if envValue != emptyString {
		return resolvedValue{Value: envValue, Source: "env"}
	}

	if projectValue != emptyString {
		return resolvedValue{Value: projectValue, Source: "project"}
	}

	if userValue != emptyString {
		return resolvedValue{Value: userValue, Source: "user"}
	}

	return resolvedValue{Value: emptyString, Source: "none"}
}

func randomState() string {
	data := make([]byte, authStateSizeBytes)

	_, err := rand.Read(data)
	if err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), authNumberBase10)
	}

	return hex.EncodeToString(data)
}

func classifyTokenError(err error) error {
	if err == nil {
		return nil
	}

	var netErr networkError

	if errors.As(err, &netErr) {
		return newExitError(exitCodeNetwork, err)
	}

	var apiErr apiError

	if errors.As(err, &apiErr) {
		return newExitError(exitCodeAPI, err)
	}

	return newExitError(exitCodeFailure, err)
}

func parseTime(value string) time.Time {
	if value == emptyString {
		return time.Time{}
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}

	return parsed
}

func isExpired(expiresAt time.Time) bool {
	if expiresAt.IsZero() {
		return false
	}

	return time.Now().After(expiresAt)
}

func presentLabel(value string) string {
	if value == emptyString {
		return "absent"
	}

	return "present"
}

func defaultIfEmpty(value, fallback string) string {
	if value == emptyString {
		return fallback
	}

	return value
}
