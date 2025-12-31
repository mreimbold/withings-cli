package auth

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/mreimbold/withings-cli/internal/app"
	"github.com/mreimbold/withings-cli/internal/withings"
)

const tokenRefreshSkew = 30 * time.Second

type tokenState struct {
	AccessToken   string
	AccessSource  string
	RefreshToken  string
	RefreshSource string
	ExpiresAt     time.Time
}

// EnsureAccessToken resolves a usable access token, refreshing if needed.
func EnsureAccessToken(
	ctx context.Context,
	opts app.Options,
) (string, error) {
	state, userConfig, err := loadTokenState(opts)
	if err != nil {
		return emptyString, err
	}

	if token := usableAccessToken(state); token != emptyString {
		return token, nil
	}

	return refreshAccessToken(ctx, opts, userConfig, state)
}

func loadTokenState(
	opts app.Options,
) (tokenState, *configFile, error) {
	sources, err := loadConfigSources(opts.Config)
	if err != nil {
		return tokenState{}, nil, err
	}

	state := buildTokenState(sources.Project, sources.User)

	return state, sources.User, nil
}

func usableAccessToken(state tokenState) string {
	if state.AccessToken == emptyString {
		return emptyString
	}

	if shouldRefresh(state.ExpiresAt) {
		return emptyString
	}

	return state.AccessToken
}

func refreshAccessToken(
	ctx context.Context,
	opts app.Options,
	userConfig *configFile,
	state tokenState,
) (string, error) {
	if state.RefreshToken == emptyString {
		return emptyString, app.NewExitError(app.ExitCodeAuth, errAuthRequired)
	}

	authConfig := resolveAuthConfig(emptyString)
	if authConfig.ClientID == emptyString ||
		authConfig.ClientSecret == emptyString {
		return emptyString, app.NewExitError(
			app.ExitCodeAuth,
			errClientCredentialsMissing,
		)
	}

	tokenURL := tokenEndpoint(withings.APIBaseURL(opts.BaseURL, opts.Cloud))

	token, err := refreshToken(
		ctx,
		tokenURL,
		authConfig.ClientID,
		authConfig.ClientSecret,
		state.RefreshToken,
	)
	if err != nil {
		return emptyString, classifyRefreshError(err)
	}

	if shouldPersistRefreshedTokens(state.RefreshSource) {
		err = persistTokens(userConfig, token)
		if err != nil {
			return emptyString, err
		}
	}

	return token.AccessToken, nil
}

func buildTokenState(projectConfig, userConfig *configFile) tokenState {
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

	expiresAt := parseTime(userConfig.Value(configKeyTokenExpiresAt))

	return tokenState{
		AccessToken:   accessToken.Value,
		AccessSource:  accessToken.Source,
		RefreshToken:  refreshToken.Value,
		RefreshSource: refreshToken.Source,
		ExpiresAt:     expiresAt,
	}
}

func shouldRefresh(expiresAt time.Time) bool {
	if expiresAt.IsZero() {
		return false
	}

	return time.Now().After(expiresAt.Add(-tokenRefreshSkew))
}

func shouldPersistRefreshedTokens(source string) bool {
	return source == "user"
}

func classifyRefreshError(err error) error {
	if err == nil {
		return nil
	}

	var netErr networkError
	if errors.As(err, &netErr) {
		return app.NewExitError(app.ExitCodeNetwork, err)
	}

	return app.NewExitError(app.ExitCodeAuth, err)
}
