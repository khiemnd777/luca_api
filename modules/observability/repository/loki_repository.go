package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	moduleModel "github.com/khiemnd777/andy_api/modules/observability/model"
	sharedconfig "github.com/khiemnd777/andy_api/shared/config"
)

type ObservabilityRepository interface {
	ListLogs(ctx context.Context, query moduleModel.ListLogsQuery) ([]moduleModel.LogEntry, error)
	GetSummary(ctx context.Context, query moduleModel.ListLogsQuery) (moduleModel.Summary, error)
}

type observabilityRepository struct {
	client *http.Client
	cfg    sharedconfig.ObservabilityConfig
}

func NewObservabilityRepository() ObservabilityRepository {
	cfg := sharedconfig.Get().Observability
	timeout := cfg.Loki.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &observabilityRepository{
		client: &http.Client{Timeout: timeout},
		cfg:    cfg,
	}
}

func (r *observabilityRepository) ListLogs(ctx context.Context, query moduleModel.ListLogsQuery) ([]moduleModel.LogEntry, error) {
	end := query.To
	if end.IsZero() {
		end = time.Now()
	}
	start := query.From
	if start.IsZero() {
		start = end.Add(-1 * time.Hour)
	}

	params := url.Values{}
	params.Set("query", r.buildStreamQuery(query))
	params.Set("start", strconv.FormatInt(start.UnixNano(), 10))
	params.Set("end", strconv.FormatInt(end.UnixNano(), 10))
	params.Set("limit", strconv.Itoa(r.normalizeLimit(query.Limit)))
	params.Set("direction", normalizeDirection(query.Direction))

	var resp lokiRangeResponse
	if err := r.doRequest(ctx, "/loki/api/v1/query_range", params, &resp); err != nil {
		return nil, err
	}

	items := make([]moduleModel.LogEntry, 0, r.normalizeLimit(query.Limit))
	for _, stream := range resp.Data.Result {
		for _, value := range stream.Values {
			if len(value) < 2 {
				continue
			}
			entry := parseLogEntry(value[0], value[1], stream.Stream)
			items = append(items, entry)
		}
	}

	return items, nil
}

func (r *observabilityRepository) GetSummary(ctx context.Context, query moduleModel.ListLogsQuery) (moduleModel.Summary, error) {
	end := query.To
	if end.IsZero() {
		end = time.Now()
	}
	start := query.From
	if start.IsZero() {
		start = end.Add(-1 * time.Hour)
	}

	summary := moduleModel.Summary{
		From: start,
		To:   end,
		Counts: map[string]int{
			"warn":  0,
			"error": 0,
		},
	}

	for _, level := range []string{"warn", "error"} {
		value, err := r.queryCount(ctx, query, level, end, end.Sub(start))
		if err != nil {
			return moduleModel.Summary{}, err
		}
		summary.Counts[level] = value
	}

	return summary, nil
}

func (r *observabilityRepository) queryCount(ctx context.Context, query moduleModel.ListLogsQuery, level string, at time.Time, window time.Duration) (int, error) {
	if window <= 0 {
		window = time.Hour
	}

	params := url.Values{}
	params.Set("query", fmt.Sprintf(`sum(count_over_time((%s)[%s]))`, r.buildStreamQuery(withLevels(query, level)), formatRange(window)))
	params.Set("time", strconv.FormatInt(at.UnixNano(), 10))

	var resp lokiInstantResponse
	if err := r.doRequest(ctx, "/loki/api/v1/query", params, &resp); err != nil {
		return 0, err
	}

	if len(resp.Data.Result) == 0 || len(resp.Data.Result[0].Value) < 2 {
		return 0, nil
	}

	rawCount, _ := resp.Data.Result[0].Value[1].(string)
	count, err := strconv.ParseFloat(rawCount, 64)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *observabilityRepository) buildStreamQuery(query moduleModel.ListLogsQuery) string {
	selector := strings.TrimSpace(r.cfg.Loki.StreamSelector)
	if selector == "" {
		selector = `{app="luca_api"}`
	}

	parts := []string{selector}
	if query.Keyword != "" {
		parts = append(parts, fmt.Sprintf(`|= %s`, strconv.Quote(query.Keyword)))
	}

	parts = append(parts, "| json")
	parts = append(parts, `| __error__=""`)

	if len(query.Levels) > 0 {
		if len(query.Levels) == 1 {
			parts = append(parts, fmt.Sprintf(`| level=%s`, strconv.Quote(strings.ToLower(query.Levels[0]))))
		} else {
			levels := make([]string, 0, len(query.Levels))
			for _, level := range query.Levels {
				levels = append(levels, regexp.QuoteMeta(strings.ToLower(level)))
			}
			parts = append(parts, fmt.Sprintf(`| level=~%s`, strconv.Quote(strings.Join(levels, "|"))))
		}
	}
	if query.Module != "" {
		parts = append(parts, fmt.Sprintf(`| module=%s`, strconv.Quote(query.Module)))
	}
	if query.Service != "" {
		parts = append(parts, fmt.Sprintf(`| service=%s`, strconv.Quote(query.Service)))
	}
	if query.Env != "" {
		parts = append(parts, fmt.Sprintf(`| env=%s`, strconv.Quote(query.Env)))
	}
	if query.RequestID != "" {
		parts = append(parts, fmt.Sprintf(`| request_id=%s`, strconv.Quote(query.RequestID)))
	}
	if query.UserID != nil {
		parts = append(parts, fmt.Sprintf(`| user_id=%s`, strconv.Quote(strconv.Itoa(*query.UserID))))
	}
	if query.DeptID != nil {
		parts = append(parts, fmt.Sprintf(`| department_id=%s`, strconv.Quote(strconv.Itoa(*query.DeptID))))
	}

	return strings.Join(parts, " ")
}

