// Package activity handles Withings activity endpoints.
package activity

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

	"github.com/mreimbold/withings-cli/internal/app"
	"github.com/mreimbold/withings-cli/internal/errs"
	"github.com/mreimbold/withings-cli/internal/filters"
	"github.com/mreimbold/withings-cli/internal/output"
	"github.com/mreimbold/withings-cli/internal/params"
	"github.com/mreimbold/withings-cli/internal/withings"
)

const (
	serviceName     = "v2/measure"
	serviceShort    = "measure"
	serviceV2Suffix = "/v2"
	actionGet       = "getactivity"
	startDateParam  = "startdateymd"
	endDateParam    = "enddateymd"
	lastUpdateParam = "lastupdate"
	userIDParam     = "userid"
	limitParam      = "limit"
	offsetParam     = "offset"
	floatBitSize    = 64
	rowsHeaderCount = 1
	tableMinWidth   = 0
	tableTabWidth   = 0
	tablePadding    = 2
	tablePadChar    = ' '
	tableFlags      = 0
	tableHeader     = "Date\tSteps\tDistance\tCalories\t" +
		"Total Calories\tActive\tElevation\tSoft\tModerate\tIntense"
	plainHeader = "date\tsteps\tdistance\tcalories\t" +
		"total_calories\tactive\televation\tsoft\tmoderate\tintense"
	defaultInt  = 0
	emptyString = ""
)

// Options captures activity query parameters.
type Options struct {
	TimeRange  params.TimeRange
	Date       params.Date
	Pagination params.Pagination
	User       params.User
	LastUpdate params.LastUpdate
}

// Run fetches activity summaries and writes output.
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

	req, _, err := withings.BuildRequest(
		ctx,
		withings.APIBaseURL(appOpts.BaseURL, appOpts.Cloud),
		serviceForBase(withings.APIBaseURL(appOpts.BaseURL, appOpts.Cloud)),
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

type response struct {
	Status int    `json:"status"`
	Body   body   `json:"body"`
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

type body struct {
	Timezone   string `json:"timezone"`
	Activities []item `json:"activities"`
	More       bool   `json:"more"`
	Offset     int    `json:"offset"`
}

type item struct {
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

type row struct {
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
	rows := make([]row, defaultInt, len(body.Activities))

	for _, item := range body.Activities {
		rows = append(rows, row{
			Date:          item.Date,
			Steps:         formatFloat(item.Steps),
			Distance:      formatFloat(item.Distance),
			Calories:      formatFloat(item.Calories),
			TotalCalories: formatFloat(item.TotalCalories),
			Active:        formatFloat(item.Active),
			Elevation:     formatFloat(item.Elevation),
			Soft:          formatFloat(item.Soft),
			Moderate:      formatFloat(item.Moderate),
			Intense:       formatFloat(item.Intense),
		})
	}

	return rows
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, floatBitSize)
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

func formatLines(rows []row) []string {
	lines := make([]string, defaultInt, len(rows)+rowsHeaderCount)
	lines = append(lines, plainHeader)

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
