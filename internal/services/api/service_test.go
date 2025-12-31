//nolint:testpackage,revive // test unexported helpers; package name matches Withings API endpoint.
package api

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mreimbold/withings-cli/internal/withings"
)

const (
	apiMeasureService    = "measure"
	apiMeasureEndpoint   = "https://wbsapi.withings.net/measure"
	apiMeasureEndpointV2 = "https://wbsapi.withings.net/v2/measure"
	apiParamNameKey      = "name"
	apiParseErrFormat    = "parseParams: %v"
	apiNameGotFormat     = "name got %q want %q"
	apiParamsFilePerm    = 0o600
	apiParamValueTest    = "test"
	apiParamValueFile    = "file"
	apiParamValueStdin   = "stdin"
)

// TestServiceEndpoint covers endpoint composition with and without /v2.
func TestServiceEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		service string
		want    string
	}{
		{
			name:    "base-no-version",
			baseURL: "https://wbsapi.withings.net",
			service: apiMeasureService,
			want:    apiMeasureEndpoint,
		},
		{
			name:    "base-trailing-slash",
			baseURL: "https://wbsapi.withings.net/",
			service: apiMeasureService,
			want:    apiMeasureEndpoint,
		},
		{
			name:    "base-with-version",
			baseURL: "https://wbsapi.withings.net/v2",
			service: apiMeasureService,
			want:    apiMeasureEndpointV2,
		},
		{
			name:    "base-with-version-slash",
			baseURL: "https://wbsapi.withings.net/v2/",
			service: apiMeasureService,
			want:    apiMeasureEndpointV2,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := withings.ServiceEndpoint(testCase.baseURL, testCase.service)
			if got != testCase.want {
				t.Fatalf("got %q want %q", got, testCase.want)
			}
		})
	}
}

// TestParseParamsRaw ensures raw JSON objects are converted to form values.
func TestParseParamsRaw(t *testing.T) {
	t.Parallel()

	raw := `{"name":"test","limit":10,"active":true}`

	values, err := parseParams(raw)
	if err != nil {
		t.Fatalf(apiParseErrFormat, err)
	}

	if got := values.Get(apiParamNameKey); got != apiParamValueTest {
		t.Fatalf(apiNameGotFormat, got, apiParamValueTest)
	}

	if got := values.Get("limit"); got != "10" {
		t.Fatalf("limit got %q want %q", got, "10")
	}

	if got := values.Get("active"); got != "true" {
		t.Fatalf("active got %q want %q", got, "true")
	}
}

// TestParseParamsFromFile covers @file.json input.
func TestParseParamsFromFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "params.json")

	err := os.WriteFile(
		path,
		[]byte(`{"name":"file"}`),
		apiParamsFilePerm,
	)
	if err != nil {
		t.Fatalf("write temp params: %v", err)
	}

	values, err := parseParams("@" + path)
	if err != nil {
		t.Fatalf(apiParseErrFormat, err)
	}

	if got := values.Get(apiParamNameKey); got != apiParamValueFile {
		t.Fatalf(apiNameGotFormat, got, apiParamValueFile)
	}
}

// TestParseParamsFromStdin covers stdin JSON input.
func TestParseParamsFromStdin(t *testing.T) {
	t.Parallel()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	oldStdin := os.Stdin
	os.Stdin = reader

	t.Cleanup(func() {
		_ = reader.Close()
		os.Stdin = oldStdin
	})

	_, _ = writer.WriteString(`{"name":"stdin"}`)
	_ = writer.Close()

	values, err := parseParams("-")
	if err != nil {
		t.Fatalf(apiParseErrFormat, err)
	}

	if got := values.Get(apiParamNameKey); got != apiParamValueStdin {
		t.Fatalf(apiNameGotFormat, got, apiParamValueStdin)
	}
}

// TestParseParamsNotObject rejects non-object JSON.
func TestParseParamsNotObject(t *testing.T) {
	t.Parallel()

	_, err := parseParams(`["nope"]`)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errParamsNotObject) {
		t.Fatalf("expected errParamsNotObject, got %v", err)
	}
}

// TestEncodeParamValueUnsupported rejects unsupported param types.
func TestEncodeParamValueUnsupported(t *testing.T) {
	t.Parallel()

	_, err := encodeParamValue("bad", map[string]any{"x": 1})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errUnsupportedParamType) {
		t.Fatalf("expected errUnsupportedParamType, got %v", err)
	}
}
