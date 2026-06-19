package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestInsertUpstreamMultiplierSampleStoresAndReturnsRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOpsRepository(db)
	now := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
	httpStatus := 200
	standard := 0.1
	actual := 0.012
	multiplier := 0.12

	rows := sqlmock.NewRows([]string{
		"id", "account_id", "account_name_snapshot", "platform", "base_url_snapshot", "key_prefix_snapshot",
		"model", "status", "http_status", "standard_cost_delta", "actual_cost_delta", "multiplier",
		"balance_before", "balance_after", "error_message", "measured_at", "created_at",
	}).AddRow(
		int64(1), int64(12), "xixi", "openai", "https://upstream.example.com", "sk-test-",
		"gpt-5.4", "success", httpStatus, standard, actual, multiplier,
		nil, nil, "", now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ops_upstream_multiplier_samples")).
		WithArgs(
			int64(12), "xixi", "openai", "https://upstream.example.com", "sk-test-",
			"gpt-5.4", "success", &httpStatus, &standard, &actual, &multiplier,
			nil, nil, "", now,
		).
		WillReturnRows(rows)

	got, err := repo.InsertUpstreamMultiplierSample(context.Background(), &service.OpsUpstreamMultiplierSample{
		AccountID:           12,
		AccountNameSnapshot: "xixi",
		Platform:            "openai",
		BaseURLSnapshot:     "https://upstream.example.com",
		KeyPrefixSnapshot:   "sk-test-",
		Model:               "gpt-5.4",
		Status:              "success",
		HTTPStatus:          &httpStatus,
		StandardCostDelta:   &standard,
		ActualCostDelta:     &actual,
		Multiplier:          &multiplier,
		MeasuredAt:          now,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), got.ID)
	require.Equal(t, "sk-test-", got.KeyPrefixSnapshot)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListUpstreamMultiplierSamplesFiltersByModelAndAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewOpsRepository(db)
	accountID := int64(12)
	now := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "account_id", "account_name_snapshot", "platform", "base_url_snapshot", "key_prefix_snapshot",
		"model", "status", "http_status", "standard_cost_delta", "actual_cost_delta", "multiplier",
		"balance_before", "balance_after", "error_message", "measured_at", "created_at",
	}).AddRow(
		int64(2), accountID, "xixi", "openai", "https://upstream.example.com", "sk-test-",
		"gpt-5.4", "success", 200, 0.1, 0.012, 0.12,
		nil, nil, "", now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, account_id, account_name_snapshot")).
		WithArgs("gpt-5.4", accountID, 50).
		WillReturnRows(rows)

	got, err := repo.ListUpstreamMultiplierSamples(context.Background(), &service.OpsUpstreamMultiplierSamplesFilter{
		Model:     "gpt-5.4",
		AccountID: &accountID,
		Limit:     50,
	})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.InDelta(t, 0.12, *got[0].Multiplier, 0.000001)
	require.NoError(t, mock.ExpectationsWereMet())
}
