package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mreimbold/withings-cli/internal/withings"
)

const (
	defaultAuthScope      = "user.metrics,user.activity"
	withingsAccountEU     = "https://account.withings.com"
	withingsAccountUS     = "https://account.us.withingsmed.com"
	withingsAuthorizePath = "/oauth2_user/authorize2"

	//nolint:gosec // Static API path, not a credential.
	withingsOAuthTokenPath = "/v2/oauth2"

	oauthActionKey          = "action"
	oauthActionRequestToken = "requesttoken"
	oauthClientIDKey        = "client_id"
	oauthClientSecretKey    = "client_secret"
	oauthCodeKey            = "code"
	oauthContentTypeForm    = "application/x-www-form-urlencoded"
	oauthGrantTypeKey       = "grant_type"
	oauthGrantAuthorization = "authorization_code"
	oauthGrantRefresh       = "refresh_token"
	oauthRedirectURIKey     = "redirect_uri"
	oauthRefreshTokenKey    = "refresh_token"
	oauthResponseTypeKey    = "response_type"
	oauthResponseTypeCode   = "code"
	oauthScopeKey           = "scope"
	oauthStateKey           = "state"
	tokenRequestTimeout     = 30 * time.Second
	tokenNullLiteral        = "null"
	tokenQuoteByte          = '"'
)

type tokenResponse struct {
	Status int       `json:"status"`
	Body   tokenBody `json:"body"`
	Error  string    `json:"error"`
	Detail string    `json:"detail"`
}

type tokenUserID string

// UnmarshalJSON accepts string or numeric user IDs from Withings.
func (u *tokenUserID) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == emptyString || trimmed == tokenNullLiteral {
		*u = emptyString

		return nil
	}

	if trimmed[defaultInt] == tokenQuoteByte {
		var value string

		err := json.Unmarshal(data, &value)
		if err != nil {
			return fmt.Errorf("%w: %w", errTokenUserIDDecode, err)
		}

		*u = tokenUserID(value)

		return nil
	}

	var value json.Number

	err := json.Unmarshal(data, &value)
	if err != nil {
		return fmt.Errorf("%w: %w", errTokenUserIDType, err)
	}

	*u = tokenUserID(value.String())

	return nil
}

//nolint:tagliatelle // Withings API uses snake_case JSON fields.
type tokenBody struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	TokenType    string      `json:"token_type"`
	Scope        string      `json:"scope"`
	UserID       tokenUserID `json:"userid"`
}

func buildAuthorizeURL(
	baseURL string,
	clientID string,
	redirectURI string,
	scope string,
	state string,
) (string, error) {
	resolvedScope := scope
	if resolvedScope == emptyString {
		resolvedScope = defaultAuthScope
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return emptyString, fmt.Errorf("invalid authorize base URL: %w", err)
	}

	parsedURL.Path = withingsAuthorizePath
	query := parsedURL.Query()
	query.Set(oauthResponseTypeKey, oauthResponseTypeCode)
	query.Set(oauthClientIDKey, clientID)
	query.Set(oauthRedirectURIKey, redirectURI)
	query.Set(oauthStateKey, state)
	query.Set(oauthScopeKey, resolvedScope)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

func exchangeToken(
	ctx context.Context,
	tokenURL string,
	clientID string,
	clientSecret string,
	code string,
	redirectURI string,
) (tokenBody, error) {
	values := url.Values{}
	values.Set(oauthActionKey, oauthActionRequestToken)
	values.Set(oauthGrantTypeKey, oauthGrantAuthorization)
	values.Set(oauthClientIDKey, clientID)
	values.Set(oauthClientSecretKey, clientSecret)
	values.Set(oauthCodeKey, code)
	values.Set(oauthRedirectURIKey, redirectURI)

	return doTokenRequest(ctx, tokenURL, values)
}

func refreshToken(
	ctx context.Context,
	tokenURL string,
	clientID string,
	clientSecret string,
	refresh string,
) (tokenBody, error) {
	values := url.Values{}
	values.Set(oauthActionKey, oauthActionRequestToken)
	values.Set(oauthGrantTypeKey, oauthGrantRefresh)
	values.Set(oauthClientIDKey, clientID)
	values.Set(oauthClientSecretKey, clientSecret)
	values.Set(oauthRefreshTokenKey, refresh)

	return doTokenRequest(ctx, tokenURL, values)
}

func doTokenRequest(
	ctx context.Context,
	tokenURL string,
	values url.Values,
) (tokenBody, error) {
	requestCtx, cancel := context.WithTimeout(ctx, tokenRequestTimeout)
	defer cancel()

	req, err := buildTokenRequest(requestCtx, tokenURL, values)
	if err != nil {
		return tokenBody{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return tokenBody{}, networkError{err: err}
	}

	payload, err := readTokenPayload(resp.Body)

	closeErr := resp.Body.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("close token response: %w", closeErr)
	}

	if err != nil {
		return tokenBody{}, apiError{err: errors.Join(err, closeErr)}
	}

	if closeErr != nil {
		return tokenBody{}, apiError{err: closeErr}
	}

	err = ensureTokenHTTPStatus(resp.StatusCode, payload)
	if err != nil {
		return tokenBody{}, apiError{err: err}
	}

	body, err := decodeTokenResponse(payload)
	if err != nil {
		return tokenBody{}, apiError{err: err}
	}

	return body, nil
}

func buildTokenRequest(
	ctx context.Context,
	tokenURL string,
	values url.Values,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		tokenURL,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}

	req.Header.Set("Content-Type", oauthContentTypeForm)

	return req, nil
}

func readTokenPayload(body io.Reader) ([]byte, error) {
	payload, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	return payload, nil
}

func ensureTokenHTTPStatus(statusCode int, payload []byte) error {
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return fmt.Errorf(
			"%w: HTTP %d: %s",
			errTokenRequestFailed,
			statusCode,
			strings.TrimSpace(string(payload)),
		)
	}

	return nil
}

func decodeTokenResponse(payload []byte) (tokenBody, error) {
	var decoded tokenResponse

	err := json.Unmarshal(payload, &decoded)
	if err != nil {
		return tokenBody{}, fmt.Errorf("decode token response: %w", err)
	}

	if decoded.Status != withings.StatusOK {
		message := decoded.Error
		if message == emptyString {
			message = decoded.Detail
		}

		if message == emptyString {
			message = strings.TrimSpace(string(payload))
		}

		return tokenBody{}, fmt.Errorf(
			"%w: %d: %s",
			errWithingsAPI,
			decoded.Status,
			message,
		)
	}

	return decoded.Body, nil
}

type networkError struct {
	err error
}

// Error returns the wrapped error message.
func (e networkError) Error() string {
	return e.err.Error()
}

// Unwrap returns the underlying error.
func (e networkError) Unwrap() error {
	return e.err
}

type apiError struct {
	err error
}

// Error returns the wrapped error message.
func (e apiError) Error() string {
	return e.err.Error()
}

// Unwrap returns the underlying error.
func (e apiError) Unwrap() error {
	return e.err
}

func accountBaseURL(cloud string) string {
	if cloud == "us" {
		return withingsAccountUS
	}

	return withingsAccountEU
}

func tokenEndpoint(baseURL string) string {
	return strings.TrimRight(baseURL, "/") + withingsOAuthTokenPath
}
