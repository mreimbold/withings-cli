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
	heartServiceName     = "heart"
	heartActionGet       = "get"
	heartStartDateParam  = "startdate"
	heartEndDateParam    = "enddate"
	heartLastUpdateParam = "lastupdate"
	heartUserIDParam     = "userid"
	heartLimitParam      = "limit"
	heartOffsetParam     = "offset"
	heartSignalParam     = "signal"
	heartSignalEnabled   = "1"
	heartNumberBase10    = 10
	heartRowsHeaderCount = 1
	heartTableMinWidth   = 0
	heartTableTabWidth   = 0
	heartTablePadding    = 2
	heartTablePadChar    = ' '
	heartTableFlags      = 0
	heartSignalYes       = "yes"
)

func runHeartGet(cmd *cobra.Command, opts heartGetOptions) error {
	globalOpts, err := readGlobalOptions(cmd)
	if err != nil {
		return err
	}

	ctx := cmd.Context()

	accessToken, err := ensureAccessToken(ctx, globalOpts)
	if err != nil {
		return err
	}

	params, err := buildHeartParams(opts)
	if err != nil {
		return newExitError(exitCodeUsage, err)
	}

	apiOpts := apiCallOptions{
		Service: heartServiceName,
		Action:  heartActionGet,
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

	return writeHeartResponse(globalOpts, payload)
}

func buildHeartParams(opts heartGetOptions) (url.Values, error) {
	values := url.Values{}

	err := applyHeartTimeFilters(&values, opts.TimeRange, opts.LastUpdate)
	if err != nil {
		return nil, err
	}

	applyHeartUser(&values, opts.User)
	applyHeartPagination(&values, opts.Pagination)

	if opts.Signal {
		values.Set(heartSignalParam, heartSignalEnabled)
	}

	return values, nil
}

func applyHeartTimeFilters(
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
			heartLastUpdateParam,
			strconv.FormatInt(lastUpdate.LastUpdate, heartNumberBase10),
		)
	}

	err := applyHeartTimeValue(
		values,
		heartStartDateParam,
		timeRange.Start,
		errInvalidStartTime,
	)
	if err != nil {
		return err
	}

	return applyHeartTimeValue(
		values,
		heartEndDateParam,
		timeRange.End,
		errInvalidEndTime,
	)
}

func applyHeartTimeValue(
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

	values.Set(key, strconv.FormatInt(epoch, heartNumberBase10))

	return nil
}

func applyHeartUser(values *url.Values, user userOption) {
	if user.UserID == emptyString {
		return
	}

	values.Set(heartUserIDParam, user.UserID)
}

func applyHeartPagination(values *url.Values, pagination paginationOptions) {
	if pagination.Limit > defaultInt {
		values.Set(heartLimitParam, strconv.Itoa(pagination.Limit))
	}

	if pagination.Offset > defaultInt {
		values.Set(heartOffsetParam, strconv.Itoa(pagination.Offset))
	}
}

type heartResponse struct {
	Status int       `json:"status"`
	Body   heartBody `json:"body"`
	Error  string    `json:"error"`
	Detail string    `json:"detail"`
}

type heartBody struct {
	Timezone string        `json:"timezone"`
	Series   []heartSeries `json:"series"`
}

//nolint:tagliatelle // Withings API uses snake_case JSON fields.
type heartSeries struct {
	ID        int64           `json:"id"`
	SignalID  int64           `json:"signalid"`
	StartDate int64           `json:"startdate"`
	EndDate   int64           `json:"enddate"`
	Timestamp int64           `json:"timestamp"`
	DeviceID  string          `json:"deviceid"`
	Model     int             `json:"model"`
	ECG       int             `json:"ecg"`
	AFib      int             `json:"afib"`
	HeartRate int             `json:"heart_rate"`
	Signal    json.RawMessage `json:"signal"`
}

type heartRow struct {
	Time      string
	HeartRate string
	Model     string
	Device    string
	SignalID  string
	ECG       string
	AFib      string
	Signal    string
}

func writeHeartResponse(opts globalOptions, payload []byte) error {
	decoded, err := decodeHeartResponse(payload)
	if err != nil {
		return err
	}

	if opts.Quiet {
		return nil
	}

	if opts.JSON {
		return writeRawJSON(opts, decoded.Body)
	}

	rows := buildHeartRows(decoded.Body)

	if opts.Plain {
		return writeLines(formatHeartLines(rows))
	}

	table, err := formatHeartTable(rows)
	if err != nil {
		return err
	}

	return writeLine(table)
}