func (r *observabilityRepository) doRequest(ctx context.Context, path string, params url.Values, out any) error {
	baseURL := strings.TrimRight(r.cfg.Loki.BaseURL, "/")
	if baseURL == "" {
		return fmt.Errorf("observability loki base_url is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}

	if r.cfg.Loki.TenantID != "" {
		req.Header.Set("X-Scope-OrgID", r.cfg.Loki.TenantID)
	}
	if r.cfg.Loki.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.cfg.Loki.BearerToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("loki request failed: %s", strings.TrimSpace(string(body)))
	}

	return json.Unmarshal(body, out)
}

func parseLogEntry(rawTS, rawLine string, stream map[string]string) moduleModel.LogEntry {
	entry := moduleModel.LogEntry{
		Raw: rawLine,
	}

	if parsedTS, err := strconv.ParseInt(rawTS, 10, 64); err == nil {
		entry.Timestamp = time.Unix(0, parsedTS).UTC()
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(rawLine), &payload); err != nil {
		entry.Message = rawLine
		entry.Level = "unknown"
		return entry
	}

	entry.Timestamp = parseTime(payload["ts"], entry.Timestamp)
	entry.Level = firstNonEmpty(toString(payload["level"]), toString(payload["severity"]))
	entry.Message = firstNonEmpty(toString(payload["message"]), toString(payload["msg"]))
	entry.Service = firstNonEmpty(toString(payload["service"]), stream["service"])
	entry.Module = firstNonEmpty(toString(payload["module"]), stream["module"])
	entry.Environment = firstNonEmpty(toString(payload["env"]), toString(payload["environment"]), stream["env"])
	entry.RequestID = firstNonEmpty(toString(payload["request_id"]), stream["request_id"])
	entry.UserID = firstNonZero(toInt(payload["user_id"]), toInt(stream["user_id"]))
	entry.DepartmentID = firstNonZero(toInt(payload["department_id"]), toInt(stream["department_id"]))
	entry.Method = firstNonEmpty(toString(payload["method"]), stream["method"])
	entry.Path = firstNonEmpty(toString(payload["path"]), stream["path"])
	entry.Source = firstNonEmpty(toString(payload["source"]), stream["source"])
	entry.Error = firstNonEmpty(toString(payload["error"]), stream["error"])
	entry.Stacktrace = firstNonEmpty(toString(payload["stacktrace"]), stream["stacktrace"])
	entry.Fields = extractExtraFields(payload)

	return entry
}

func extractExtraFields(payload map[string]any) map[string]any {
	known := map[string]struct{}{
		"ts":            {},
		"level":         {},
		"severity":      {},
		"message":       {},
		"msg":           {},
		"service":       {},
		"module":        {},
		"env":           {},
		"environment":   {},
		"request_id":    {},
		"user_id":       {},
		"department_id": {},
		"method":        {},
		"path":          {},
		"source":        {},
		"error":         {},
		"stacktrace":    {},
	}

	extra := make(map[string]any)
	for k, v := range payload {
		if _, ok := known[k]; ok {
			continue
		}
		extra[k] = v
	}
	if len(extra) == 0 {
		return nil
	}
	return extra
}

func withLevels(query moduleModel.ListLogsQuery, levels ...string) moduleModel.ListLogsQuery {
	query.Levels = levels
	return query
}

func normalizeDirection(direction string) string {
	switch strings.ToUpper(strings.TrimSpace(direction)) {
	case "FORWARD":
		return "FORWARD"
	default:
		return "BACKWARD"
	}
}

func formatRange(window time.Duration) string {
	seconds := int(window.Seconds())
	if seconds <= 0 {
		seconds = 1
	}
	return fmt.Sprintf("%ds", seconds)
}

func parseTime(value any, fallback time.Time) time.Time {
	raw := toString(value)
	if raw == "" {
		return fallback
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return fallback
	}
	return parsed.UTC()
}

func toString(value any) string {
	if value == nil {
		return ""
	}
	switch vv := value.(type) {
	case string:
		return vv
	default:
		return fmt.Sprintf("%v", value)
	}
}

func toInt(value any) int {
	switch vv := value.(type) {
	case int:
		return vv
	case int8:
		return int(vv)
	case int16:
		return int(vv)
	case int32:
		return int(vv)
	case int64:
		return int(vv)
	case float64:
		return int(vv)
	case float32:
		return int(vv)
	case json.Number:
		n, _ := vv.Int64()
		return int(n)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(vv))
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func (r *observabilityRepository) normalizeLimit(limit int) int {
	maxLimit := r.cfg.Loki.MaxQueryLimit
	if maxLimit <= 0 {
		maxLimit = 200
	}
	if limit <= 0 {
		return 50
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

type lokiRangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type lokiInstantResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []any             `json:"value"`
		} `json:"result"`
	} `json:"data"`
}
