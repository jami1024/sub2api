//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func querySingleFloat(t *testing.T, ctx context.Context, client *dbent.Client, query string, args ...any) float64 {
	t.Helper()
	rows, err := client.QueryContext(ctx, query, args...)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	require.True(t, rows.Next(), "expected one row")
	var value float64
	require.NoError(t, rows.Scan(&value))
	require.NoError(t, rows.Err())
	return value
}

func querySingleInt(t *testing.T, ctx context.Context, client *dbent.Client, query string, args ...any) int {
	t.Helper()
	rows, err := client.QueryContext(ctx, query, args...)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	require.True(t, rows.Next(), "expected one row")
	var value int
	require.NoError(t, rows.Scan(&value))
	require.NoError(t, rows.Err())
	return value
}

func mustCreatePaymentOrder(t *testing.T, ctx context.Context, client *dbent.Client, user *service.User, outTradeNo string) int64 {
	t.Helper()
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("TEST-ORDER-" + outTradeNo).
		SetOutTradeNo(outTradeNo).
		SetPaymentType("alipay").
		SetPaymentTradeNo("trade-" + outTradeNo).
		SetOrderType("balance_package").
		SetStatus("COMPLETED").
		SetPaidAt(time.Now()).
		SetCompletedAt(time.Now()).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("example.com").
		Save(ctx)
	require.NoError(t, err)
	return order.ID
}

func TestAffiliateRepository_TransferQuotaToBalance_UsesClaimedQuotaBeforeClear(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	txCtx := dbent.NewTxContext(ctx, tx)
	client := tx.Client()

	repo := NewAffiliateRepository(client, integrationDB)

	u := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("affiliate-transfer-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Balance:      5.5,
		Concurrency:  5,
	})

	affCode := fmt.Sprintf("AFF%09d", time.Now().UnixNano()%1_000_000_000)
	_, err := client.ExecContext(txCtx, `
INSERT INTO user_affiliates (user_id, aff_code, aff_quota, aff_history_quota, created_at, updated_at)
VALUES ($1, $2, $3, $3, NOW(), NOW())`, u.ID, affCode, 12.34)
	require.NoError(t, err)

	transferred, balance, err := repo.TransferQuotaToBalance(txCtx, u.ID)
	require.NoError(t, err)
	require.InDelta(t, 12.34, transferred, 1e-9)
	require.InDelta(t, 17.84, balance, 1e-9)

	affQuota := querySingleFloat(t, txCtx, client,
		"SELECT aff_quota::double precision FROM user_affiliates WHERE user_id = $1", u.ID)
	require.InDelta(t, 0.0, affQuota, 1e-9)

	persistedBalance := querySingleFloat(t, txCtx, client,
		"SELECT balance::double precision FROM users WHERE id = $1", u.ID)
	require.InDelta(t, 17.84, persistedBalance, 1e-9)

	ledgerCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM user_affiliate_ledger WHERE user_id = $1 AND action = 'transfer'", u.ID)
	require.Equal(t, 1, ledgerCount)
}

// TestAffiliateRepository_AccrueQuota_ReusesOuterTransaction guards the
// cross-layer tx propagation invariant: when AccrueQuota is called with a ctx
// that already carries a transaction (via dbent.NewTxContext), repo.withTx
// must reuse that tx rather than opening a nested one. If this invariant
// breaks, AccrueQuota would commit independently and survive a rollback of
// the outer tx, which would violate payment_fulfillment's all-or-nothing
// semantics.
func TestAffiliateRepository_AccrueQuota_ReusesOuterTransaction(t *testing.T) {
	ctx := context.Background()

	outerTx, err := integrationEntClient.Tx(ctx)
	require.NoError(t, err, "begin outer tx")
	// Defensive cleanup: if any require.* below fires before the explicit
	// Rollback, this prevents the tx from leaking until container teardown.
	// Rollback is idempotent at the driver level (extra rollback returns an
	// error we ignore).
	t.Cleanup(func() { _ = outerTx.Rollback() })
	client := outerTx.Client()
	txCtx := dbent.NewTxContext(ctx, outerTx)

	inviter := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("affiliate-inviter-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Concurrency:  5,
	})
	invitee := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("affiliate-invitee-%d@example.com", time.Now().UnixNano()+1),
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Concurrency:  5,
	})

	repo := NewAffiliateRepository(client, integrationDB)
	_, err = repo.EnsureUserAffiliate(txCtx, inviter.ID)
	require.NoError(t, err)
	_, err = repo.EnsureUserAffiliate(txCtx, invitee.ID)
	require.NoError(t, err)

	bound, err := repo.BindInviter(txCtx, invitee.ID, inviter.ID)
	require.NoError(t, err)
	require.True(t, bound, "invitee must bind to inviter")

	applied, err := repo.AccrueQuota(txCtx, inviter.ID, invitee.ID, 3.5)
	require.NoError(t, err)
	require.True(t, applied, "AccrueQuota must report applied=true")

	// Visible inside the outer tx.
	innerQuota := querySingleFloat(t, txCtx, client,
		"SELECT aff_quota::double precision FROM user_affiliates WHERE user_id = $1", inviter.ID)
	require.InDelta(t, 3.5, innerQuota, 1e-9)

	// Roll back the outer tx; if AccrueQuota had opened its own inner tx and
	// committed it, the rows would still be visible to the global client.
	require.NoError(t, outerTx.Rollback())

	rows, err := integrationEntClient.QueryContext(ctx,
		"SELECT COUNT(*) FROM user_affiliates WHERE user_id IN ($1, $2)",
		inviter.ID, invitee.ID)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	require.True(t, rows.Next())
	var postRollbackCount int
	require.NoError(t, rows.Scan(&postRollbackCount))
	require.Equal(t, 0, postRollbackCount,
		"AccrueQuota must propagate the outer tx — found persisted rows after rollback")
}