func decodeHeartResponse(payload []byte) (heartResponse, error) {
	var decoded heartResponse

	err := json.Unmarshal(payload, &decoded)
	if err != nil {
		return heartResponse{}, newExitError(
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

		return heartResponse{}, newExitError(
			exitCodeAPI,
			fmt.Errorf("%w: %d: %s", errWithingsAPI, decoded.Status, message),
		)
	}

	return decoded, nil
}

func buildHeartRows(body heartBody) []heartRow {
	location := heartLocation(body.Timezone)
	rows := make([]heartRow, defaultInt, len(body.Series))

	for _, series := range body.Series {
		timestamp := formatHeartTime(heartSeriesTimestamp(series), location)
		rows = append(rows, heartRow{
			Time:      timestamp,
			HeartRate: formatHeartInt(series.HeartRate),
			Model:     formatHeartInt(series.Model),
			Device:    series.DeviceID,
			SignalID:  formatHeartInt64(heartSeriesSignalID(series)),
			ECG:       formatHeartInt(series.ECG),
			AFib:      formatHeartInt(series.AFib),
			Signal:    formatHeartSignal(series.Signal),
		})
	}

	return rows
}

func heartSeriesTimestamp(series heartSeries) int64 {
	switch {
	case series.StartDate != defaultInt64:
		return series.StartDate
	case series.Timestamp != defaultInt64:
		return series.Timestamp
	case series.EndDate != defaultInt64:
		return series.EndDate
	default:
		return defaultInt64
	}
}

func heartSeriesSignalID(series heartSeries) int64 {
	if series.SignalID != defaultInt64 {
		return series.SignalID
	}

	return series.ID
}

func heartLocation(timezone string) *time.Location {
	if timezone == emptyString {
		return time.UTC
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}

	return location
}

func formatHeartTime(epoch int64, location *time.Location) string {
	if epoch == defaultInt64 {
		return emptyString
	}

	return time.Unix(epoch, defaultInt64).In(location).Format(time.RFC3339)
}

func formatHeartInt(value int) string {
	if value == defaultInt {
		return emptyString
	}

	return strconv.Itoa(value)
}

func formatHeartInt64(value int64) string {
	if value == defaultInt64 {
		return emptyString
	}

	return strconv.FormatInt(value, heartNumberBase10)
}

func formatHeartSignal(signal json.RawMessage) string {
	trimmed := bytes.TrimSpace(signal)
	if len(trimmed) == defaultInt {
		return emptyString
	}

	if bytes.Equal(trimmed, []byte("null")) {
		return emptyString
	}

	return heartSignalYes
}

func formatHeartTable(rows []heartRow) (string, error) {
	var buffer bytes.Buffer

	writer := tabwriter.NewWriter(
		&buffer,
		heartTableMinWidth,
		heartTableTabWidth,
		heartTablePadding,
		heartTablePadChar,
		heartTableFlags,
	)
	_, _ = fmt.Fprintln(
		writer,
		"Time\tHeart Rate\tModel\tDevice\tSignal ID\tECG\tAFib\tSignal",
	)

	for _, row := range rows {
		_, _ = fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			row.Time,
			row.HeartRate,
			row.Model,
			row.Device,
			row.SignalID,
			row.ECG,
			row.AFib,
			row.Signal,
		)
	}

	err := writer.Flush()
	if err != nil {
		return emptyString, fmt.Errorf("render heart table: %w", err)
	}

	return strings.TrimRight(buffer.String(), "\n"), nil
}

func formatHeartLines(rows []heartRow) []string {
	lines := make([]string, defaultInt, len(rows)+heartRowsHeaderCount)
	lines = append(
		lines,
		"time\theart_rate\tmodel\tdevice\tsignal_id\tecg\tafib\tsignal",
	)

	for _, row := range rows {
		lines = append(lines, strings.Join([]string{
			row.Time,
			row.HeartRate,
			row.Model,
			row.Device,
			row.SignalID,
			row.ECG,
			row.AFib,
			row.Signal,
		}, "\t"))
	}

	return lines
}
