// Package api provides the generic API passthrough command.
package api //nolint:revive // Package name matches the Withings API endpoint.

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

	"github.com/mreimbold/withings-cli/internal/app"
	"github.com/mreimbold/withings-cli/internal/output"
	"github.com/mreimbold/withings-cli/internal/withings"
)

const (
	floatBitSize    = 64
	paramFilePrefix = "@"
)

var (
	errParamsNotObject      = errors.New("params must be a JSON object")
	errUnsupportedParamType = errors.New("param has unsupported type")
)

// Options captures API call parameters.
type Options struct {
	Service string
	Action  string
	Params  string
	DryRun  bool
}

// Run executes an API call and writes output.
func Run(
	ctx context.Context,
	opts Options,
	appOpts app.Options,
	accessToken string,
) error {
	params, err := parseParams(opts.Params)
	if err != nil {
		return app.NewExitError(app.ExitCodeUsage, err)
	}

	req, body, err := withings.BuildRequest(
		ctx,
		withings.APIBaseURL(appOpts.BaseURL, appOpts.Cloud),
		opts.Service,
		opts.Action,
		accessToken,
		params,
	)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	if opts.DryRun {
		return writeDryRun(appOpts, req.URL.String(), body)
	}

	//nolint:bodyclose // ReadPayload closes the response body.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return app.NewExitError(app.ExitCodeNetwork, err)
	}

	payload, err := withings.ReadPayload(resp)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	return writeResponse(appOpts, payload)
}

func parseParams(raw string) (url.Values, error) {
	if raw == "" {
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

	return encodeParams(params)
}

func readParamsPayload(raw string) ([]byte, error) {
	if raw == "-" {
		return readTrimmed(os.Stdin)
	}

	if path, ok := strings.CutPrefix(raw, paramFilePrefix); ok {
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

func encodeParams(params map[string]any) (url.Values, error) {
	values := url.Values{}

	for key, value := range params {
		if value == nil {
			continue
		}

		encoded, err := encodeParamValue(key, value)
		if err != nil {
			return nil, err
		}

		values.Set(key, encoded)
	}

	return values, nil
}

func encodeParamValue(key string, value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case bool:
		return strconv.FormatBool(typed), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, floatBitSize), nil
	default:
		return "", fmt.Errorf(
			"%w: %q (%T)",
			errUnsupportedParamType,
			key,
			value,
		)
	}
}

func writeDryRun(opts app.Options, endpoint, body string) error {
	lines := []string{
		"POST " + endpoint,
		body,
	}

	err := output.WriteOutput(opts, lines)
	if err != nil {
		return fmt.Errorf("write dry run output: %w", err)
	}

	return nil
}

func writeResponse(opts app.Options, payload []byte) error {
	if opts.JSON {
		var decoded any

		err := json.Unmarshal(payload, &decoded)
		if err != nil {
			return app.NewExitError(
				app.ExitCodeFailure,
				fmt.Errorf("decode api response: %w", err),
			)
		}

		err = output.WriteOutput(opts, decoded)
		if err != nil {
			return fmt.Errorf("write json output: %w", err)
		}

		return nil
	}

	err := output.WriteLine(string(payload))
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
