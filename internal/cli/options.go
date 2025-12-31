package cli

import (
	"fmt"

	"github.com/mreimbold/withings-cli/internal/app"
	"github.com/spf13/pflag"
)

type flagReader interface {
	GetBool(name string) (bool, error)
	GetCount(name string) (int, error)
	GetString(name string) (string, error)
}

const flagReadErrorFormat = "read --%s: %w"

func readGlobalOptions(flags *pflag.FlagSet) (app.Options, error) {
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

func defaultGlobalOptions() app.Options {
	return app.Options{
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

func applyOutputFlags(flags flagReader, opts *app.Options) error {
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

func applyConfigFlags(flags flagReader, opts *app.Options) error {
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