func TestAffiliateRepository_TransferQuotaToBalance_EmptyQuota(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	txCtx := dbent.NewTxContext(ctx, tx)
	client := tx.Client()

	repo := NewAffiliateRepository(client, integrationDB)

	u := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("affiliate-empty-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Balance:      3.21,
		Concurrency:  5,
	})

	affCode := fmt.Sprintf("AFF%09d", time.Now().UnixNano()%1_000_000_000)
	_, err := client.ExecContext(txCtx, `
INSERT INTO user_affiliates (user_id, aff_code, aff_quota, aff_history_quota, created_at, updated_at)
VALUES ($1, $2, 0, 0, NOW(), NOW())`, u.ID, affCode)
	require.NoError(t, err)

	transferred, balance, err := repo.TransferQuotaToBalance(txCtx, u.ID)
	require.ErrorIs(t, err, service.ErrAffiliateQuotaEmpty)
	require.InDelta(t, 0.0, transferred, 1e-9)
	require.InDelta(t, 0.0, balance, 1e-9)

	persistedBalance := querySingleFloat(t, txCtx, client,
		"SELECT balance::double precision FROM users WHERE id = $1", u.ID)
	require.InDelta(t, 3.21, persistedBalance, 1e-9)
}

func TestAffiliateRepository_ReleaseDuePendingRebateRecords(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	txCtx := dbent.NewTxContext(ctx, tx)
	client := tx.Client()

	repo := NewAffiliateRepository(client, integrationDB)

	u := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("affiliate-release-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Balance:      0,
		Concurrency:  5,
	})

	affCode := fmt.Sprintf("REL%09d", time.Now().UnixNano()%1_000_000_000)
	_, err := client.ExecContext(txCtx, `
INSERT INTO user_affiliates (user_id, aff_code, aff_quota, aff_history_quota, created_at, updated_at)
VALUES ($1, $2, 0, 0, NOW(), NOW())`, u.ID, affCode)
	require.NoError(t, err)

	orderID1 := mustCreatePaymentOrder(t, txCtx, client, u, fmt.Sprintf("release-1-%d", time.Now().UnixNano()))
	orderID2 := mustCreatePaymentOrder(t, txCtx, client, u, fmt.Sprintf("release-2-%d", time.Now().UnixNano()))
	orderID3 := mustCreatePaymentOrder(t, txCtx, client, u, fmt.Sprintf("release-3-%d", time.Now().UnixNano()))

	_, err = client.ExecContext(txCtx, `
INSERT INTO affiliate_rebate_records (user_id, source_user_id, source_order_id, level, rate, base_amount, rebate_amount, status, available_at, created_at, updated_at)
VALUES
  ($1, $1, $2, 1, 6, 100, 6, 'pending', NOW() - INTERVAL '1 hour', NOW(), NOW()),
  ($1, $1, $3, 2, 3, 100, 3, 'pending', NOW() - INTERVAL '2 hour', NOW(), NOW()),
  ($1, $1, $4, 3, 1, 100, 1, 'pending', NOW() + INTERVAL '2 hour', NOW(), NOW())
`, u.ID, orderID1, orderID2, orderID3)
	require.NoError(t, err)

	released, err := repo.ReleaseDuePendingRebateRecords(txCtx, time.Now())
	require.NoError(t, err)
	require.Equal(t, 2, released)

	affQuota := querySingleFloat(t, txCtx, client,
		"SELECT aff_quota::double precision FROM user_affiliates WHERE user_id = $1", u.ID)
	require.InDelta(t, 9.0, affQuota, 1e-9)

	affHistoryQuota := querySingleFloat(t, txCtx, client,
		"SELECT aff_history_quota::double precision FROM user_affiliates WHERE user_id = $1", u.ID)
	require.InDelta(t, 9.0, affHistoryQuota, 1e-9)

	availableCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_rebate_records WHERE user_id = $1 AND status = 'available'", u.ID)
	require.Equal(t, 2, availableCount)

	pendingCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_rebate_records WHERE user_id = $1 AND status = 'pending'", u.ID)
	require.Equal(t, 1, pendingCount)
}

