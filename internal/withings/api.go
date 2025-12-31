// Package withings provides shared API helpers.
package withings

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mreimbold/withings-cli/internal/app"
)

const (
	apiActionKey       = "action"
	apiContentTypeForm = "application/x-www-form-urlencoded"
	apiPathSeparator   = "/"
)

// APIBaseURL resolves the base API URL from overrides and cloud selection.
func APIBaseURL(baseOverride, cloud string) string {
	if baseOverride != "" {
		return strings.TrimRight(baseOverride, "/")
	}

	if cloud == "us" {
		return apiBaseUS
	}

	return apiBaseEU
}

// ServiceEndpoint joins the base URL and service path.
func ServiceEndpoint(baseURL, service string) string {
	trimmed := strings.TrimRight(baseURL, "/")

	return trimmed + apiPathSeparator + service
}

// BuildRequest constructs an authenticated Withings POST request.
func BuildRequest(
	ctx context.Context,
	baseURL string,
	service string,
	action string,
	accessToken string,
	params url.Values,
) (*http.Request, string, error) {
	endpoint := ServiceEndpoint(baseURL, service)

	values := url.Values{}
	values.Set(apiActionKey, action)

	for key, entries := range params {
		for _, entry := range entries {
			values.Add(key, entry)
		}
	}

	body := values.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(body),
	)
	if err != nil {
		return nil, "", fmt.Errorf("build api request: %w", err)
	}

	req.Header.Set("Content-Type", apiContentTypeForm)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	return req, body, nil
}

// ReadPayload reads and validates an API response payload.
func ReadPayload(resp *http.Response) ([]byte, error) {
	payload, err := io.ReadAll(resp.Body)

	closeErr := resp.Body.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("close api response: %w", closeErr)
	}

	if err != nil {
		return nil, app.NewExitError(
			app.ExitCodeFailure,
			errors.Join(fmt.Errorf("read api response: %w", err), closeErr),
		)
	}

	if closeErr != nil {
		return nil, app.NewExitError(app.ExitCodeFailure, closeErr)
	}

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		return nil, app.NewExitError(
			app.ExitCodeAPI,
			fmt.Errorf("%w: %s", ErrAPI, resp.Status),
		)
	}

	return payload, nil
}
