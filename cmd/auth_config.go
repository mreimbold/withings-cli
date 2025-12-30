package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	configKeyClientID       = "client_id"
	configKeyClientSecret   = "client_secret"
	configKeyRedirectURI    = "redirect_uri"
	configKeyAccessToken    = "access_token"
	configKeyRefreshToken   = "refresh_token"
	configKeyScope          = "scope"
	configKeyTokenType      = "token_type"
	configKeyUserID         = "user_id"
	configKeyTokenExpiresAt = "token_expires_at"
	configKeyTokenObtained  = "token_obtained_at"
)

const (
	defaultUserConfigRelPath = ".config/withings-cli/config.toml"
	projectConfigFilename    = "withings-cli.toml"
)

const (
	configDirMode          = 0o700
	configFileMode         = 0o600
	configSplitParts       = 2
	configIndexOffset      = 1
	configLineCountBase    = 0
	configLineEnding       = "\n"
	configCommentLookahead = 1
)

type configFile struct {
	Path     string
	Lines    []string
	Values   map[string]string
	KeyIndex map[string]int
	Exists   bool
}

type configSources struct {
	Project *configFile
	User    *configFile
}

type configKeyValue struct {
	Key   string
	Value string
}

func emptyConfigKeyValue() configKeyValue {
	return configKeyValue{Key: emptyString, Value: emptyString}
}

func loadConfigSources(
	configPath string,
) (configSources, error) {
	projectPath, err := projectConfigPath()
	if err != nil {
		return configSources{}, err
	}

	projectConfig, err := loadConfigFile(projectPath)
	if err != nil {
		return configSources{}, err
	}

	userPath, err := userConfigPath(configPath)
	if err != nil {
		return configSources{}, err
	}

	userConfig, err := loadConfigFile(userPath)
	if err != nil {
		return configSources{}, err
	}

	return configSources{
		Project: projectConfig,
		User:    userConfig,
	}, nil
}

func projectConfigPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return emptyString, fmt.Errorf("get working directory: %w", err)
	}

	return filepath.Join(wd, projectConfigFilename), nil
}

func userConfigPath(override string) (string, error) {
	if override != emptyString {
		return override, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return emptyString, fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, defaultUserConfigRelPath), nil
}

func loadConfigFile(path string) (*configFile, error) {
	config := &configFile{
		Path:     path,
		Lines:    []string{},
		Values:   map[string]string{},
		KeyIndex: map[string]int{},
		Exists:   false,
	}

	//nolint:gosec // Config path is user-controlled by design.
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config, nil
		}

		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	config.Exists = true
	config.Lines = strings.Split(string(data), configLineEnding)
	config.parseLines()

	return config, nil
}

// Value returns the stored value for a key.
func (c *configFile) Value(key string) string {
	return c.Values[key]
}

// Set stores a key/value pair in the config.
func (c *configFile) Set(key, value string) {
	line := fmt.Sprintf("%s = %s", key, tomlQuote(value))
	if idx, ok := c.KeyIndex[key]; ok {
		c.Lines[idx] = line
		c.Values[key] = value

		return
	}

	c.Lines = append(c.Lines, line)
	c.KeyIndex[key] = len(c.Lines) - configIndexOffset
	c.Values[key] = value
}

// Unset removes a key from the config.
func (c *configFile) Unset(key string) {
	idx, ok := c.KeyIndex[key]
	if !ok {
		return
	}

	c.Lines = append(c.Lines[:idx], c.Lines[idx+1:]...)
	delete(c.Values, key)
	delete(c.KeyIndex, key)

	for existingKey, existingIdx := range c.KeyIndex {
		if existingIdx > idx {
			c.KeyIndex[existingKey] = existingIdx - configIndexOffset
		}
	}
}

// Save writes the config to disk.
func (c *configFile) Save() error {
	err := os.MkdirAll(filepath.Dir(c.Path), configDirMode)
	if err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data := strings.Join(c.Lines, configLineEnding)
	if len(c.Lines) > configLineCountBase &&
		!strings.HasSuffix(data, configLineEnding) {
		data += configLineEnding
	}

	err = os.WriteFile(c.Path, []byte(data), configFileMode)
	if err != nil {
		return fmt.Errorf("write config %s: %w", c.Path, err)
	}

	c.Exists = true

	return nil
}

func (c *configFile) parseLines() {
	c.Values = map[string]string{}
	c.KeyIndex = map[string]int{}

	for idx, line := range c.Lines {
		pair, ok := parseConfigLine(line)
		if !ok {
			continue
		}

		c.Values[pair.Key] = pair.Value
		c.KeyIndex[pair.Key] = idx
	}
}

func parseConfigLine(line string) (configKeyValue, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == emptyString {
		return emptyConfigKeyValue(), false
	}

	if isCommentLine(trimmed) || isSectionLine(trimmed) {
		return emptyConfigKeyValue(), false
	}

	withoutComment := stripInlineComment(trimmed)

	pair, ok := splitConfigKeyValue(withoutComment)
	if !ok {
		return emptyConfigKeyValue(), false
	}

	pair.Value = parseConfigValue(pair.Value)

	return pair, true
}

func isCommentLine(line string) bool {
	return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//")
}

func isSectionLine(line string) bool {
	return strings.HasPrefix(line, "[")
}

func splitConfigKeyValue(line string) (configKeyValue, bool) {
	parts := strings.SplitN(line, "=", configSplitParts)
	if len(parts) != configSplitParts {
		return emptyConfigKeyValue(), false
	}

	key := strings.TrimSpace(parts[0])
	if key == emptyString {
		return emptyConfigKeyValue(), false
	}

	value := strings.TrimSpace(parts[1])

	return configKeyValue{Key: key, Value: value}, true
}

func parseConfigValue(value string) string {
	if value == emptyString {
		return emptyString
	}

	unquoted, err := strconv.Unquote(value)
	if err == nil {
		return unquoted
	}

	return strings.Trim(value, "\"'")
}

const commentNotFound = -1

func stripInlineComment(line string) string {
	index := commentStartIndex(line)
	if index == commentNotFound {
		return line
	}

	return line[:index]
}

func commentStartIndex(line string) int {
	state := quoteState{inSingle: false, inDouble: false}
	for index, char := range line {
		if state.toggleQuote(char) {
			continue
		}

		if isCommentStart(line, index, char, state) {
			return index
		}
	}

	return commentNotFound
}

type quoteState struct {
	inSingle bool
	inDouble bool
}

func (state *quoteState) toggleQuote(char rune) bool {
	switch char {
	case '"':
		if !state.inSingle {
			state.inDouble = !state.inDouble
		}

		return true
	case '\'':
		if !state.inDouble {
			state.inSingle = !state.inSingle
		}

		return true
	default:
		return false
	}
}

func isCommentStart(
	line string,
	index int,
	char rune,
	state quoteState,
) bool {
	if state.inSingle || state.inDouble {
		return false
	}

	if char == '#' {
		return true
	}

	if char != '/' {
		return false
	}

	nextIndex := index + configCommentLookahead
	if nextIndex >= len(line) {
		return false
	}

	return line[nextIndex] == '/'
}

func tomlQuote(value string) string {
	return strconv.Quote(value)
}
