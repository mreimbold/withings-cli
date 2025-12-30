package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	apiActionKey       = "action"
	apiContentTypeForm = "application/x-www-form-urlencoded"
	apiPathSeparator   = "/"
	apiVersionSegment  = "/v2"
	apiFloatBitSize    = 64
)

var (
	errParamsNotObject      = errors.New("params must be a JSON object")
	errUnsupportedParamType = errors.New("param has unsupported type")
)

func runAPICall(cmd *cobra.Command, opts apiCallOptions) error {
	globalOpts, err := readGlobalOptions(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	accessToken, err := ensureAccessToken(ctx, globalOpts)
	if err != nil {
		return err
	}

	params, err := parseAPIParams(opts.Params)
	if err != nil {
		return newExitError(exitCodeUsage, err)
	}

	req, body, err := buildAPICallRequest(
		ctx,
		globalOpts,
		opts,
		accessToken,
		params,
	)
	if err != nil {
		return err
	}

	if opts.DryRun {
		return writeAPIDryRun(globalOpts, req.URL.String(), body)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return newExitError(exitCodeNetwork, err)
	}

	payload, err := readAPIPayload(resp)
	if err != nil {
		return err
	}

	return writeAPIResponse(globalOpts, payload)
}

func buildAPICallRequest(
	ctx context.Context,
	globalOpts globalOptions,
	apiOpts apiCallOptions,
	accessToken string,
	params url.Values,
) (*http.Request, string, error) {
	endpoint := apiServiceEndpoint(
		apiBaseURL(globalOpts.BaseURL, globalOpts.Cloud),
		apiOpts.Service,
	)

	values := url.Values{}
	values.Set(apiActionKey, apiOpts.Action)

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
		return nil, emptyString, fmt.Errorf("build api request: %w", err)
	}

	req.Header.Set("Content-Type", apiContentTypeForm)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	return req, body, nil
}

func apiServiceEndpoint(baseURL, service string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, apiVersionSegment) {
		return trimmed + apiPathSeparator + service
	}

	return trimmed + apiVersionSegment + apiPathSeparator + service
}

func parseAPIParams(raw string) (url.Values, error) {
	if raw == emptyString {
		return url.Values{}, nil
	}

	payload, err := readParamsPayload(raw)
	if err != nil {
		return nil, err
	}

	var decoded any

	err = json.Unmarshal(payload, &decoded)
	if err != nil {
		return nil, fmt.Errorf("decode params: %w", err)
	}

	params, ok := decoded.(map[string]any)
	if !ok {
		return nil, errParamsNotObject
	}

	return encodeAPIParams(params)
}

func readParamsPayload(raw string) ([]byte, error) {
	if raw == "-" {
		return readTrimmed(os.Stdin)
	}

	if path, ok := strings.CutPrefix(raw, "@"); ok {
		return readTrimmedPath(path)
	}

	return []byte(strings.TrimSpace(raw)), nil
}

func readTrimmedPath(path string) ([]byte, error) {
	//nolint:gosec // User-supplied path is expected for CLI params.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read params %s: %w", path, err)
	}

	return bytes.TrimSpace(data), nil
}

func readTrimmed(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read params: %w", err)
	}

	return bytes.TrimSpace(data), nil
}

func encodeAPIParams(params map[string]any) (url.Values, error) {
	values := url.Values{}

	for key, value := range params {
		if value == nil {
			continue
		}

		encoded, err := encodeAPIParamValue(key, value)
		if err != nil {
			return nil, err
		}

		values.Set(key, encoded)
	}

	return values, nil
}

func encodeAPIParamValue(key string, value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case bool:
		return strconv.FormatBool(typed), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, apiFloatBitSize), nil
	default:
		return emptyString, fmt.Errorf(
			"%w: %q (%T)",
			errUnsupportedParamType,
			key,
			value,
		)
	}
}

func readAPIPayload(resp *http.Response) ([]byte, error) {
	payload, err := io.ReadAll(resp.Body)

	closeErr := resp.Body.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("close api response: %w", closeErr)
	}

	if err != nil {
		return nil, newExitError(
			exitCodeFailure,
			errors.Join(fmt.Errorf("read api response: %w", err), closeErr),
		)
	}

	if closeErr != nil {
		return nil, newExitError(exitCodeFailure, closeErr)
	}

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newExitError(
			exitCodeAPI,
			fmt.Errorf("%w: %s", errWithingsAPI, resp.Status),
		)
	}

	return payload, nil
}

func writeAPIDryRun(
	opts globalOptions,
	endpoint string,
	body string,
) error {
	lines := []string{
		"POST " + endpoint,
		body,
	}

	return writeOutput(opts, lines)
}

func writeAPIResponse(opts globalOptions, payload []byte) error {
	if opts.JSON {
		var decoded any

		err := json.Unmarshal(payload, &decoded)
		if err != nil {
			return newExitError(
				exitCodeFailure,
				fmt.Errorf("decode api response: %w", err),
			)
		}

		return writeOutput(opts, decoded)
	}

	return writeLine(string(payload))
}
