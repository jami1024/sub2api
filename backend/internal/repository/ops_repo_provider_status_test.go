package repository

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestQueryProviderStatusTimelineUsesSetBasedAggregation(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		if strings.Contains(strings.ToUpper(actualSQL), "LEFT JOIN LATERAL") {
			t.Fatalf("timeline query must not use per-bucket LATERAL probes")
		}
		return nil
	})))
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := &opsRepository{db: db}
	start := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	mock.ExpectQuery("provider timeline should be set based").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider",
			"bucket_start",
			"request_count",
			"success_count",
			"failure_count",
			"p50_ms",
			"p95_ms",
			"p99_ms",
			"duration_avg_ms",
			"ttft_avg_ms",
			"ttft_sample_count",
			"upstream_ttft_avg_ms",
			"gateway_ttft_avg_ms",
			"timeout_524_count",
			"timeout_524_avg_ms",
		}).AddRow("openai", start, int64(1), int64(1), int64(0), float64(100), float64(100), float64(100), float64(100), float64(80), int64(1), float64(30), float64(50), int64(0), nil))

	points, err := repo.queryProviderStatusTimeline(context.Background(), start, end, 300, []string{"openai"})
	if err != nil {
		t.Fatalf("queryProviderStatusTimeline returned error: %v", err)
	}
	if len(points) != 1 || points[0].Provider != "openai" {
		t.Fatalf("unexpected points: %#v", points)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestQueryProviderStatusUsesAccountNameAsProvider(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		if !strings.Contains(actualSQL, "NULLIF(a.name, '')") {
			t.Fatalf("provider status query must prefer accounts.name as provider, sql=%s", actualSQL)
		}
		if strings.Contains(actualSQL, "COALESCE(NULLIF(g.platform, ''), NULLIF(a.platform, ''), 'unknown') AS provider") {
			t.Fatalf("provider status query must not group all accounts by platform")
		}
		return nil
	})))
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := &opsRepository{db: db}
	start := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	mock.ExpectQuery("provider summary should use account name").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider",
			"request_count",
			"success_count",
			"failure_count",
			"business_limited_count",
			"cache_read_rate",
			"p50_ms",
			"p95_ms",
			"p99_ms",
			"duration_avg_ms",
			"duration_max_ms",
			"ttft_avg_ms",
			"ttft_p95_ms",
			"ttft_sample_count",
			"upstream_ttft_avg_ms",
			"gateway_ttft_avg_ms",
			"timeout_524_count",
			"timeout_524_avg_ms",
			"last_seen",
		}).AddRow("gzw plus", int64(80), int64(80), int64(0), int64(0), nil, float64(120), float64(300), float64(500), float64(250), float64(500), float64(180), float64(300), int64(10), nil, nil, int64(0), nil, end))

	items, err := repo.queryProviderStatusSummary(context.Background(), start, end, 50)
	if err != nil {
		t.Fatalf("queryProviderStatusSummary returned error: %v", err)
	}
	if len(items) != 1 || items[0].Provider != "gzw plus" {
		t.Fatalf("provider = %#v, want gzw plus", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestQueryProviderStatusExcludesUnlinkedErrors(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		if strings.Contains(actualSQL, "NULLIF(oel.platform, '')") {
			t.Fatalf("unlinked provider errors must not be displayed as the raw platform, sql=%s", actualSQL)
		}
		if !strings.Contains(actualSQL, "oel.account_id IS NOT NULL") {
			t.Fatalf("provider status should exclude errors that are not linked to an account, sql=%s", actualSQL)
		}
		return nil
	})))
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := &opsRepository{db: db}
	start := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	mock.ExpectQuery("provider summary should exclude unlinked errors").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider",
			"request_count",
			"success_count",
			"failure_count",
			"business_limited_count",
			"cache_read_rate",
			"p50_ms",
			"p95_ms",
			"p99_ms",
			"duration_avg_ms",
			"duration_max_ms",
			"ttft_avg_ms",
			"ttft_p95_ms",
			"ttft_sample_count",
			"upstream_ttft_avg_ms",
			"gateway_ttft_avg_ms",
			"timeout_524_count",
			"timeout_524_avg_ms",
			"last_seen",
		}))

	items, err := repo.queryProviderStatusSummary(context.Background(), start, end, 50)
	if err != nil {
		t.Fatalf("queryProviderStatusSummary returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items = %#v, want no unlinked provider rows", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestQueryProviderStatusSummaryIncludesTimingDiagnostics(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		for _, want := range []string{
			"ul.first_token_ms",
			"ul.upstream_latency_ms",
			"oel.duration_ms",
			"oel.time_to_first_token_ms",
			"oel.upstream_status_code",
			"COUNT(*) FILTER (WHERE is_timeout_524)",
			"AVG(duration_ms) FILTER (WHERE is_timeout_524",
		} {
			if !strings.Contains(actualSQL, want) {
				t.Fatalf("provider status diagnostics query missing %q, sql=%s", want, actualSQL)
			}
		}
		return nil
	})))
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := &opsRepository{db: db}
	start := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	mock.ExpectQuery("provider summary should include diagnostics").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider",
			"request_count",
			"success_count",
			"failure_count",
			"business_limited_count",
			"cache_read_rate",
			"p50_ms",
			"p95_ms",
			"p99_ms",
			"duration_avg_ms",
			"duration_max_ms",
			"ttft_avg_ms",
			"ttft_p95_ms",
			"ttft_sample_count",
			"upstream_ttft_avg_ms",
			"gateway_ttft_avg_ms",
			"timeout_524_count",
			"timeout_524_avg_ms",
			"last_seen",
		}).AddRow("gzw plus", int64(3), int64(2), int64(1), int64(0), nil, float64(120), float64(300), float64(500), float64(250), float64(90000), float64(1800), float64(3000), int64(2), float64(600), float64(1200), int64(1), float64(90000), end))

	items, err := repo.queryProviderStatusSummary(context.Background(), start, end, 50)
	if err != nil {
		t.Fatalf("queryProviderStatusSummary returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	item := items[0]
	if item.DurationAvgMs == nil || *item.DurationAvgMs != 250 {
		t.Fatalf("duration_avg_ms = %#v, want 250", item.DurationAvgMs)
	}
	if item.DurationMaxMs == nil || *item.DurationMaxMs != 90000 {
		t.Fatalf("duration_max_ms = %#v, want 90000", item.DurationMaxMs)
	}
	if item.TTFTAvgMs == nil || *item.TTFTAvgMs != 1800 {
		t.Fatalf("ttft_avg_ms = %#v, want 1800", item.TTFTAvgMs)
	}
	if item.TTFTP95Ms == nil || *item.TTFTP95Ms != 3000 {
		t.Fatalf("ttft_p95_ms = %#v, want 3000", item.TTFTP95Ms)
	}
	if item.TTFTSampleCount != 2 {
		t.Fatalf("ttft_sample_count = %d, want 2", item.TTFTSampleCount)
	}
	if item.UpstreamTTFTAvgMs == nil || *item.UpstreamTTFTAvgMs != 600 {
		t.Fatalf("upstream_ttft_avg_ms = %#v, want 600", item.UpstreamTTFTAvgMs)
	}
	if item.GatewayTTFTAvgMs == nil || *item.GatewayTTFTAvgMs != 1200 {
		t.Fatalf("gateway_ttft_avg_ms = %#v, want 1200", item.GatewayTTFTAvgMs)
	}
	if item.Timeout524Count != 1 {
		t.Fatalf("timeout_524_count = %d, want 1", item.Timeout524Count)
	}
	if item.Timeout524AvgMs == nil || *item.Timeout524AvgMs != 90000 {
		t.Fatalf("timeout_524_avg_ms = %#v, want 90000", item.Timeout524AvgMs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestQueryProviderStatusSummaryIncludesCacheReadRate(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		for _, want := range []string{
			"ul.input_tokens",
			"ul.cache_read_tokens",
			"ul.cache_creation_tokens",
			"cache_read_rate",
		} {
			if !strings.Contains(actualSQL, want) {
				t.Fatalf("provider status cache query missing %q, sql=%s", want, actualSQL)
			}
		}
		return nil
	})))
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := &opsRepository{db: db}
	start := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	mock.ExpectQuery("provider summary should include cache read rate").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider",
			"request_count",
			"success_count",
			"failure_count",
			"business_limited_count",
			"cache_read_rate",
			"p50_ms",
			"p95_ms",
			"p99_ms",
			"duration_avg_ms",
			"duration_max_ms",
			"ttft_avg_ms",
			"ttft_p95_ms",
			"ttft_sample_count",
			"upstream_ttft_avg_ms",
			"gateway_ttft_avg_ms",
			"timeout_524_count",
			"timeout_524_avg_ms",
			"last_seen",
		}).AddRow("gzw plus", int64(3), int64(3), int64(0), int64(0), float64(86.4823), float64(120), float64(300), float64(500), float64(250), float64(90000), float64(1800), float64(3000), int64(2), float64(600), float64(1200), int64(0), nil, end))

	items, err := repo.queryProviderStatusSummary(context.Background(), start, end, 50)
	if err != nil {
		t.Fatalf("queryProviderStatusSummary returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if items[0].CacheReadRate == nil || *items[0].CacheReadRate != 86.4823 {
		t.Fatalf("cache_read_rate = %#v, want 86.4823", items[0].CacheReadRate)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetProviderStatusAttachesLatestFingerprints(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		if strings.Contains(actualSQL, "SELECT provider, headers_json") && !strings.Contains(actualSQL, "jsonb_array_elements") {
			t.Fatalf("fingerprint query must expand upstream_errors events, sql=%s", actualSQL)
		}
		if strings.Contains(actualSQL, "SELECT provider, headers_json") && !strings.Contains(actualSQL, "ev->'fingerprint'->'headers'") {
			t.Fatalf("fingerprint query must read event fingerprint headers, sql=%s", actualSQL)
		}
		return nil
	})))
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := &opsRepository{db: db}
	start := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	mock.ExpectQuery("provider summary").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider",
			"request_count",
			"success_count",
			"failure_count",
			"business_limited_count",
			"cache_read_rate",
			"p50_ms",
			"p95_ms",
			"p99_ms",
			"duration_avg_ms",
			"duration_max_ms",
			"ttft_avg_ms",
			"ttft_p95_ms",
			"ttft_sample_count",
			"upstream_ttft_avg_ms",
			"gateway_ttft_avg_ms",
			"timeout_524_count",
			"timeout_524_avg_ms",
			"last_seen",
		}).AddRow("xixi", int64(2), int64(1), int64(1), int64(0), nil, float64(120), float64(300), float64(500), float64(250), float64(500), float64(180), float64(300), int64(1), nil, nil, int64(1), float64(90000), end))
	mock.ExpectQuery("provider timeline").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider",
			"bucket_start",
			"request_count",
			"success_count",
			"failure_count",
			"p50_ms",
			"p95_ms",
			"p99_ms",
			"duration_avg_ms",
			"ttft_avg_ms",
			"ttft_sample_count",
			"upstream_ttft_avg_ms",
			"gateway_ttft_avg_ms",
			"timeout_524_count",
			"timeout_524_avg_ms",
		}).AddRow("xixi", start, int64(2), int64(1), int64(1), float64(120), float64(300), float64(500), float64(250), float64(180), int64(1), nil, nil, int64(1), float64(90000)))
	mock.ExpectQuery("provider fingerprints").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider",
			"headers_json",
			"last_seen",
		}).AddRow("xixi", `{"server":"cloudflare","cf-ray":"abc-HKG","authorization":"Bearer should-not-exist"}`, end))

	resp, err := repo.GetProviderStatus(context.Background(), &service.OpsProviderStatusFilter{StartTime: start, EndTime: end, BucketSeconds: 300, Limit: 50})
	if err != nil {
		t.Fatalf("GetProviderStatus returned error: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(resp.Items))
	}
	fp := resp.Items[0].Fingerprint
	if fp == nil {
		t.Fatal("expected provider fingerprint")
	}
	if fp.Headers["server"] != "cloudflare" || fp.Headers["cf-ray"] != "abc-HKG" {
		t.Fatalf("fingerprint headers = %#v", fp.Headers)
	}
	if _, ok := fp.Headers["authorization"]; ok {
		t.Fatalf("sensitive header leaked: %#v", fp.Headers)
	}
	if fp.LastSeen == nil || !fp.LastSeen.Equal(end) {
		t.Fatalf("fingerprint last_seen = %#v, want %s", fp.LastSeen, end)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestScanProviderStatusSummaryCalculatesAvailability(t *testing.T) {
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	scanner := &mockSummaryScanner{values: []any{
		"anthropic",
		int64(10),
		int64(8),
		int64(2),
		int64(1),
		sql.NullFloat64{},
		sql.NullFloat64{Float64: 120, Valid: true},
		sql.NullFloat64{Float64: 300, Valid: true},
		sql.NullFloat64{Float64: 500, Valid: true},
		sql.NullFloat64{Float64: 260, Valid: true},
		sql.NullFloat64{Float64: 800, Valid: true},
		sql.NullFloat64{Float64: 1100, Valid: true},
		sql.NullFloat64{Float64: 2500, Valid: true},
		int64(6),
		sql.NullFloat64{Float64: 400, Valid: true},
		sql.NullFloat64{Float64: 700, Valid: true},
		int64(1),
		sql.NullFloat64{Float64: 88000, Valid: true},
		sql.NullTime{Time: now, Valid: true},
	}}

	item, err := scanProviderStatusSummary(scanner)
	if err != nil {
		t.Fatalf("scanProviderStatusSummary returned error: %v", err)
	}
	if item.Availability != 80 || item.ErrorRate != 20 {
		t.Fatalf("availability/error_rate = %.2f/%.2f, want 80/20", item.Availability, item.ErrorRate)
	}
	if item.P95Ms == nil || *item.P95Ms != 300 {
		t.Fatalf("p95_ms = %#v, want 300", item.P95Ms)
	}
	if item.TTFTSampleCount != 6 || item.Timeout524Count != 1 {
		t.Fatalf("ttft_sample_count/timeout_524_count = %d/%d, want 6/1", item.TTFTSampleCount, item.Timeout524Count)
	}
}

type mockSummaryScanner struct {
	values []any
}

func (m *mockSummaryScanner) Scan(dest ...any) error {
	for i := range dest {
		switch d := dest[i].(type) {
		case *string:
			*d = m.values[i].(string)
		case *int64:
			*d = m.values[i].(int64)
		case *sql.NullFloat64:
			*d = m.values[i].(sql.NullFloat64)
		case *sql.NullTime:
			*d = m.values[i].(sql.NullTime)
		default:
			return sql.ErrNoRows
		}
	}
	return nil
}

var _ service.OpsRepository = (*opsRepository)(nil)
