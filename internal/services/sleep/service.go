// Package sleep handles Withings sleep endpoints.
package sleep

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
	serviceName     = "v2/sleep"
	serviceShort    = "sleep"
	serviceV2Suffix = "/v2"
	actionGet       = "getsummary"
	startDateParam  = "startdateymd"
	endDateParam    = "enddateymd"
	lastUpdateParam = "lastupdate"
	userIDParam     = "userid"
	modelParam      = "model"
	limitParam      = "limit"
	offsetParam     = "offset"
	numberBase10    = 10
	rowsHeaderCount = 1
	tableMinWidth   = 0
	tableTabWidth   = 0
	tablePadding    = 2
	tablePadChar    = ' '
	tableFlags      = 0
	tableHeader     = "Start\tEnd\tDuration\tScore\tWakeups\tModel"
	plainHeader     = "start\tend\tduration\tscore\twakeups\tmodel"
	defaultInt      = 0
	defaultInt64    = 0
	emptyString     = ""
)

// Options captures sleep query parameters.
type Options struct {
	TimeRange  params.TimeRange
	Date       params.Date
	Pagination params.Pagination
	User       params.User
	LastUpdate params.LastUpdate
	Model      int
}

// Run fetches sleep summaries and writes output.
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
		actionGet,
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

	err := applyTimeFilters(&values, opts.Date, opts.TimeRange, opts.LastUpdate)
	if err != nil {
		return nil, err
	}

	applyUser(&values, opts.User)
	applyPagination(&values, opts.Pagination)
	applyModel(&values, opts.Model)

	return values, nil
}

func applyTimeFilters(
	values *url.Values,
	date params.Date,
	timeRange params.TimeRange,
	lastUpdate params.LastUpdate,
) error {
	err := filters.ApplyLastUpdateFilter(
		values,
		lastUpdateParam,
		lastUpdate,
		date,
		timeRange,
		errs.ErrInvalidLastUpdate,
		errs.ErrLastUpdateConflict,
	)
	if err != nil {
		return fmt.Errorf("apply last-update filter: %w", err)
	}

	dateRange, err := filters.ResolveDateRange(
		date,
		timeRange,
		errs.ErrInvalidStartTime,
		errs.ErrInvalidEndTime,
	)
	if err != nil {
		return fmt.Errorf("resolve date range: %w", err)
	}

	filters.ApplyDateRangeParams(
		values,
		startDateParam,
		endDateParam,
		dateRange,
	)

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

func applyModel(values *url.Values, model int) {
	if model <= defaultInt {
		return
	}

	values.Set(modelParam, strconv.Itoa(model))
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
	More     bool     `json:"more"`
	Offset   int      `json:"offset"`
}

//nolint:tagliatelle // Withings API uses snake_case JSON fields.
type series struct {
	Date      string `json:"date"`
	StartDate int64  `json:"startdate"`
	EndDate   int64  `json:"enddate"`
	Duration  int64  `json:"duration"`
	Score     int    `json:"sleep_score"`
	Wakeups   int    `json:"wakeupcount"`
	Model     int    `json:"model"`
}

type row struct {
	Start    string
	End      string
	Duration string
	Score    string
	Wakeups  string
	Model    string
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
	location := sleepLocation(body.Timezone)
	rows := make([]row, defaultInt, len(body.Series))

	for _, series := range body.Series {
		rows = append(rows, row{
			Start:    formatStart(series, location),
			End:      formatEnd(series, location),
			Duration: formatInt64(series.Duration),
			Score:    formatInt(series.Score),
			Wakeups:  formatInt(series.Wakeups),
			Model:    formatInt(series.Model),
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

func formatStart(series series, location *time.Location) string {
	if series.StartDate != defaultInt64 {
		return formatTime(series.StartDate, location)
	}

	return series.Date
}

func formatEnd(series series, location *time.Location) string {
	if series.EndDate != defaultInt64 {
		return formatTime(series.EndDate, location)
	}

	return emptyString
}

func formatTime(epoch int64, location *time.Location) string {
	if epoch == defaultInt64 {
		return emptyString
	}

	return time.Unix(epoch, defaultInt64).In(location).Format(time.RFC3339)
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}

func formatInt64(value int64) string {
	return strconv.FormatInt(value, numberBase10)
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
	_, _ = fmt.Fprintln(writer, tableHeader)

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

func formatLines(rows []row) []string {
	lines := make([]string, defaultInt, len(rows)+rowsHeaderCount)
	lines = append(lines, plainHeader)

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
