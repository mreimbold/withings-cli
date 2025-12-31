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
	"time"

	"github.com/spf13/cobra"
)

const (
	sleepServiceName     = "v2/sleep"
	sleepServiceShort    = "sleep"
	sleepServiceV2Suffix = "/v2"
	sleepActionGet       = "getsummary"
	sleepStartDateParam  = "startdateymd"
	sleepEndDateParam    = "enddateymd"
	sleepLastUpdateParam = "lastupdate"
	sleepUserIDParam     = "userid"
	sleepModelParam      = "model"
	sleepLimitParam      = "limit"
	sleepOffsetParam     = "offset"
	sleepNumberBase10    = 10
	sleepRowsHeaderCount = 1
	sleepTableMinWidth   = 0
	sleepTableTabWidth   = 0
	sleepTablePadding    = 2
	sleepTablePadChar    = ' '
	sleepTableFlags      = 0
	sleepTableHeader     = "Start\tEnd\tDuration\tScore\tWakeups\tModel"
	sleepPlainHeader     = "start\tend\tduration\tscore\twakeups\tmodel"
)

func runSleepGet(cmd *cobra.Command, opts sleepGetOptions) error {
	globalOpts, err := readGlobalOptions(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	accessToken, err := ensureAccessToken(ctx, globalOpts)
	if err != nil {
		return err
	}

	params, err := buildSleepParams(opts)
	if err != nil {
		return newExitError(exitCodeUsage, err)
	}

	apiOpts := apiCallOptions{
		Service: sleepServiceForBase(
			apiBaseURL(globalOpts.BaseURL, globalOpts.Cloud),
		),
		Action: sleepActionGet,
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

	return writeSleepResponse(globalOpts, payload)
}

func sleepServiceForBase(baseURL string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, sleepServiceV2Suffix) {
		return sleepServiceShort
	}

	return sleepServiceName
}

func buildSleepParams(opts sleepGetOptions) (url.Values, error) {
	values := url.Values{}

	err := applySleepTimeFilters(
		&values,
		opts.Date,
		opts.TimeRange,
		opts.LastUpdate,
	)
	if err != nil {
		return nil, err
	}

	applySleepUser(&values, opts.User)
	applySleepPagination(&values, opts.Pagination)
	applySleepModel(&values, opts.Model)

	return values, nil
}

func applySleepTimeFilters(
	values *url.Values,
	date dateOption,
	timeRange timeRangeOptions,
	lastUpdate lastUpdateOption,
) error {
	err := applyLastUpdateFilter(
		values,
		sleepLastUpdateParam,
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
		sleepStartDateParam,
		sleepEndDateParam,
		dateRange,
	)

	return nil
}

func applySleepUser(values *url.Values, user userOption) {
	if user.UserID == emptyString {
		return
	}

	values.Set(sleepUserIDParam, user.UserID)
}

func applySleepPagination(values *url.Values, pagination paginationOptions) {
	if pagination.Limit > defaultInt {
		values.Set(sleepLimitParam, strconv.Itoa(pagination.Limit))
	}

	if pagination.Offset > defaultInt {
		values.Set(sleepOffsetParam, strconv.Itoa(pagination.Offset))
	}
}

func applySleepModel(values *url.Values, model int) {
	if model <= defaultInt {
		return
	}

	values.Set(sleepModelParam, strconv.Itoa(model))
}

type sleepResponse struct {
	Status int       `json:"status"`
	Body   sleepBody `json:"body"`
	Error  string    `json:"error"`
	Detail string    `json:"detail"`
}

type sleepBody struct {
	Timezone string        `json:"timezone"`
	Series   []sleepSeries `json:"series"`
	More     bool          `json:"more"`
	Offset   int           `json:"offset"`
}

//nolint:tagliatelle // Withings API uses snake_case JSON fields.
type sleepSeries struct {
	Date      string `json:"date"`
	StartDate int64  `json:"startdate"`
	EndDate   int64  `json:"enddate"`
	Duration  int64  `json:"duration"`
	Score     int    `json:"sleep_score"`
	Wakeups   int    `json:"wakeupcount"`
	Model     int    `json:"model"`
}

type sleepRow struct {
	Start    string
	End      string
	Duration string
	Score    string
	Wakeups  string
	Model    string
}

func writeSleepResponse(opts globalOptions, payload []byte) error {
	decoded, err := decodeSleepResponse(payload)
	if err != nil {
		return err
	}

	if opts.Quiet {
		return nil
	}

	if opts.JSON {
		return writeRawJSON(opts, decoded.Body)
	}

	rows := buildSleepRows(decoded.Body)

	if opts.Plain {
		return writeLines(formatSleepLines(rows))
	}

	table, err := formatSleepTable(rows)
	if err != nil {
		return err
	}

	return writeLine(table)
}

func decodeSleepResponse(payload []byte) (sleepResponse, error) {
	var decoded sleepResponse

	err := json.Unmarshal(payload, &decoded)
	if err != nil {
		return sleepResponse{}, newExitError(
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

		return sleepResponse{}, newExitError(
			exitCodeAPI,
			fmt.Errorf("%w: %d: %s", errWithingsAPI, decoded.Status, message),
		)
	}

	return decoded, nil
}

func buildSleepRows(body sleepBody) []sleepRow {
	location := sleepLocation(body.Timezone)
	rows := make([]sleepRow, defaultInt, len(body.Series))

	for _, series := range body.Series {
		rows = append(rows, sleepRow{
			Start:    formatSleepStart(series, location),
			End:      formatSleepEnd(series, location),
			Duration: formatSleepInt64(series.Duration),
			Score:    formatSleepInt(series.Score),
			Wakeups:  formatSleepInt(series.Wakeups),
			Model:    formatSleepInt(series.Model),
		})
	}

	return rows
}

func sleepLocation(timezone string) *time.Location {
	if timezone == emptyString {
		return time.UTC
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}

	return location
}

func formatSleepStart(series sleepSeries, location *time.Location) string {
	if series.StartDate != defaultInt64 {
		return formatSleepTime(series.StartDate, location)
	}

	return series.Date
}

func formatSleepEnd(series sleepSeries, location *time.Location) string {
	if series.EndDate != defaultInt64 {
		return formatSleepTime(series.EndDate, location)
	}

	return emptyString
}

func formatSleepTime(epoch int64, location *time.Location) string {
	if epoch == defaultInt64 {
		return emptyString
	}

	return time.Unix(epoch, defaultInt64).In(location).Format(time.RFC3339)
}

func formatSleepInt(value int) string {
	return strconv.Itoa(value)
}

func formatSleepInt64(value int64) string {
	return strconv.FormatInt(value, sleepNumberBase10)
}

func formatSleepTable(rows []sleepRow) (string, error) {
	var buffer bytes.Buffer

	writer := tabwriter.NewWriter(
		&buffer,
		sleepTableMinWidth,
		sleepTableTabWidth,
		sleepTablePadding,
		sleepTablePadChar,
		sleepTableFlags,
	)
	_, _ = fmt.Fprintln(writer, sleepTableHeader)

	for _, row := range rows {
		_, _ = fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			row.Start,
			row.End,
			row.Duration,
			row.Score,
			row.Wakeups,
			row.Model,
		)
	}

	err := writer.Flush()
	if err != nil {
		return emptyString, fmt.Errorf("render sleep table: %w", err)
	}

	return strings.TrimRight(buffer.String(), "\n"), nil
}

func formatSleepLines(rows []sleepRow) []string {
	lines := make([]string, defaultInt, len(rows)+sleepRowsHeaderCount)
	lines = append(lines, sleepPlainHeader)

	for _, row := range rows {
		lines = append(lines, strings.Join([]string{
			row.Start,
			row.End,
			row.Duration,
			row.Score,
			row.Wakeups,
			row.Model,
		}, "\t"))
	}

	return lines
}
