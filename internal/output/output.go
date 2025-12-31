// Package output centralizes CLI rendering helpers.
package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mreimbold/withings-cli/internal/app"
)

type envelope struct {
	Ok   bool `json:"ok"`
	Data any  `json:"data,omitempty"`
	Meta any  `json:"meta,omitempty"`
}

// WriteOutput writes data based on output flags.
func WriteOutput(opts app.Options, data any) error {
	if opts.Quiet {
		return nil
	}

	if opts.JSON {
		return writeJSONEnvelope(data)
	}

	switch value := data.(type) {
	case []string:
		return WriteLines(value)
	case string:
		return WriteLine(value)
	default:
		return WriteFormatted("%v\n", value)
	}
}

// WriteRawJSON writes data as pretty JSON.
func WriteRawJSON(opts app.Options, data any) error {
	if opts.Quiet {
		return nil
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(data)
	if err != nil {
		return fmt.Errorf("encode json output: %w", err)
	}

	return nil
}

func writeJSONEnvelope(data any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(envelope{Ok: true, Data: data, Meta: nil})
	if err != nil {
		return fmt.Errorf("encode json output: %w", err)
	}

	return nil
}

// WriteLine writes a single line to stdout.
func WriteLine(value string) error {
	_, err := fmt.Fprintln(os.Stdout, value)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

// WriteLines writes a list of lines to stdout.
func WriteLines(lines []string) error {
	for _, line := range lines {
		err := WriteLine(line)
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteFormatted writes a formatted line to stdout.
func WriteFormatted(format string, value any) error {
	_, err := fmt.Fprintf(os.Stdout, format, value)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
