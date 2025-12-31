package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/mreimbold/withings-cli/internal/app"
)

func readLine(prompt string, opts app.Options) (string, error) {
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

func confirm(prompt string, opts app.Options) (bool, error) {
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
		return statusUnknownText
	}

	return expiresAt.Format(time.RFC3339)
}
