// Package heart handles Withings heart endpoints.
package heart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/mreimbold/withings-cli/internal/app"
	"github.com/mreimbold/withings-cli/internal/errs"
	"github.com/mreimbold/withings-cli/internal/filters"
	"github.com/mreimbold/withings-cli/internal/output"
	"github.com/mreimbold/withings-cli/internal/params"
	"github.com/mreimbold/withings-cli/internal/withings"
)

const (
	serviceName     = "v2/heart"
	serviceShort    = "heart"
	serviceV2Suffix = "/v2"
	actionList      = "list"
	startDateParam  = "startdate"
	endDateParam    = "enddate"
	lastUpdateParam = "lastupdate"
	userIDParam     = "userid"
	limitParam      = "limit"
	offsetParam     = "offset"
	signalParam     = "signal"
	signalEnabled   = "1"
	numberBase10    = 10
	rowsHeaderCount = 1
	tableMinWidth   = 0
	tableTabWidth   = 0
	tablePadding    = 2
	tablePadChar    = ' '
	tableFlags      = 0
	defaultInt      = 0
	defaultInt64    = 0
	signalYes       = "yes"
	emptyString     = ""
)

// Options captures heart query parameters.
type Options struct {
	TimeRange  params.TimeRange
	Pagination params.Pagination
	User       params.User
	LastUpdate params.LastUpdate
	Signal     bool
}

// Run fetches heart data and writes output.
func Run(
	ctx context.Context,
	opts Options,
	appOpts app.Options,
	accessToken string,
) error {
	values, err := buildParams(opts)
	if err != nil {
		return app.NewExitError(app.ExitCodeUsage, err)
	}

	baseURL := withings.APIBaseURL(appOpts.BaseURL, appOpts.Cloud)
	service := serviceForBase(baseURL)

	req, _, err := withings.BuildRequest(
		ctx,
		baseURL,
		service,
		actionList,
		accessToken,
		values,
	)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	//nolint:bodyclose // ReadPayload closes the response body.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return app.NewExitError(app.ExitCodeNetwork, err)
	}

	payload, err := withings.ReadPayload(resp)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	return writeResponse(appOpts, payload)
}

func serviceForBase(baseURL string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, serviceV2Suffix) {
		return serviceShort
	}

	return serviceName
}

func buildParams(opts Options) (url.Values, error) {
	values := url.Values{}

	err := applyTimeFilters(&values, opts.TimeRange, opts.LastUpdate)
	if err != nil {
		return nil, err
	}

	applyUser(&values, opts.User)
	applyPagination(&values, opts.Pagination)

	if opts.Signal {
		values.Set(signalParam, signalEnabled)
	}

	return values, nil
}

func applyTimeFilters(
	values *url.Values,
	timeRange params.TimeRange,
	lastUpdate params.LastUpdate,
) error {
	err := filters.ApplyLastUpdateFilter(
		values,
		lastUpdateParam,
		lastUpdate,
		params.Date{Date: emptyString},
		timeRange,
		errs.ErrInvalidLastUpdate,
		errs.ErrLastUpdateConflict,
	)
	if err != nil {
		return fmt.Errorf("apply last-update filter: %w", err)
	}

	err = applyTimeValue(
		values,
		startDateParam,
		timeRange.Start,
		errs.ErrInvalidStartTime,
	)
	if err != nil {
		return err
	}

	return applyTimeValue(
		values,
		endDateParam,
		timeRange.End,
		errs.ErrInvalidEndTime,
	)
}

func applyTimeValue(
	values *url.Values,
	key string,
	raw string,
	errInvalid error,
) error {
	if raw == emptyString {
		return nil
	}

	epoch, err := filters.ParseEpoch(raw)
	if err != nil {
		return fmt.Errorf("%w: %w", errInvalid, err)
	}

	values.Set(key, strconv.FormatInt(epoch, numberBase10))

	return nil
}

func applyUser(values *url.Values, user params.User) {
	if user.UserID == emptyString {
		return
	}

	values.Set(userIDParam, user.UserID)
}

func applyPagination(values *url.Values, pagination params.Pagination) {
	if pagination.Limit > defaultInt {
		values.Set(limitParam, strconv.Itoa(pagination.Limit))
	}

	if pagination.Offset > defaultInt {
		values.Set(offsetParam, strconv.Itoa(pagination.Offset))
	}
}

