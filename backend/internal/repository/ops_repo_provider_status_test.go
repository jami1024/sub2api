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
		}).AddRow("openai", start, int64(1), int64(1), int64(0), float64(100), float64(100), float64(100)))

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
			"p50_ms",
			"p95_ms",
			"p99_ms",
			"last_seen",
		}).AddRow("gzw plus", int64(80), int64(80), int64(0), int64(0), float64(120), float64(300), float64(500), end))

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

func TestScanProviderStatusSummaryCalculatesAvailability(t *testing.T) {
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	scanner := &mockSummaryScanner{values: []any{
		"anthropic",
		int64(10),
		int64(8),
		int64(2),
		int64(1),
		sql.NullFloat64{Float64: 120, Valid: true},
		sql.NullFloat64{Float64: 300, Valid: true},
		sql.NullFloat64{Float64: 500, Valid: true},
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
