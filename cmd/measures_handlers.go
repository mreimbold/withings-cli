package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
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
	measureCategoryRealText = "real"
	measureCategoryGoalText = "goal"
	measureTypeDelimiter    = ","
	measureAliasBodyWeight  = "bodyweight"
	measureAliasTemperature = "temperature"
	measureNumberBase10     = 10
	measureEpochBitSize     = 64
	measureZeroString       = "0"
	measureUnitBase         = "1"
	measureUnitExponent     = "1e"
	measureNegativeSign     = "-"
	measureDecimalSeparator = "."
	measureRowsHeaderCount  = 1
	measureTableMinWidth    = 0
	measureTableTabWidth    = 0
	measureTablePadding     = 2
	measureTablePadChar     = ' '
	measureTableFlags       = 0
	measureScalePad         = 1
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

	return writeMeasuresResponse(globalOpts, payload)
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
	case measureCategoryReal, measureCategoryRealText:
		return measureCategoryReal, nil
	case measureCategoryGoal, measureCategoryGoalText:
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

type measuresResponse struct {
	Status int          `json:"status"`
	Body   measuresBody `json:"body"`
	Error  string       `json:"error"`
	Detail string       `json:"detail"`
}

type measuresBody struct {
	UpdateTime    int64          `json:"updatetime"`
	Timezone      string         `json:"timezone"`
	MeasureGroups []measureGroup `json:"measuregrps"`
}

type measureGroup struct {
	GroupID  int64         `json:"grpid"`
	Attrib   int           `json:"attrib"`
	Date     int64         `json:"date"`
	Category int           `json:"category"`
	Measures []measureItem `json:"measures"`
}

type measureItem struct {
	Type  int   `json:"type"`
	Value int64 `json:"value"`
	Unit  int   `json:"unit"`
}

type measureRow struct {
	Time     string
	Type     string
	Value    string
	Unit     string
	Category string
}

//nolint:gochecknoglobals // Static lookup tables for measure metadata.
var (
	measureTypeNameByID = map[string]string{
		"1":  "weight",
		"5":  "fat_free_mass",
		"6":  "fat_ratio",
		"8":  "fat_mass",
		"9":  "bp_dia",
		"10": "bp_sys",
		"11": "heart_rate",
		"12": "temp",
		"54": "spo2",
		"71": "body_temp",
		"73": "skin_temp",
		"76": "muscle_mass",
		"77": "hydration",
		"88": "bone_mass",
		"91": "pulse_wave_velocity",
	}
	measureUnitByTypeID = map[string]string{
		"1":  "kg",
		"5":  "kg",
		"6":  "%",
		"8":  "kg",
		"9":  "mmHg",
		"10": "mmHg",
		"11": "bpm",
		"12": "C",
		"54": "%",
		"71": "C",
		"73": "C",
		"76": "kg",
		"77": "%",
		"88": "kg",
		"91": "m/s",
	}
)

func writeMeasuresResponse(opts globalOptions, payload []byte) error {
	decoded, err := decodeMeasuresResponse(payload)
	if err != nil {
		return err
	}

	if opts.Quiet {
		return nil
	}

	if opts.JSON {
		return writeRawJSON(opts, decoded.Body)
	}

	rows := buildMeasureRows(decoded.Body)

	if opts.Plain {
		return writeLines(formatMeasureLines(rows))
	}

	table, err := formatMeasureTable(rows)
	if err != nil {
		return err
	}

	return writeLine(table)
}