func TestAffiliateRepository_CreateWithdrawalRequest(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	txCtx := dbent.NewTxContext(ctx, tx)
	client := tx.Client()

	repo := NewAffiliateRepository(client, integrationDB)

	u := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("affiliate-withdraw-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Balance:      0,
		Concurrency:  5,
	})

	affCode := fmt.Sprintf("WD%010d", time.Now().UnixNano()%10_000_000_000)
	_, err := client.ExecContext(txCtx, `
INSERT INTO user_affiliates (user_id, aff_code, aff_quota, aff_history_quota, created_at, updated_at)
VALUES ($1, $2, 180, 220, NOW(), NOW())`, u.ID, affCode)
	require.NoError(t, err)

	orderID1 := mustCreatePaymentOrder(t, txCtx, client, u, fmt.Sprintf("withdraw-1-%d", time.Now().UnixNano()))
	orderID2 := mustCreatePaymentOrder(t, txCtx, client, u, fmt.Sprintf("withdraw-2-%d", time.Now().UnixNano()))

	_, err = client.ExecContext(txCtx, `
INSERT INTO affiliate_rebate_records (user_id, source_user_id, source_order_id, level, rate, base_amount, rebate_amount, status, available_at, created_at, updated_at)
VALUES
  ($1, $1, $2, 1, 6, 100, 60, 'available', NOW() - INTERVAL '1 day', NOW(), NOW()),
  ($1, $1, $3, 2, 3, 100, 60, 'available', NOW() - INTERVAL '2 day', NOW(), NOW())
`, u.ID, orderID1, orderID2)
	require.NoError(t, err)

	item, err := repo.CreateWithdrawalRequest(txCtx, u.ID, 120, "manual payout")
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, u.ID, item.UserID)
	require.InDelta(t, 120.0, item.Amount, 1e-9)
	require.Equal(t, "pending", item.Status)

	affQuota := querySingleFloat(t, txCtx, client,
		"SELECT aff_quota::double precision FROM user_affiliates WHERE user_id = $1", u.ID)
	require.InDelta(t, 60.0, affQuota, 1e-9)

	requestCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_withdrawal_requests WHERE user_id = $1 AND status = 'pending'", u.ID)
	require.Equal(t, 1, requestCount)

	withdrawRequestedCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_rebate_records WHERE user_id = $1 AND status = 'withdraw_requested'", u.ID)
	require.Equal(t, 2, withdrawRequestedCount)

	requestItemCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_withdrawal_request_items WHERE withdrawal_request_id = $1", item.ID)
	require.Equal(t, 2, requestItemCount)
}

