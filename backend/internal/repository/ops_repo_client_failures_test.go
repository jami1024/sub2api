package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestQueryClientFailureStatsAggregatesByUserOnly(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		if strings.Contains(actualSQL, "account_id") || strings.Contains(actualSQL, "accounts") {
			t.Fatalf("client failure stats must not aggregate by upstream account, sql=%s", actualSQL)
		}
		for _, want := range []string{
			"GROUP BY user_id, user_email",
			"COUNT(DISTINCT api_key_id)",
			"error_owner = 'client'",
			"error_source = 'client_request'",
			"Failed to read request body",
		} {
			if !strings.Contains(actualSQL, want) {
				t.Fatalf("client failure query missing %q, sql=%s", want, actualSQL)
			}
		}
		return nil
	})))
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := &opsRepository{db: db}
	start := time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	mock.ExpectQuery("client failure stats should aggregate by user").
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id",
			"user_email",
			"failure_count",
			"affected_key_count",
			"top_error_message",
			"top_error_count",
			"last_seen",
			"top_inbound_endpoint",
			"top_platform",
		}).AddRow(int64(123), "3027896911@qq.com", int64(5), int64(2), "Failed to read request body", int64(4), end, "/responses", "openai"))

	items, err := repo.queryClientFailureStats(context.Background(), start, end, 50)
	if err != nil {
		t.Fatalf("queryClientFailureStats returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	item := items[0]
	if item.UserEmail != "3027896911@qq.com" || item.FailureCount != 5 || item.AffectedKeyCount != 2 {
		t.Fatalf("unexpected item: %#v", item)
	}
	if item.TopErrorMessage != "Failed to read request body" || item.TopInboundEndpoint != "/responses" || item.TopPlatform != "openai" {
		t.Fatalf("unexpected top fields: %#v", item)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