func decodeMeasuresResponse(payload []byte) (measuresResponse, error) {
	var decoded measuresResponse

	err := json.Unmarshal(payload, &decoded)
	if err != nil {
		return measuresResponse{}, newExitError(
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

		return measuresResponse{}, newExitError(
			exitCodeAPI,
			fmt.Errorf("%w: %d: %s", errWithingsAPI, decoded.Status, message),
		)
	}

	return decoded, nil
}

func writeRawJSON(opts globalOptions, data any) error {
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

func buildMeasureRows(body measuresBody) []measureRow {
	location := measureLocation(body.Timezone)
	rows := make([]measureRow, defaultInt, len(body.MeasureGroups))

	for _, group := range body.MeasureGroups {
		timestamp := formatMeasureTime(group.Date, location)
		category := formatMeasureCategory(group.Category)

		for _, item := range group.Measures {
			typeID := strconv.Itoa(item.Type)
			rows = append(rows, measureRow{
				Time:     timestamp,
				Type:     formatMeasureType(typeID),
				Value:    formatScaledValue(item.Value, item.Unit),
				Unit:     formatMeasureUnit(typeID, item.Unit),
				Category: category,
			})
		}
	}

	return rows
}

func measureLocation(timezone string) *time.Location {
	if timezone == emptyString {
		return time.UTC
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}

	return location
}

func formatMeasureTime(epoch int64, location *time.Location) string {
	if epoch == defaultInt64 {
		return emptyString
	}

	return time.Unix(epoch, defaultInt64).In(location).Format(time.RFC3339)
}

func formatMeasureCategory(category int) string {
	switch strconv.Itoa(category) {
	case measureCategoryReal:
		return measureCategoryRealText
	case measureCategoryGoal:
		return measureCategoryGoalText
	default:
		return strconv.Itoa(category)
	}
}

func formatMeasureType(typeID string) string {
	if name, ok := measureTypeNameByID[typeID]; ok {
		return name
	}

	return typeID
}

func formatMeasureUnit(typeID string, unit int) string {
	if label, ok := measureUnitByTypeID[typeID]; ok {
		return label
	}

	return formatUnitPower(unit)
}

func formatUnitPower(unit int) string {
	if unit == defaultInt {
		return measureUnitBase
	}

	return measureUnitExponent + strconv.Itoa(unit)
}

func formatScaledValue(value int64, unit int) string {
	if unit == defaultInt {
		return strconv.FormatInt(value, measureNumberBase10)
	}

	scaled := value
	sign := emptyString

	if scaled < defaultInt64 {
		sign = measureNegativeSign
		scaled = -scaled
	}

	digits := strconv.FormatInt(scaled, measureNumberBase10)

	if unit > defaultInt {
		return sign + digits + strings.Repeat(measureZeroString, unit)
	}

	scale := -unit
	if len(digits) <= scale {
		digits = strings.Repeat(
			measureZeroString,
			scale-len(digits)+measureScalePad,
		) + digits
	}

	point := len(digits) - scale
	whole := digits[:point]
	frac := strings.TrimRight(digits[point:], measureZeroString)

	if frac == emptyString {
		return sign + whole
	}

	return sign + whole + measureDecimalSeparator + frac
}

func formatMeasureTable(rows []measureRow) (string, error) {
	var buffer bytes.Buffer

	writer := tabwriter.NewWriter(
		&buffer,
		measureTableMinWidth,
		measureTableTabWidth,
		measureTablePadding,
		measureTablePadChar,
		measureTableFlags,
	)
	_, _ = fmt.Fprintln(writer, "Time\tType\tValue\tUnit\tCategory")

	for _, row := range rows {
		_, _ = fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\n",
			row.Time,
			row.Type,
			row.Value,
			row.Unit,
			row.Category,
		)
	}

	err := writer.Flush()
	if err != nil {
		return emptyString, fmt.Errorf("render measures table: %w", err)
	}

	return strings.TrimRight(buffer.String(), "\n"), nil
}

func formatMeasureLines(rows []measureRow) []string {
	lines := make([]string, defaultInt, len(rows)+measureRowsHeaderCount)
	lines = append(lines, "time\ttype\tvalue\tunit\tcategory")

	for _, row := range rows {
		lines = append(lines, strings.Join([]string{
			row.Time,
			row.Type,
			row.Value,
			row.Unit,
			row.Category,
		}, "\t"))
	}

	return lines
}
