package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const (
	activityServiceName     = "v2/measure"
	activityServiceShort    = "measure"
	activityServiceV2Suffix = "/v2"
	activityActionGet       = "getactivity"
	activityStartDateParam  = "startdateymd"
	activityEndDateParam    = "enddateymd"
	activityLastUpdateParam = "lastupdate"
	activityUserIDParam     = "userid"
	activityLimitParam      = "limit"
	activityOffsetParam     = "offset"
	activityFloatBitSize    = 64
	activityRowsHeaderCount = 1
	activityTableMinWidth   = 0
	activityTableTabWidth   = 0
	activityTablePadding    = 2
	activityTablePadChar    = ' '
	activityTableFlags      = 0
	activityTableHeader     = "Date\tSteps\tDistance\tCalories\t" +
		"Total Calories\tActive\tElevation\tSoft\tModerate\tIntense"
	activityPlainHeader = "date\tsteps\tdistance\tcalories\t" +
		"total_calories\tactive\televation\tsoft\tmoderate\tintense"
)

func runActivityGet(cmd *cobra.Command, opts activityGetOptions) error {
	globalOpts, err := readGlobalOptions(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	accessToken, err := ensureAccessToken(ctx, globalOpts)
	if err != nil {
		return err
	}

	params, err := buildActivityParams(opts)
	if err != nil {
		return newExitError(exitCodeUsage, err)
	}

	apiOpts := apiCallOptions{
		Service: activityServiceForBase(
			apiBaseURL(globalOpts.BaseURL, globalOpts.Cloud),
		),
		Action: activityActionGet,
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

	return writeActivityResponse(globalOpts, payload)
}

func activityServiceForBase(baseURL string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, activityServiceV2Suffix) {
		return activityServiceShort
	}

	return activityServiceName
}

func buildActivityParams(opts activityGetOptions) (url.Values, error) {
	values := url.Values{}

	err := applyActivityTimeFilters(
		&values,
		opts.Date,
		opts.TimeRange,
		opts.LastUpdate,
	)
	if err != nil {
		return nil, err
	}

	applyActivityUser(&values, opts.User)
	applyActivityPagination(&values, opts.Pagination)

	return values, nil
}

func applyActivityTimeFilters(
	values *url.Values,
	date dateOption,
	timeRange timeRangeOptions,
	lastUpdate lastUpdateOption,
) error {
	err := applyLastUpdateFilter(
		values,
		activityLastUpdateParam,
		lastUpdate,
		date,
		timeRange,
	)
	if err != nil {
		return err
	}

	dateRange, err := resolveDateRange(
		date,
		timeRange,
		errInvalidStartTime,
		errInvalidEndTime,
	)
	if err != nil {
		return err
	}

	applyDateRangeParams(
		values,
		activityStartDateParam,
		activityEndDateParam,
		dateRange,
	)

	return nil
}

func applyActivityUser(values *url.Values, user userOption) {
	if user.UserID == emptyString {
		return
	}

	values.Set(activityUserIDParam, user.UserID)
}

func applyActivityPagination(values *url.Values, pagination paginationOptions) {
	if pagination.Limit > defaultInt {
		values.Set(activityLimitParam, strconv.Itoa(pagination.Limit))
	}

	if pagination.Offset > defaultInt {
		values.Set(activityOffsetParam, strconv.Itoa(pagination.Offset))
	}
}

type activityResponse struct {
	Status int          `json:"status"`
	Body   activityBody `json:"body"`
	Error  string       `json:"error"`
	Detail string       `json:"detail"`
}

type activityBody struct {
	Timezone   string         `json:"timezone"`
	Activities []activityItem `json:"activities"`
	More       bool           `json:"more"`
	Offset     int            `json:"offset"`
}

type activityItem struct {
	Date          string  `json:"date"`
	Steps         float64 `json:"steps"`
	Distance      float64 `json:"distance"`
	Calories      float64 `json:"calories"`
	TotalCalories float64 `json:"totalcalories"`
	Active        float64 `json:"active"`
	Elevation     float64 `json:"elevation"`
	Soft          float64 `json:"soft"`
	Moderate      float64 `json:"moderate"`
	Intense       float64 `json:"intense"`
}

type activityRow struct {
	Date          string
	Steps         string
	Distance      string
	Calories      string
	TotalCalories string
	Active        string
	Elevation     string
	Soft          string
	Moderate      string
	Intense       string
}

func writeActivityResponse(opts globalOptions, payload []byte) error {
	decoded, err := decodeActivityResponse(payload)
	if err != nil {
		return err
	}

	if opts.Quiet {
		return nil
	}

	if opts.JSON {
		return writeRawJSON(opts, decoded.Body)
	}

	rows := buildActivityRows(decoded.Body)

	if opts.Plain {
		return writeLines(formatActivityLines(rows))
	}

	table, err := formatActivityTable(rows)
	if err != nil {
		return err
	}

	return writeLine(table)
}

func decodeActivityResponse(payload []byte) (activityResponse, error) {
	var decoded activityResponse

	err := json.Unmarshal(payload, &decoded)
	if err != nil {
		return activityResponse{}, newExitError(
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

		return activityResponse{}, newExitError(
			exitCodeAPI,
			fmt.Errorf("%w: %d: %s", errWithingsAPI, decoded.Status, message),
		)
	}

	return decoded, nil
}

func buildActivityRows(body activityBody) []activityRow {
	rows := make([]activityRow, defaultInt, len(body.Activities))

	for _, item := range body.Activities {
		rows = append(rows, activityRow{
			Date:          item.Date,
			Steps:         formatActivityFloat(item.Steps),
			Distance:      formatActivityFloat(item.Distance),
			Calories:      formatActivityFloat(item.Calories),
			TotalCalories: formatActivityFloat(item.TotalCalories),
			Active:        formatActivityFloat(item.Active),
			Elevation:     formatActivityFloat(item.Elevation),
			Soft:          formatActivityFloat(item.Soft),
			Moderate:      formatActivityFloat(item.Moderate),
			Intense:       formatActivityFloat(item.Intense),
		})
	}

	return rows
}

func formatActivityFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, activityFloatBitSize)
}

func formatActivityTable(rows []activityRow) (string, error) {
	var buffer bytes.Buffer

	writer := tabwriter.NewWriter(
		&buffer,
		activityTableMinWidth,
		activityTableTabWidth,
		activityTablePadding,
		activityTablePadChar,
		activityTableFlags,
	)
	_, _ = fmt.Fprintln(writer, activityTableHeader)

	for _, row := range rows {
		_, _ = fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.Date,
			row.Steps,
			row.Distance,
			row.Calories,
			row.TotalCalories,
			row.Active,
			row.Elevation,
			row.Soft,
			row.Moderate,
			row.Intense,
		)
	}

	err := writer.Flush()
	if err != nil {
		return emptyString, fmt.Errorf("render activity table: %w", err)
	}

	return strings.TrimRight(buffer.String(), "\n"), nil
}

func formatActivityLines(rows []activityRow) []string {
	lines := make([]string, defaultInt, len(rows)+activityRowsHeaderCount)
	lines = append(lines, activityPlainHeader)

	for _, row := range rows {
		lines = append(lines, strings.Join([]string{
			row.Date,
			row.Steps,
			row.Distance,
			row.Calories,
			row.TotalCalories,
			row.Active,
			row.Elevation,
			row.Soft,
			row.Moderate,
			row.Intense,
		}, "\t"))
	}

	return lines
}
