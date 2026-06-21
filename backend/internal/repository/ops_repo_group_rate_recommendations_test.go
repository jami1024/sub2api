package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestOpsRepositoryGetGroupRateRecommendationSourceData(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		matched, err := regexp.MatchString(expectedSQL, actualSQL)
		if err != nil {
			return err
		}
		if !matched {
			return sqlmock.ErrCancelled
		}
		return nil
	})))
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(`SELECT id, name, price, credit_amount, package_scope,`).
		WithArgs("codex").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "price", "credit_amount", "package_scope", "revenue_per_credit"}).
			AddRow(int64(2), "专属包-进阶级", 100.0, 400.0, "codex", 0.25))

	mock.ExpectQuery(`SELECT\s+g\.id,`).
		WithArgs("codex").
		WillReturnRows(sqlmock.NewRows([]string{
			"group_id", "group_name", "rate_multiplier", "package_scope", "allow_image_generation",
			"account_id", "account_name", "platform", "type", "status", "schedulable", "current_priority", "binding_priority", "base_url", "key_prefix",
		}).AddRow(int64(8), "gpt pro", 1.0, "codex", false, int64(9), "xixi", "openai", "apikey", "active", true, 1, 1, "https://xixiapi.cc", "sk-12345"))

	mock.ExpectQuery(`WITH base AS`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"group_id", "account_id", "request_count", "request_share", "standard_cost", "standard_cost_share"}).
			AddRow(int64(8), int64(9), int64(10), 1.0, 12.5, 1.0))

	mock.ExpectQuery(`SELECT DISTINCT ON \(account_id\)`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "account_id", "account_name_snapshot", "platform", "base_url_snapshot", "key_prefix_snapshot",
			"model", "status", "http_status", "standard_cost_delta", "actual_cost_delta", "multiplier",
			"balance_before", "balance_after", "error_message", "measured_at", "created_at",
		}))

	repo := NewOpsRepository(db)
	got, err := repo.GetGroupRateRecommendationSourceData(context.Background(), &service.OpsGroupRateRecommendationFilter{Model: "gpt-5.4", PackageScope: "codex", UsageDays: 7})
	if err != nil {
		t.Fatalf("GetGroupRateRecommendationSourceData error: %v", err)
	}
	if len(got.Packages) != 1 || got.Packages[0].RevenuePerCredit != 0.25 {
		t.Fatalf("packages = %#v", got.Packages)
	}
	if len(got.Groups) != 1 || len(got.Groups[0].Accounts) != 1 {
		t.Fatalf("groups = %#v", got.Groups)
	}
	if got.Usage[8][9].RequestCount != 10 {
		t.Fatalf("usage = %#v", got.Usage)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
