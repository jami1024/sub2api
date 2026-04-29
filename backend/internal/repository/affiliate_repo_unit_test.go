package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func newAffiliateRepoSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func affiliateSummaryRows(debtQuota float64) *sqlmock.Rows {
	now := time.Date(2026, 4, 28, 9, 0, 0, 0, time.UTC)
	return sqlmock.NewRows([]string{
		"user_id",
		"aff_code",
		"aff_code_custom",
		"aff_rebate_rate_percent",
		"inviter_id",
		"aff_count",
		"aff_quota",
		"aff_frozen_quota",
		"aff_history_quota",
		"debt_quota",
		"created_at",
		"updated_at",
	}).AddRow(
		int64(11),
		"AFFCODE11",
		false,
		nil,
		nil,
		2,
		120.0,
		0.0,
		180.0,
		debtQuota,
		now,
		now,
	)
}

func TestQueryAffiliateByUserIDReadsDebtQuota(t *testing.T) {
	ctx := context.Background()
	db, mock := newAffiliateRepoSQLMock(t)

	mock.ExpectQuery(`(?s)SELECT user_id,.*debt_quota::double precision.*FROM user_affiliates\s+WHERE user_id = \$1`).
		WithArgs(int64(11)).
		WillReturnRows(affiliateSummaryRows(25.5))

	summary, err := queryAffiliateByUserID(ctx, db, 11)

	require.NoError(t, err)
	require.InDelta(t, 25.5, summary.DebtQuota, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryAffiliateByCodeReadsDebtQuota(t *testing.T) {
	ctx := context.Background()
	db, mock := newAffiliateRepoSQLMock(t)

	mock.ExpectQuery(`(?s)SELECT user_id,.*debt_quota::double precision.*FROM user_affiliates\s+WHERE aff_code = \$1\s+LIMIT 1`).
		WithArgs(regexp.QuoteMeta("AFFCODE11")).
		WillReturnRows(affiliateSummaryRows(33.25))

	summary, err := queryAffiliateByCode(ctx, db, "affcode11")

	require.NoError(t, err)
	require.InDelta(t, 33.25, summary.DebtQuota, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}
