//nolint:testpackage // test unexported helpers in cmd.
package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const (
	apiMeasureService    = "measure"
	apiMeasureEndpoint   = "https://wbsapi.withings.net/measure"
	apiMeasureEndpointV2 = "https://wbsapi.withings.net/v2/measure"
	apiParamNameKey      = "name"
	apiParseErrFormat    = "parseAPIParams: %v"
	apiNameGotFormat     = "name got %q want %q"
	apiParamsFilePerm    = 0o600
	apiParamValueTest    = "test"
	apiParamValueFile    = "file"
	apiParamValueStdin   = "stdin"
)

// TestAPIServiceEndpoint covers endpoint composition with and without /v2.
func TestAPIServiceEndpoint(t *testing.T) {
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

			got := apiServiceEndpoint(testCase.baseURL, testCase.service)
			if got != testCase.want {
				t.Fatalf("got %q want %q", got, testCase.want)
			}
		})
	}
}

// TestParseAPIParamsRaw ensures raw JSON objects are converted to form values.
func TestParseAPIParamsRaw(t *testing.T) {
	t.Parallel()

	raw := `{"name":"test","limit":10,"active":true}`

	values, err := parseAPIParams(raw)
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

// TestParseAPIParamsFromFile covers @file.json input.
func TestParseAPIParamsFromFile(t *testing.T) {
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

	values, err := parseAPIParams("@" + path)
	if err != nil {
		t.Fatalf(apiParseErrFormat, err)
	}

	if got := values.Get(apiParamNameKey); got != apiParamValueFile {
		t.Fatalf(apiNameGotFormat, got, apiParamValueFile)
	}
}

// TestParseAPIParamsFromStdin covers stdin JSON input.
func TestParseAPIParamsFromStdin(t *testing.T) {
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

	values, err := parseAPIParams("-")
	if err != nil {
		t.Fatalf(apiParseErrFormat, err)
	}

	if got := values.Get(apiParamNameKey); got != apiParamValueStdin {
		t.Fatalf(apiNameGotFormat, got, apiParamValueStdin)
	}
}

// TestParseAPIParamsNotObject rejects non-object JSON.
func TestParseAPIParamsNotObject(t *testing.T) {
	t.Parallel()

	_, err := parseAPIParams(`["nope"]`)
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errParamsNotObject) {
		t.Fatalf("expected errParamsNotObject, got %v", err)
	}
}

// TestEncodeAPIParamValueUnsupported rejects unsupported param types.
func TestEncodeAPIParamValueUnsupported(t *testing.T) {
	t.Parallel()

	_, err := encodeAPIParamValue("bad", map[string]any{"x": 1})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, errUnsupportedParamType) {
		t.Fatalf("expected errUnsupportedParamType, got %v", err)
	}
}
