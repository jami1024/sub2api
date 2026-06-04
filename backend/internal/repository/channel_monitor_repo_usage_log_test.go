package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestChannelMonitorRepositoryListLatestSuccessfulOpenAIUsageByModels(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer func() { _ = db.Close() }()

	createdAt := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"target_model", "duration_ms", "created_at"}).
		AddRow("gpt-5.4", int64(120), createdAt).
		AddRow("gpt-5.4-mini", nil, createdAt.Add(-time.Minute))

	mock.ExpectQuery(regexp.QuoteMeta("WITH targets AS")).
		WithArgs(pq.Array([]string{"gpt-5.4", "gpt-5.4-mini"})).
		WillReturnRows(rows)

	repo := &channelMonitorRepository{db: db}
	got, err := repo.ListLatestSuccessfulOpenAIUsageByModels(
		context.Background(),
		[]string{" gpt-5.4 ", "gpt-5.4-mini", "gpt-5.4"},
	)
	if err != nil {
		t.Fatalf("ListLatestSuccessfulOpenAIUsageByModels returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got["gpt-5.4"] == nil || got["gpt-5.4"].DurationMs == nil || *got["gpt-5.4"].DurationMs != 120 {
		t.Fatalf("gpt-5.4 latest = %#v, want duration 120", got["gpt-5.4"])
	}
	if got["gpt-5.4-mini"] == nil || got["gpt-5.4-mini"].DurationMs != nil {
		t.Fatalf("gpt-5.4-mini latest = %#v, want nil duration", got["gpt-5.4-mini"])
	}
}
