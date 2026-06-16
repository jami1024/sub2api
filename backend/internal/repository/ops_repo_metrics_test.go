package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestInsertSystemMetricsSkipsDuplicateGlobalMinute(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewOpsRepository(db)
	createdAt := time.Date(2026, 6, 7, 17, 12, 0, 0, time.UTC)

	mock.ExpectExec(`(?s)WHERE NOT EXISTS\s+\(\s+SELECT 1\s+FROM ops_system_metrics`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.InsertSystemMetrics(context.Background(), &service.OpsInsertSystemMetricsInput{
		CreatedAt:       createdAt,
		WindowMinutes:   1,
		CPUUsagePercent: ptrFloat64ForOpsRepoMetricsTest(88.8),
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertSystemMetricsCastsNullableDimensionParams(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewOpsRepository(db)
	groupID := int64(12)
	platform := "openai"

	mock.ExpectExec(`(?s)\$3::varchar.*\$4::bigint.*existing\.platform IS NOT DISTINCT FROM \$3::varchar.*existing\.group_id IS NOT DISTINCT FROM \$4::bigint`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.InsertSystemMetrics(context.Background(), &service.OpsInsertSystemMetricsInput{
		CreatedAt:     time.Date(2026, 6, 16, 14, 10, 0, 0, time.UTC),
		WindowMinutes: 1,
		Platform:      &platform,
		GroupID:       &groupID,
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func ptrFloat64ForOpsRepoMetricsTest(v float64) *float64 {
	return &v
}
