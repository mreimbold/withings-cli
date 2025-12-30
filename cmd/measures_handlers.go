package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	measureServiceName      = "measure"
	measureActionGet        = "getmeas"
	measureTypeParam        = "meastypes"
	measureCategoryParam    = "category"
	measureStartDateParam   = "startdate"
	measureEndDateParam     = "enddate"
	measureLastUpdateParam  = "lastupdate"
	measureUserIDParam      = "userid"
	measureLimitParam       = "limit"
	measureOffsetParam      = "offset"
	measureCategoryReal     = "1"
	measureCategoryGoal     = "2"
	measureTypeDelimiter    = ","
	measureAliasBodyWeight  = "bodyweight"
	measureAliasTemperature = "temperature"
	measureNumberBase10     = 10
	measureEpochBitSize     = 64
)

var (
	errInvalidMeasureType     = errors.New("invalid measure type")
	errInvalidMeasureCategory = errors.New("invalid measure category")
	errInvalidStartTime       = errors.New(
		"invalid --start (expected RFC3339 or epoch)",
	)
	errInvalidEndTime = errors.New(
		"invalid --end (expected RFC3339 or epoch)",
	)
	errInvalidLastUpdate  = errors.New("invalid --last-update (expected epoch)")
	errLastUpdateConflict = errors.New(
		"--last-update cannot be combined with --start or --end",
	)
	errMeasureTypesMissing = errors.New("measure type list is empty")
	errEmptyTimeValue      = errors.New("empty time value")
)

//nolint:gochecknoglobals // Static lookup table for CLI aliases.
var measureTypeMap = map[string]string{
	"weight":                "1",
	"fat_free_mass":         "5",
	"fat_ratio":             "6",
	"fat_mass":              "8",
	"fat_mass_weight":       "8",
	"bp_dia":                "9",
	"bp_sys":                "10",
	"heart_rate":            "11",
	"temp":                  "12",
	"spo2":                  "54",
	"body_temp":             "71",
	"skin_temp":             "73",
	"muscle_mass":           "76",
	"hydration":             "77",
	"bone_mass":             "88",
	"pulse_wave_velocity":   "91",
	measureAliasBodyWeight:  "1",
	measureAliasTemperature: "12",
}

func runMeasuresGet(cmd *cobra.Command, opts measuresGetOptions) error {
	globalOpts, err := readGlobalOptions(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	accessToken, err := ensureAccessToken(ctx, globalOpts)
	if err != nil {
		return err
	}

	params, err := buildMeasureParams(opts)
	if err != nil {
		return newExitError(exitCodeUsage, err)
	}

	apiOpts := apiCallOptions{
		Service: measureServiceName,
		Action:  measureActionGet,
		Params:  emptyString,
		DryRun:  false,
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

	return writeAPIResponse(globalOpts, payload)
}

func buildMeasureParams(opts measuresGetOptions) (url.Values, error) {
	values := url.Values{}

	err := applyMeasureTypes(&values, opts.Types)
	if err != nil {
		return nil, err
	}

	err = applyMeasureCategory(&values, opts.Category)
	if err != nil {
		return nil, err
	}

	err = applyMeasureTimeFilters(&values, opts.TimeRange, opts.LastUpdate)
	if err != nil {
		return nil, err
	}

	applyMeasureUser(&values, opts.User)
	applyMeasurePagination(&values, opts.Pagination)

	return values, nil
}

func applyMeasureTypes(values *url.Values, raw string) error {
	if raw == emptyString {
		return nil
	}

	types, err := parseMeasureTypes(raw)
	if err != nil {
		return err
	}

	if types == emptyString {
		return errMeasureTypesMissing
	}

	values.Set(measureTypeParam, types)

	return nil
}

func applyMeasureCategory(values *url.Values, raw string) error {
	if raw == emptyString {
		return nil
	}

	category, err := parseMeasureCategory(raw)
	if err != nil {
		return err
	}

	values.Set(measureCategoryParam, category)

	return nil
}

func applyMeasureTimeFilters(
	values *url.Values,
	timeRange timeRangeOptions,
	lastUpdate lastUpdateOption,
) error {
	if lastUpdate.LastUpdate < defaultInt64 {
		return errInvalidLastUpdate
	}

	if lastUpdate.LastUpdate > defaultInt64 &&
		(timeRange.Start != emptyString || timeRange.End != emptyString) {
		return errLastUpdateConflict
	}

	if lastUpdate.LastUpdate > defaultInt64 {
		values.Set(
			measureLastUpdateParam,
			strconv.FormatInt(lastUpdate.LastUpdate, measureNumberBase10),
		)
	}

	err := applyMeasureTimeValue(
		values,
		measureStartDateParam,
		timeRange.Start,
		errInvalidStartTime,
	)
	if err != nil {
		return err
	}

	return applyMeasureTimeValue(
		values,
		measureEndDateParam,
		timeRange.End,
		errInvalidEndTime,
	)
}

func applyMeasureTimeValue(
	values *url.Values,
	key string,
	raw string,
	errInvalid error,
) error {
	if raw == emptyString {
		return nil
	}

	epoch, err := parseEpoch(raw)
	if err != nil {
		return fmt.Errorf("%w: %w", errInvalid, err)
	}

	values.Set(key, strconv.FormatInt(epoch, measureNumberBase10))

	return nil
}

func applyMeasureUser(values *url.Values, user userOption) {
	if user.UserID == emptyString {
		return
	}

	values.Set(measureUserIDParam, user.UserID)
}

func applyMeasurePagination(values *url.Values, pagination paginationOptions) {
	if pagination.Limit > defaultInt {
		values.Set(measureLimitParam, strconv.Itoa(pagination.Limit))
	}

	if pagination.Offset > defaultInt {
		values.Set(measureOffsetParam, strconv.Itoa(pagination.Offset))
	}
}

func parseMeasureCategory(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))

	switch normalized {
	case measureCategoryReal, "real":
		return measureCategoryReal, nil
	case measureCategoryGoal, "goal":
		return measureCategoryGoal, nil
	default:
		return emptyString, fmt.Errorf(
			"%w: %q",
			errInvalidMeasureCategory,
			value,
		)
	}
}

func parseMeasureTypes(value string) (string, error) {
	parts := strings.Split(value, measureTypeDelimiter)
	types := make([]string, defaultInt, len(parts))
	seen := map[string]bool{}

	for _, raw := range parts {
		trimmed := strings.ToLower(strings.TrimSpace(raw))
		if trimmed == emptyString {
			continue
		}

		resolved, err := resolveMeasureType(trimmed)
		if err != nil {
			return emptyString, err
		}

		if seen[resolved] {
			continue
		}

		seen[resolved] = true
		types = append(types, resolved)
	}

	return strings.Join(types, measureTypeDelimiter), nil
}

func resolveMeasureType(value string) (string, error) {
	if isDigits(value) {
		return value, nil
	}

	mapped, ok := measureTypeMap[value]
	if !ok {
		return emptyString, fmt.Errorf("%w: %q", errInvalidMeasureType, value)
	}

	return mapped, nil
}

func isDigits(value string) bool {
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}

	return value != emptyString
}

func parseEpoch(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == emptyString {
		return defaultInt64, errEmptyTimeValue
	}

	epoch, err := strconv.ParseInt(
		trimmed,
		measureNumberBase10,
		measureEpochBitSize,
	)
	if err == nil {
		return epoch, nil
	}

	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return defaultInt64, fmt.Errorf("parse RFC3339 time: %w", err)
	}

	return parsed.Unix(), nil
}
