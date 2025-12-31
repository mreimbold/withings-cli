package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type outputEnvelope struct {
	Ok   bool `json:"ok"`
	Data any  `json:"data,omitempty"`
	Meta any  `json:"meta,omitempty"`
}

func readGlobalOptions(cmd *cobra.Command) (globalOptions, error) {
	flags := cmd.Root().PersistentFlags()

	opts := defaultGlobalOptions()

	err := applyOutputFlags(flags, &opts)
	if err != nil {
		return opts, err
	}

	err = applyConfigFlags(flags, &opts)
	if err != nil {
		return opts, err
	}

	return opts, nil
}

type flagReader interface {
	GetBool(name string) (bool, error)
	GetCount(name string) (int, error)
	GetString(name string) (string, error)
}

const flagReadErrorFormat = "read --%s: %w"

func defaultGlobalOptions() globalOptions {
	return globalOptions{
		Verbose: defaultInt,
		Quiet:   false,
		JSON:    false,
		Plain:   false,
		NoColor: false,
		NoInput: false,
		Config:  emptyString,
		Cloud:   emptyString,
		BaseURL: emptyString,
	}
}

func applyOutputFlags(flags flagReader, opts *globalOptions) error {
	verbose, err := getFlagCount(flags, "verbose")
	if err != nil {
		return err
	}

	opts.Verbose = verbose

	quiet, err := getFlagBool(flags, "quiet")
	if err != nil {
		return err
	}

	opts.Quiet = quiet

	jsonOutput, err := getFlagBool(flags, "json")
	if err != nil {
		return err
	}

	opts.JSON = jsonOutput

	plainOutput, err := getFlagBool(flags, "plain")
	if err != nil {
		return err
	}

	opts.Plain = plainOutput

	noColor, err := getFlagBool(flags, "no-color")
	if err != nil {
		return err
	}

	opts.NoColor = noColor

	noInput, err := getFlagBool(flags, "no-input")
	if err != nil {
		return err
	}

	opts.NoInput = noInput

	return nil
}

func applyConfigFlags(flags flagReader, opts *globalOptions) error {
	configPath, err := getFlagString(flags, "config")
	if err != nil {
		return err
	}

	opts.Config = configPath

	cloud, err := getFlagString(flags, "cloud")
	if err != nil {
		return err
	}

	opts.Cloud = cloud

	baseURL, err := getFlagString(flags, "base-url")
	if err != nil {
		return err
	}

	opts.BaseURL = baseURL

	return nil
}

func getFlagCount(flags flagReader, name string) (int, error) {
	value, err := flags.GetCount(name)
	if err != nil {
		return defaultInt, fmt.Errorf(flagReadErrorFormat, name, err)
	}

	return value, nil
}

func getFlagBool(flags flagReader, name string) (bool, error) {
	value, err := flags.GetBool(name)
	if err != nil {
		return false, fmt.Errorf(flagReadErrorFormat, name, err)
	}

	return value, nil
}

func getFlagString(flags flagReader, name string) (string, error) {
	value, err := flags.GetString(name)
	if err != nil {
		return emptyString, fmt.Errorf(flagReadErrorFormat, name, err)
	}

	return value, nil
}

func writeOutput(opts globalOptions, data any) error {
	if opts.Quiet {
		return nil
	}

	if opts.JSON {
		return writeJSONOutput(data)
	}

	switch value := data.(type) {
	case []string:
		return writeLines(value)
	case string:
		return writeLine(value)
	default:
		return writeFormatted("%v\n", value)
	}
}

func writeJSONOutput(data any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	err := encoder.Encode(outputEnvelope{Ok: true, Data: data, Meta: nil})
	if err != nil {
		return fmt.Errorf("encode json output: %w", err)
	}

	return nil
}

func writeLine(value string) error {
	_, err := fmt.Fprintln(os.Stdout, value)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func writeLines(lines []string) error {
	for _, line := range lines {
		err := writeLine(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeFormatted(format string, value any) error {
	_, err := fmt.Fprintf(os.Stdout, format, value)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func readLine(prompt string, opts globalOptions) (string, error) {
	if opts.NoInput || !isTerminal(os.Stdin) {
		return emptyString, errInputRequired
	}

	if prompt != emptyString {
		_, err := fmt.Fprint(os.Stderr, prompt)
		if err != nil {
			return emptyString, fmt.Errorf("write prompt: %w", err)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return emptyString, fmt.Errorf("read input: %w", err)
	}

	return strings.TrimSpace(line), nil
}

func confirm(prompt string, opts globalOptions) (bool, error) {
	answer, err := readLine(prompt, opts)
	if err != nil {
		return false, err
	}

	answer = strings.ToLower(strings.TrimSpace(answer))

	return answer == "y" || answer == "yes", nil
}

const emptyFileMode os.FileMode = 0

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != emptyFileMode
}

func openBrowser(ctx context.Context, target string) error {
	var command *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		command = exec.CommandContext(ctx, "open", target)
	case "windows":
		command = exec.CommandContext(
			ctx,
			"rundll32",
			"url.dll,FileProtocolHandler",
			target,
		)
	default:
		command = exec.CommandContext(ctx, "xdg-open", target)
	}

	err := command.Start()
	if err != nil {
		return fmt.Errorf("open browser: %w", err)
	}

	return nil
}

func formatExpiry(expiresAt time.Time) string {
	if expiresAt.IsZero() {
		return "unknown"
	}

	return expiresAt.Format(time.RFC3339)
}