func TestAffiliateRepository_ReviewWithdrawalRequest(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	txCtx := dbent.NewTxContext(ctx, tx)
	client := tx.Client()

	repo := NewAffiliateRepository(client, integrationDB)

	u := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("affiliate-review-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Balance:      0,
		Concurrency:  5,
	})

	affCode := fmt.Sprintf("RV%010d", time.Now().UnixNano()%10_000_000_000)
	_, err := client.ExecContext(txCtx, `
INSERT INTO user_affiliates (user_id, aff_code, aff_quota, aff_history_quota, created_at, updated_at)
VALUES ($1, $2, 240, 240, NOW(), NOW())`, u.ID, affCode)
	require.NoError(t, err)

	orderID1 := mustCreatePaymentOrder(t, txCtx, client, u, fmt.Sprintf("review-1-%d", time.Now().UnixNano()))
	orderID2 := mustCreatePaymentOrder(t, txCtx, client, u, fmt.Sprintf("review-2-%d", time.Now().UnixNano()))

	_, err = client.ExecContext(txCtx, `
INSERT INTO affiliate_rebate_records (user_id, source_user_id, source_order_id, level, rate, base_amount, rebate_amount, status, available_at, created_at, updated_at)
VALUES
  ($1, $1, $2, 1, 6, 100, 120, 'available', NOW() - INTERVAL '1 day', NOW(), NOW()),
  ($1, $1, $3, 2, 3, 100, 120, 'available', NOW() - INTERVAL '2 day', NOW(), NOW())
`, u.ID, orderID1, orderID2)
	require.NoError(t, err)

	req, err := repo.CreateWithdrawalRequest(txCtx, u.ID, 120, "manual payout")
	require.NoError(t, err)

	rejected, err := repo.RejectWithdrawalRequest(txCtx, req.ID, 9001, "reject")
	require.NoError(t, err)
	require.Equal(t, "rejected", rejected.Status)

	affQuota := querySingleFloat(t, txCtx, client,
		"SELECT aff_quota::double precision FROM user_affiliates WHERE user_id = $1", u.ID)
	require.InDelta(t, 240.0, affQuota, 1e-9)

	availableCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_rebate_records WHERE user_id = $1 AND status = 'available'", u.ID)
	require.Equal(t, 2, availableCount)

	_, err = client.ExecContext(txCtx, `
UPDATE affiliate_rebate_records
SET status = 'available', updated_at = NOW()
WHERE user_id = $1
`, u.ID)
	require.NoError(t, err)

	req2, err := repo.CreateWithdrawalRequest(txCtx, u.ID, 120, "manual payout 2")
	require.NoError(t, err)

	paid, err := repo.MarkWithdrawalPaid(txCtx, req2.ID, 9002, "paid")
	require.NoError(t, err)
	require.Equal(t, "paid", paid.Status)
	require.NotNil(t, paid.PaidAt)

	withdrawPaidCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_rebate_records WHERE user_id = $1 AND status = 'withdraw_paid'", u.ID)
	require.Equal(t, 1, withdrawPaidCount)
}

func TestAffiliateRepository_ReverseRebatesForOrder(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	txCtx := dbent.NewTxContext(ctx, tx)
	client := tx.Client()

	repo := NewAffiliateRepository(client, integrationDB)

	u := mustCreateUser(t, client, &service.User{
		Email:        fmt.Sprintf("affiliate-reverse-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Balance:      0,
		Concurrency:  5,
	})

	affCode := fmt.Sprintf("RB%010d", time.Now().UnixNano()%10_000_000_000)
	_, err := client.ExecContext(txCtx, `
INSERT INTO user_affiliates (user_id, aff_code, aff_quota, aff_history_quota, debt_quota, created_at, updated_at)
VALUES ($1, $2, 80, 200, 0, NOW(), NOW())`, u.ID, affCode)
	require.NoError(t, err)

	orderID := mustCreatePaymentOrder(t, txCtx, client, u, fmt.Sprintf("reverse-%d", time.Now().UnixNano()))

	_, err = client.ExecContext(txCtx, `
INSERT INTO affiliate_rebate_records (user_id, source_user_id, source_order_id, level, rate, base_amount, rebate_amount, status, available_at, created_at, updated_at)
VALUES
  ($1, $1, $2, 1, 6, 100, 20, 'pending', NOW() + INTERVAL '1 day', NOW(), NOW()),
  ($1, $1, $2, 2, 3, 100, 30, 'available', NOW() - INTERVAL '1 day', NOW(), NOW()),
  ($1, $1, $2, 3, 1, 100, 50, 'withdraw_paid', NOW() - INTERVAL '2 day', NOW(), NOW())
`, u.ID, orderID)
	require.NoError(t, err)

	err = repo.ReverseRebatesForOrder(txCtx, orderID)
	require.NoError(t, err)

	cancelledCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_rebate_records WHERE source_order_id = $1 AND status = 'cancelled'", orderID)
	require.Equal(t, 1, cancelledCount)

	reversedCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_rebate_records WHERE source_order_id = $1 AND status = 'reversed'", orderID)
	require.Equal(t, 1, reversedCount)

	debtOffsetCount := querySingleInt(t, txCtx, client,
		"SELECT COUNT(*) FROM affiliate_rebate_records WHERE source_order_id = $1 AND status = 'debt_offset'", orderID)
	require.Equal(t, 1, debtOffsetCount)

	affQuota := querySingleFloat(t, txCtx, client,
		"SELECT aff_quota::double precision FROM user_affiliates WHERE user_id = $1", u.ID)
	require.InDelta(t, 50.0, affQuota, 1e-9)

	debtQuota := querySingleFloat(t, txCtx, client,
		"SELECT debt_quota::double precision FROM user_affiliates WHERE user_id = $1", u.ID)
	require.InDelta(t, 50.0, debtQuota, 1e-9)
}