type response struct {
	Status int    `json:"status"`
	Body   body   `json:"body"`
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

type body struct {
	Timezone string   `json:"timezone"`
	Series   []series `json:"series"`
}

type series struct {
	ID        int64  `json:"id"`
	SignalID  int64  `json:"signalid"`
	StartDate int64  `json:"startdate"`
	EndDate   int64  `json:"enddate"`
	Timestamp int64  `json:"timestamp"`
	DeviceID  string `json:"deviceid"`
	Model     int    `json:"model"`
	ECG       int    `json:"ecg"`
	AFib      int    `json:"afib"`
	//nolint:tagliatelle // Withings API uses snake_case JSON fields.
	HeartRate int             `json:"heart_rate"`
	Signal    json.RawMessage `json:"signal"`
}

type row struct {
	Time      string
	HeartRate string
	Model     string
	Device    string
	SignalID  string
	ECG       string
	AFib      string
	Signal    string
}

func writeResponse(opts app.Options, payload []byte) error {
	decoded, err := decodeResponse(payload)
	if err != nil {
		return err
	}

	return writeBody(opts, decoded.Body)
}

func writeBody(opts app.Options, body body) error {
	if opts.Quiet {
		return nil
	}

	if opts.JSON {
		return writeJSONOutput(opts, body)
	}

	rows := buildRows(body)

	if opts.Plain {
		return writePlainOutput(rows)
	}

	return writeTableOutput(rows)
}

func writeJSONOutput(opts app.Options, body body) error {
	err := output.WriteRawJSON(opts, body)
	if err != nil {
		return fmt.Errorf("write json output: %w", err)
	}

	return nil
}

func writePlainOutput(rows []row) error {
	err := output.WriteLines(formatLines(rows))
	if err != nil {
		return fmt.Errorf("write plain output: %w", err)
	}

	return nil
}

func writeTableOutput(rows []row) error {
	table, err := formatTable(rows)
	if err != nil {
		return err
	}

	err = output.WriteLine(table)
	if err != nil {
		return fmt.Errorf("write table output: %w", err)
	}

	return nil
}

func decodeResponse(payload []byte) (response, error) {
	var decoded response

	err := json.Unmarshal(payload, &decoded)
	if err != nil {
		return response{}, app.NewExitError(
			app.ExitCodeFailure,
			fmt.Errorf("decode api response: %w", err),
		)
	}

	if decoded.Status != withings.StatusOK {
		message := decoded.Error
		if message == emptyString {
			message = decoded.Detail
		}

		if message == emptyString {
			message = strings.TrimSpace(string(payload))
		}

		return response{}, app.NewExitError(
			app.ExitCodeAPI,
			fmt.Errorf("%w: %d: %s", withings.ErrAPI, decoded.Status, message),
		)
	}

	return decoded, nil
}

func buildRows(body body) []row {
	location := seriesLocation(body.Timezone)
	rows := make([]row, defaultInt, len(body.Series))

	for _, series := range body.Series {
		timestamp := formatTime(seriesTimestamp(series), location)
		rows = append(rows, row{
			Time:      timestamp,
			HeartRate: formatInt(series.HeartRate),
			Model:     formatInt(series.Model),
			Device:    series.DeviceID,
			SignalID:  formatInt64(seriesSignalID(series)),
			ECG:       formatInt(series.ECG),
			AFib:      formatInt(series.AFib),
			Signal:    formatSignal(series.Signal),
		})
	}

	return rows
}

func seriesTimestamp(series series) int64 {
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

func seriesSignalID(series series) int64 {
	if series.SignalID != defaultInt64 {
		return series.SignalID
	}

	return series.ID
}

func seriesLocation(timezone string) *time.Location {
	if timezone == emptyString {
		return time.UTC
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}

	return location
}

func formatTime(epoch int64, location *time.Location) string {
	if epoch == defaultInt64 {
		return emptyString
	}

	return time.Unix(epoch, defaultInt64).In(location).Format(time.RFC3339)
}

func formatInt(value int) string {
	if value == defaultInt {
		return emptyString
	}

	return strconv.Itoa(value)
}

func formatInt64(value int64) string {
	if value == defaultInt64 {
		return emptyString
	}

	return strconv.FormatInt(value, numberBase10)
}

func formatSignal(signal json.RawMessage) string {
	trimmed := bytes.TrimSpace(signal)
	if len(trimmed) == defaultInt {
		return emptyString
	}

	if bytes.Equal(trimmed, []byte("null")) {
		return emptyString
	}

	return signalYes
}

func formatTable(rows []row) (string, error) {
	var buffer bytes.Buffer

	writer := tabwriter.NewWriter(
		&buffer,
		tableMinWidth,
		tableTabWidth,
		tablePadding,
		tablePadChar,
		tableFlags,
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

func formatLines(rows []row) []string {
	lines := make([]string, defaultInt, len(rows)+rowsHeaderCount)
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
