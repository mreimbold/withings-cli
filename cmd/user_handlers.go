package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const (
	userServiceName     = "v2/user"
	userServiceShort    = "user"
	userServiceV2Suffix = "/v2"
	userActionMe        = "getbyuserid"
	userActionList      = "list"
	userUserIDParam     = "userid"
	userNumberBase10    = 10
	userRowsHeaderCount = 1
	userTableMinWidth   = 0
	userTableTabWidth   = 0
	userTablePadding    = 2
	userTablePadChar    = ' '
	userTableFlags      = 0
)

var errUserIDRequired = errors.New(
	"user id missing; login first or set user_id in config",
)

func runUserMe(cmd *cobra.Command) error {
	return runUserProfile(cmd, userActionMe)
}

func runUserList(cmd *cobra.Command) error {
	return runUserProfile(cmd, userActionList)
}

func runUserProfile(cmd *cobra.Command, action string) error {
	globalOpts, err := readGlobalOptions(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	accessToken, err := ensureAccessToken(ctx, globalOpts)
	if err != nil {
		return err
	}

	params, err := buildUserParams(globalOpts, action)
	if err != nil {
		return newExitError(exitCodeUsage, err)
	}

	apiOpts := apiCallOptions{
		Service: userServiceForBase(
			apiBaseURL(globalOpts.BaseURL, globalOpts.Cloud),
		),
		Action: action,
		Params: emptyString,
		DryRun: false,
	}

	req, _, err := buildAPICallRequest(
		ctx,
		globalOpts,
		apiOpts,
		accessToken,
		params,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return newExitError(exitCodeNetwork, err)
	}

	payload, err := readAPIPayload(resp)
	if err != nil {
		return err
	}

	return writeUserResponse(globalOpts, payload)
}

func userServiceForBase(baseURL string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, userServiceV2Suffix) {
		return userServiceShort
	}

	return userServiceName
}

func buildUserParams(opts globalOptions, action string) (url.Values, error) {
	if action == userActionList {
		return url.Values{}, nil
	}

	userID, err := resolveUserID(opts.Config)
	if err != nil {
		return nil, err
	}

	if userID == emptyString {
		return nil, errUserIDRequired
	}

	values := url.Values{}
	values.Set(userUserIDParam, userID)

	return values, nil
}

func resolveUserID(configPath string) (string, error) {
	sources, err := loadConfigSources(configPath)
	if err != nil {
		return emptyString, err
	}

	return resolveValue(
		emptyString,
		emptyString,
		sources.Project.Value(configKeyUserID),
		sources.User.Value(configKeyUserID),
	), nil
}

type userResponse struct {
	Status int      `json:"status"`
	Body   userBody `json:"body"`
	Error  string   `json:"error"`
	Detail string   `json:"detail"`
}

type userBody struct {
	User  map[string]any   `json:"user"`
	Users []map[string]any `json:"users"`
}

type userRow struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Birthdate string
	Gender    string
	Timezone  string
}

func writeUserResponse(opts globalOptions, payload []byte) error {
	decoded, err := decodeUserResponse(payload)
	if err != nil {
		return err
	}

	if opts.Quiet {
		return nil
	}

	if opts.JSON {
		return writeRawJSON(opts, decoded.Body)
	}

	profiles := userProfiles(decoded.Body)
	rows := buildUserRows(profiles)

	if opts.Plain {
		return writeLines(formatUserLines(rows))
	}

	table, err := formatUserTable(rows)
	if err != nil {
		return err
	}

	return writeLine(table)
}

func decodeUserResponse(payload []byte) (userResponse, error) {
	var decoded userResponse

	err := json.Unmarshal(payload, &decoded)
	if err != nil {
		return userResponse{}, newExitError(
			exitCodeFailure,
			fmt.Errorf("decode api response: %w", err),
		)
	}

	if decoded.Status != withingsStatusOK {
		message := decoded.Error
		if message == emptyString {
			message = decoded.Detail
		}

		if message == emptyString {
			message = strings.TrimSpace(string(payload))
		}

		return userResponse{}, newExitError(
			exitCodeAPI,
			fmt.Errorf("%w: %d: %s", errWithingsAPI, decoded.Status, message),
		)
	}

	return decoded, nil
}

func userProfiles(body userBody) []map[string]any {
	if len(body.Users) > defaultInt {
		return body.Users
	}

	if len(body.User) > defaultInt {
		return []map[string]any{body.User}
	}

	return []map[string]any{}
}

func buildUserRows(profiles []map[string]any) []userRow {
	rows := make([]userRow, defaultInt, len(profiles))

	for _, profile := range profiles {
		rows = append(rows, userRow{
			ID:        userString(profile, "id", "userid"),
			FirstName: userString(profile, "firstname", "first_name"),
			LastName:  userString(profile, "lastname", "last_name"),
			Email:     userString(profile, "email"),
			Birthdate: userString(profile, "birthdate"),
			Gender:    userString(profile, "gender"),
			Timezone:  userString(profile, "timezone"),
		})
	}

	return rows
}

func userString(profile map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := profile[key]
		if !ok {
			continue
		}

		formatted := formatUserValue(value)
		if formatted != emptyString {
			return formatted
		}
	}

	return emptyString
}

func formatUserValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return emptyString
	case string:
		return typed
	case float64:
		if typed == math.Trunc(typed) {
			return strconv.FormatInt(int64(typed), userNumberBase10)
		}

		return strconv.FormatFloat(typed, 'f', -1, apiFloatBitSize)
	case bool:
		return strconv.FormatBool(typed)
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func formatUserTable(rows []userRow) (string, error) {
	var buffer bytes.Buffer

	writer := tabwriter.NewWriter(
		&buffer,
		userTableMinWidth,
		userTableTabWidth,
		userTablePadding,
		userTablePadChar,
		userTableFlags,
	)
	_, _ = fmt.Fprintln(
		writer,
		"ID\tFirst Name\tLast Name\tEmail\tBirthdate\tGender\tTimezone",
	)

	for _, row := range rows {
		_, _ = fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ID,
			row.FirstName,
			row.LastName,
			row.Email,
			row.Birthdate,
			row.Gender,
			row.Timezone,
		)
	}

	err := writer.Flush()
	if err != nil {
		return emptyString, fmt.Errorf("render user table: %w", err)
	}

	return strings.TrimRight(buffer.String(), "\n"), nil
}

func formatUserLines(rows []userRow) []string {
	lines := make([]string, defaultInt, len(rows)+userRowsHeaderCount)
	lines = append(
		lines,
		"id\tfirst_name\tlast_name\temail\tbirthdate\tgender\ttimezone",
	)

	for _, row := range rows {
		lines = append(lines, strings.Join([]string{
			row.ID,
			row.FirstName,
			row.LastName,
			row.Email,
			row.Birthdate,
			row.Gender,
			row.Timezone,
		}, "\t"))
	}

	return lines
}
