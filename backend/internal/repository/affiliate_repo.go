package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

const (
	affiliateCodeLength      = 12
	affiliateCodeMaxAttempts = 12
)

var affiliateCodeCharset = []byte("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

type affiliateQueryExecer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type affiliateRepository struct {
	client *dbent.Client
}

func NewAffiliateRepository(client *dbent.Client, _ *sql.DB) service.AffiliateRepository {
	return &affiliateRepository{client: client}
}

func (r *affiliateRepository) EnsureUserAffiliate(ctx context.Context, userID int64) (*service.AffiliateSummary, error) {
	if userID <= 0 {
		return nil, service.ErrUserNotFound
	}
	client := clientFromContext(ctx, r.client)
	return ensureUserAffiliateWithClient(ctx, client, userID)
}

func (r *affiliateRepository) GetAffiliateByCode(ctx context.Context, code string) (*service.AffiliateSummary, error) {
	client := clientFromContext(ctx, r.client)
	return queryAffiliateByCode(ctx, client, code)
}

func (r *affiliateRepository) BindInviter(ctx context.Context, userID, inviterID int64) (bool, error) {
	var bound bool
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		if _, err := ensureUserAffiliateWithClient(txCtx, txClient, userID); err != nil {
			return err
		}
		if _, err := ensureUserAffiliateWithClient(txCtx, txClient, inviterID); err != nil {
			return err
		}

		res, err := txClient.ExecContext(txCtx,
			"UPDATE user_affiliates SET inviter_id = $1, updated_at = NOW() WHERE user_id = $2 AND inviter_id IS NULL",
			inviterID, userID,
		)
		if err != nil {
			return fmt.Errorf("bind inviter: %w", err)
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			bound = false
			return nil
		}

		if _, err = txClient.ExecContext(txCtx,
			"UPDATE user_affiliates SET aff_count = aff_count + 1, updated_at = NOW() WHERE user_id = $1",
			inviterID,
		); err != nil {
			return fmt.Errorf("increment inviter aff_count: %w", err)
		}
		bound = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return bound, nil
}

func (r *affiliateRepository) AccrueQuota(ctx context.Context, inviterID, inviteeUserID int64, amount float64) (bool, error) {
	if amount <= 0 {
		return false, nil
	}

	var applied bool
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		res, err := txClient.ExecContext(txCtx,
			"UPDATE user_affiliates SET aff_quota = aff_quota + $1, aff_history_quota = aff_history_quota + $1, updated_at = NOW() WHERE user_id = $2",
			amount, inviterID,
		)
		if err != nil {
			return err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			applied = false
			return nil
		}

		if _, err = txClient.ExecContext(txCtx, `
INSERT INTO user_affiliate_ledger (user_id, action, amount, source_user_id, created_at, updated_at)
VALUES ($1, 'accrue', $2, $3, NOW(), NOW())`, inviterID, amount, inviteeUserID); err != nil {
			return fmt.Errorf("insert affiliate accrue ledger: %w", err)
		}

		applied = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return applied, nil
}

func (r *affiliateRepository) TransferQuotaToBalance(ctx context.Context, userID int64) (float64, float64, error) {
	var transferred float64
	var newBalance float64

	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		if _, err := ensureUserAffiliateWithClient(txCtx, txClient, userID); err != nil {
			return err
		}

		rows, err := txClient.QueryContext(txCtx, `
WITH claimed AS (
	SELECT aff_quota::double precision AS amount
	FROM user_affiliates
	WHERE user_id = $1
	  AND aff_quota > 0
	FOR UPDATE
),
cleared AS (
	UPDATE user_affiliates ua
	SET aff_quota = 0,
	    updated_at = NOW()
	FROM claimed c
	WHERE ua.user_id = $1
	RETURNING c.amount
)
SELECT amount
FROM cleared`, userID)
		if err != nil {
			return fmt.Errorf("claim affiliate quota: %w", err)
		}

		if !rows.Next() {
			_ = rows.Close()
			if err := rows.Err(); err != nil {
				return err
			}
			return service.ErrAffiliateQuotaEmpty
		}
		if err := rows.Scan(&transferred); err != nil {
			_ = rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if transferred <= 0 {
			return service.ErrAffiliateQuotaEmpty
		}

		affected, err := txClient.User.Update().
			Where(user.IDEQ(userID)).
			AddBalance(transferred).
			AddTotalRecharged(transferred).
			Save(txCtx)
		if err != nil {
			return fmt.Errorf("credit user balance by affiliate quota: %w", err)
		}
		if affected == 0 {
			return service.ErrUserNotFound
		}

		newBalance, err = queryUserBalance(txCtx, txClient, userID)
		if err != nil {
			return err
		}

		if _, err = txClient.ExecContext(txCtx, `
INSERT INTO user_affiliate_ledger (user_id, action, amount, source_user_id, created_at, updated_at)
VALUES ($1, 'transfer', $2, NULL, NOW(), NOW())`, userID, transferred); err != nil {
			return fmt.Errorf("insert affiliate transfer ledger: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, 0, err
	}

	return transferred, newBalance, nil
}

func (r *affiliateRepository) ListInvitees(ctx context.Context, inviterID int64, limit int) ([]service.AffiliateInvitee, error) {
	if limit <= 0 {
		limit = 100
	}
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
SELECT ua.user_id,
       COALESCE(u.email, ''),
       COALESCE(u.username, ''),
       ua.created_at
FROM user_affiliates ua
LEFT JOIN users u ON u.id = ua.user_id
WHERE ua.inviter_id = $1
ORDER BY ua.created_at DESC
LIMIT $2`, inviterID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	invitees := make([]service.AffiliateInvitee, 0)
	for rows.Next() {
		var item service.AffiliateInvitee
		var createdAt time.Time
		if err := rows.Scan(&item.UserID, &item.Email, &item.Username, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = &createdAt
		invitees = append(invitees, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return invitees, nil
}

func (r *affiliateRepository) ListAncestors(ctx context.Context, userID int64, maxDepth int) ([]service.AffiliateAncestor, error) {
	if userID <= 0 || maxDepth <= 0 {
		return nil, nil
	}
	client := clientFromContext(ctx, r.client)
	ancestors := make([]service.AffiliateAncestor, 0, maxDepth)
	visited := map[int64]struct{}{userID: {}}
	currentUserID := userID

	for level := 1; level <= maxDepth; level++ {
		summary, err := queryAffiliateByUserID(ctx, client, currentUserID)
		if err != nil {
			if errors.Is(err, service.ErrAffiliateProfileNotFound) {
				return ancestors, nil
			}
			return nil, err
		}
		if summary.InviterID == nil || *summary.InviterID <= 0 {
			return ancestors, nil
		}
		inviterID := *summary.InviterID
		if _, seen := visited[inviterID]; seen {
			return ancestors, nil
		}
		visited[inviterID] = struct{}{}
		ancestors = append(ancestors, service.AffiliateAncestor{
			UserID: inviterID,
			Level:  level,
		})
		currentUserID = inviterID
	}

	return ancestors, nil
}

func (r *affiliateRepository) CreatePendingRebateRecords(ctx context.Context, records []service.AffiliateRebateRecordInput) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}
	client := clientFromContext(ctx, r.client)
	created := 0
	for _, record := range records {
		rows, err := client.QueryContext(ctx, `
INSERT INTO affiliate_rebate_records (
  user_id,
  source_user_id,
  source_order_id,
  level,
  rate,
  base_amount,
  rebate_amount,
  status,
  available_at,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
ON CONFLICT (user_id, source_order_id, level) DO NOTHING
RETURNING id
`, record.UserID, record.SourceUserID, record.SourceOrderID, record.Level, record.Rate, record.BaseAmount, record.RebateAmount, record.Status, record.AvailableAt)
		if err != nil {
			return created, err
		}
		if rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				_ = rows.Close()
				return created, err
			}
			created++
		}
		if err := rows.Close(); err != nil {
			return created, err
		}
	}
	return created, nil
}

func (r *affiliateRepository) SumPendingRebateByUser(ctx context.Context, userID int64) (float64, error) {
	if userID <= 0 {
		return 0, nil
	}
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
SELECT COALESCE(SUM(rebate_amount)::double precision, 0)
FROM affiliate_rebate_records
WHERE user_id = $1
  AND status = 'pending'
`, userID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return 0, nil
	}
	var total float64
	if err := rows.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *affiliateRepository) ReleaseDuePendingRebateRecords(ctx context.Context, now time.Time) (int, error) {
	updatedCount := 0
	creditsByUser := make(map[int64]float64)

	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		rows, err := txClient.QueryContext(txCtx, `
UPDATE affiliate_rebate_records
SET status = 'available',
    updated_at = NOW()
WHERE status = 'pending'
  AND available_at <= $1
RETURNING user_id, rebate_amount
`, now)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var userID int64
			var rebateAmount float64
			if err := rows.Scan(&userID, &rebateAmount); err != nil {
				return err
			}
			updatedCount++
			creditsByUser[userID] += rebateAmount
		}
		if err := rows.Err(); err != nil {
			return err
		}

		for userID, amount := range creditsByUser {
			if amount <= 0 {
				continue
			}
			if _, err := ensureUserAffiliateWithClient(txCtx, txClient, userID); err != nil {
				return err
			}
			if _, err := txClient.ExecContext(txCtx,
				"UPDATE user_affiliates SET aff_quota = aff_quota + $1, aff_history_quota = aff_history_quota + $1, updated_at = NOW() WHERE user_id = $2",
				amount, userID,
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return updatedCount, nil
}

func (r *affiliateRepository) CreateWithdrawalRequest(ctx context.Context, userID int64, amount float64, applicantNote string) (*service.AffiliateWithdrawalRequest, error) {
	if userID <= 0 || amount <= 0 {
		return nil, service.ErrAffiliateWithdrawAmount
	}

	var out *service.AffiliateWithdrawalRequest
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		if _, err := ensureUserAffiliateWithClient(txCtx, txClient, userID); err != nil {
			return err
		}

		type allocatedRebate struct {
			ID     int64
			Amount float64
		}
		allocations := make([]allocatedRebate, 0)
		remaining := amount
		rows, err := txClient.QueryContext(txCtx, `
SELECT id, rebate_amount::double precision
FROM affiliate_rebate_records
WHERE user_id = $1
  AND status = 'available'
ORDER BY available_at ASC, id ASC
FOR UPDATE
`, userID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var rebateID int64
			var rebateAmount float64
			if err := rows.Scan(&rebateID, &rebateAmount); err != nil {
				return err
			}
			if remaining <= 0 {
				break
			}
			useAmount := math.Min(remaining, rebateAmount)
			if useAmount <= 0 {
				continue
			}
			allocations = append(allocations, allocatedRebate{ID: rebateID, Amount: useAmount})
			remaining -= useAmount
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if remaining > 1e-9 {
			return service.ErrAffiliateWithdrawAmount
		}

		res, err := txClient.ExecContext(txCtx,
			"UPDATE user_affiliates SET aff_quota = aff_quota - $1, updated_at = NOW() WHERE user_id = $2 AND aff_quota >= $1",
			amount, userID,
		)
		if err != nil {
			return err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return service.ErrAffiliateWithdrawAmount
		}

		withdrawalRows, err := txClient.QueryContext(txCtx, `
INSERT INTO affiliate_withdrawal_requests (
  user_id,
  amount,
  status,
  applicant_note,
  admin_note,
  created_at,
  updated_at
)
VALUES ($1, $2, 'pending', $3, '', NOW(), NOW())
RETURNING id, user_id, amount, status, applicant_note, admin_note, reviewed_by, reviewed_at, paid_at, created_at, updated_at
`, userID, amount, applicantNote)
		if err != nil {
			return err
		}
		if !withdrawalRows.Next() {
			if err := withdrawalRows.Err(); err != nil {
				_ = withdrawalRows.Close()
				return err
			}
			return errors.New("failed to create affiliate withdrawal request")
		}

		item := &service.AffiliateWithdrawalRequest{}
		var reviewedBy sql.NullInt64
		var reviewedAt sql.NullTime
		var paidAt sql.NullTime
		if err := withdrawalRows.Scan(
			&item.ID,
			&item.UserID,
			&item.Amount,
			&item.Status,
			&item.ApplicantNote,
			&item.AdminNote,
			&reviewedBy,
			&reviewedAt,
			&paidAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return err
		}
		if reviewedBy.Valid {
			item.ReviewedBy = &reviewedBy.Int64
		}
		if reviewedAt.Valid {
			item.ReviewedAt = &reviewedAt.Time
		}
		if paidAt.Valid {
			item.PaidAt = &paidAt.Time
		}
		if err := withdrawalRows.Close(); err != nil {
			return err
		}
		for _, allocation := range allocations {
			if _, err := txClient.ExecContext(txCtx, `
INSERT INTO affiliate_withdrawal_request_items (withdrawal_request_id, rebate_record_id, amount, created_at)
VALUES ($1, $2, $3, NOW())
`, item.ID, allocation.ID, allocation.Amount); err != nil {
				return err
			}
			if _, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_rebate_records
SET status = 'withdraw_requested',
    updated_at = NOW()
WHERE id = $1
`, allocation.ID); err != nil {
				return err
			}
		}
		out = item
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *affiliateRepository) ListWithdrawalRequests(ctx context.Context, status string, limit int) ([]service.AffiliateWithdrawalRequest, error) {
	if limit <= 0 {
		limit = 100
	}
	client := clientFromContext(ctx, r.client)
	query := `
SELECT id, user_id, amount::double precision, status, applicant_note, admin_note, reviewed_by, reviewed_at, paid_at, created_at, updated_at
FROM affiliate_withdrawal_requests`
	args := make([]any, 0, 2)
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
		query += " ORDER BY created_at DESC LIMIT $2"
		args = append(args, limit)
	} else {
		query += " ORDER BY created_at DESC LIMIT $1"
		args = append(args, limit)
	}
	rows, err := client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.AffiliateWithdrawalRequest, 0)
	for rows.Next() {
		item, err := scanAffiliateWithdrawalRequest(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *affiliateRepository) ListUserWithdrawalRequests(ctx context.Context, userID int64, limit int) ([]service.AffiliateWithdrawalRequest, error) {
	if limit <= 0 {
		limit = 100
	}
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
SELECT id, user_id, amount::double precision, status, applicant_note, admin_note, reviewed_by, reviewed_at, paid_at, created_at, updated_at
FROM affiliate_withdrawal_requests
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.AffiliateWithdrawalRequest, 0)
	for rows.Next() {
		item, err := scanAffiliateWithdrawalRequest(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *affiliateRepository) ListUserRebateRecords(ctx context.Context, userID int64, limit int) ([]service.AffiliateRebateRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
SELECT affiliate_rebate_records.id,
       affiliate_rebate_records.source_order_id,
       affiliate_rebate_records.user_id,
       affiliate_rebate_records.source_user_id,
       COALESCE(u.email, ''),
       COALESCE(u.username, ''),
       affiliate_rebate_records.level,
       affiliate_rebate_records.rate::double precision,
       affiliate_rebate_records.base_amount::double precision,
       affiliate_rebate_records.rebate_amount::double precision,
       affiliate_rebate_records.status,
       affiliate_rebate_records.available_at,
       affiliate_rebate_records.created_at,
       affiliate_rebate_records.updated_at
FROM affiliate_rebate_records
LEFT JOIN users u ON u.id = affiliate_rebate_records.source_user_id
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.AffiliateRebateRecord, 0)
	for rows.Next() {
		item := service.AffiliateRebateRecord{}
		var availableAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.SourceOrderID,
			&item.UserID,
			&item.SourceUserID,
			&item.SourceEmail,
			&item.SourceUsername,
			&item.Level,
			&item.Rate,
			&item.BaseAmount,
			&item.RebateAmount,
			&item.Status,
			&availableAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if availableAt.Valid {
			item.AvailableAt = &availableAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *affiliateRepository) RejectWithdrawalRequest(ctx context.Context, requestID int64, reviewerID int64, adminNote string) (*service.AffiliateWithdrawalRequest, error) {
	return r.reviewWithdrawalRequest(ctx, requestID, reviewerID, adminNote, "rejected", true)
}

func (r *affiliateRepository) MarkWithdrawalPaid(ctx context.Context, requestID int64, reviewerID int64, adminNote string) (*service.AffiliateWithdrawalRequest, error) {
	return r.reviewWithdrawalRequest(ctx, requestID, reviewerID, adminNote, "paid", false)
}

func (r *affiliateRepository) ReverseRebatesForOrder(ctx context.Context, sourceOrderID int64) error {
	return r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		rows, err := txClient.QueryContext(txCtx, `
SELECT id,
       user_id,
       rebate_amount::double precision,
       status
FROM affiliate_rebate_records
WHERE source_order_id = $1
FOR UPDATE
`, sourceOrderID)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		type rebateRow struct {
			recordID     int64
			userID       int64
			rebateAmount float64
			status       string
		}
		items := make([]rebateRow, 0)
		for rows.Next() {
			var item rebateRow
			if err := rows.Scan(&item.recordID, &item.userID, &item.rebateAmount, &item.status); err != nil {
				return err
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		availableDeduct := map[int64]float64{}
		debtAdd := map[int64]float64{}
		requestDeduct := map[int64]float64{}

		for _, item := range items {
			switch item.status {
			case service.AffiliateRebateStatusPending:
				if _, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_rebate_records
SET status = 'cancelled',
    updated_at = NOW()
WHERE id = $1
`, item.recordID); err != nil {
					return err
				}
			case "available":
				if _, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_rebate_records
SET status = 'reversed',
    reversed_amount = rebate_amount,
    updated_at = NOW()
WHERE id = $1
`, item.recordID); err != nil {
					return err
				}
				availableDeduct[item.userID] += item.rebateAmount
			case "withdraw_requested":
				if _, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_rebate_records
SET status = 'reversed',
    reversed_amount = rebate_amount,
    updated_at = NOW()
WHERE id = $1
`, item.recordID); err != nil {
					return err
				}
				requestRows, err := txClient.QueryContext(txCtx, `
SELECT withdrawal_request_id
FROM affiliate_withdrawal_request_items
WHERE rebate_record_id = $1
`, item.recordID)
				if err != nil {
					return err
				}
				for requestRows.Next() {
					var withdrawalRequestID int64
					if err := requestRows.Scan(&withdrawalRequestID); err != nil {
						_ = requestRows.Close()
						return err
					}
					requestDeduct[withdrawalRequestID] += item.rebateAmount
				}
				if err := requestRows.Close(); err != nil {
					return err
				}
			case "withdraw_paid":
				if _, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_rebate_records
SET status = 'debt_offset',
    reversed_amount = rebate_amount,
    debt_amount = rebate_amount,
    updated_at = NOW()
WHERE id = $1
`, item.recordID); err != nil {
					return err
				}
				debtAdd[item.userID] += item.rebateAmount
			}
		}

		for userID, amount := range availableDeduct {
			if amount <= 0 {
				continue
			}
			if _, err := txClient.ExecContext(txCtx, `
UPDATE user_affiliates
SET aff_quota = GREATEST(aff_quota - $1, 0),
    updated_at = NOW()
WHERE user_id = $2
`, amount, userID); err != nil {
				return err
			}
		}

		for withdrawalRequestID, amount := range requestDeduct {
			if amount <= 0 {
				continue
			}
			if _, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_withdrawal_requests
SET amount = GREATEST(amount - $1, 0),
    status = CASE WHEN amount - $1 <= 0 THEN 'rejected' ELSE status END,
    admin_note = CASE WHEN amount - $1 <= 0 THEN CONCAT(COALESCE(admin_note, ''), CASE WHEN COALESCE(admin_note, '') = '' THEN '' ELSE E'\n' END, '[system] auto adjusted by refund') ELSE admin_note END,
    reviewed_at = CASE WHEN amount - $1 <= 0 THEN NOW() ELSE reviewed_at END,
    updated_at = NOW()
WHERE id = $2
  AND status = 'pending'
`, amount, withdrawalRequestID); err != nil {
				return err
			}
		}

		for userID, amount := range debtAdd {
			if amount <= 0 {
				continue
			}
			if _, err := ensureUserAffiliateWithClient(txCtx, txClient, userID); err != nil {
				return err
			}
			if _, err := txClient.ExecContext(txCtx, `
UPDATE user_affiliates
SET debt_quota = debt_quota + $1,
    updated_at = NOW()
WHERE user_id = $2
`, amount, userID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *affiliateRepository) reviewWithdrawalRequest(ctx context.Context, requestID int64, reviewerID int64, adminNote string, targetStatus string, refundQuota bool) (*service.AffiliateWithdrawalRequest, error) {
	var out *service.AffiliateWithdrawalRequest
	err := r.withTx(ctx, func(txCtx context.Context, txClient *dbent.Client) error {
		rows, err := txClient.QueryContext(txCtx, `
SELECT id, user_id, amount::double precision, status, applicant_note, admin_note, reviewed_by, reviewed_at, paid_at, created_at, updated_at
FROM affiliate_withdrawal_requests
WHERE id = $1
FOR UPDATE
`, requestID)
		if err != nil {
			return err
		}
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				_ = rows.Close()
				return err
			}
			return service.ErrAffiliateWithdrawalNotFound
		}
		item, err := scanAffiliateWithdrawalRequest(rows)
		if err != nil {
			_ = rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
		if item.Status != "pending" {
			return service.ErrAffiliateWithdrawalStatus
		}

		itemRows, err := txClient.QueryContext(txCtx, `
SELECT rebate_record_id, amount::double precision
FROM affiliate_withdrawal_request_items
WHERE withdrawal_request_id = $1
ORDER BY id ASC
`, requestID)
		if err != nil {
			return err
		}
		type requestItem struct {
			rebateRecordID int64
			amount         float64
		}
		requestItems := make([]requestItem, 0)
		for itemRows.Next() {
			var rebateRecordID int64
			var amount float64
			if err := itemRows.Scan(&rebateRecordID, &amount); err != nil {
				_ = itemRows.Close()
				return err
			}
			requestItems = append(requestItems, requestItem{rebateRecordID: rebateRecordID, amount: amount})
		}
		if err := itemRows.Close(); err != nil {
			return err
		}

		refundableAmount := 0.0
		if refundQuota {
			for _, requestItem := range requestItems {
				rows3, err := txClient.QueryContext(txCtx, `
SELECT status
FROM affiliate_rebate_records
WHERE id = $1
`, requestItem.rebateRecordID)
				if err != nil {
					return err
				}
				if rows3.Next() {
					var status string
					if err := rows3.Scan(&status); err != nil {
						_ = rows3.Close()
						return err
					}
					if status == "withdraw_requested" {
						refundableAmount += requestItem.amount
					}
				}
				if err := rows3.Close(); err != nil {
					return err
				}
			}
			if refundableAmount > 0 {
				if _, err := ensureUserAffiliateWithClient(txCtx, txClient, item.UserID); err != nil {
					return err
				}
				if _, err := txClient.ExecContext(txCtx,
					"UPDATE user_affiliates SET aff_quota = aff_quota + $1, updated_at = NOW() WHERE user_id = $2",
					refundableAmount, item.UserID,
				); err != nil {
					return err
				}
			}
		}

		recordStatus := "withdraw_paid"
		if refundQuota {
			recordStatus = "available"
		}
		for _, requestItem := range requestItems {
			if _, err := txClient.ExecContext(txCtx, `
UPDATE affiliate_rebate_records
SET status = $1,
    updated_at = NOW()
WHERE id = $2
`, recordStatus, requestItem.rebateRecordID); err != nil {
				return err
			}
		}

		updateSQL := `
UPDATE affiliate_withdrawal_requests
SET status = $1,
    admin_note = $2,
    reviewed_by = $3,
    reviewed_at = NOW(),
    updated_at = NOW()`
		args := []any{targetStatus, adminNote, reviewerID}
		if targetStatus == "paid" {
			updateSQL += ", paid_at = NOW()"
		}
		updateSQL += " WHERE id = $4"
		args = append(args, requestID)
		if _, err := txClient.ExecContext(txCtx, updateSQL, args...); err != nil {
			return err
		}

		rows2, err := txClient.QueryContext(txCtx, `
SELECT id, user_id, amount::double precision, status, applicant_note, admin_note, reviewed_by, reviewed_at, paid_at, created_at, updated_at
FROM affiliate_withdrawal_requests
WHERE id = $1
`, requestID)
		if err != nil {
			return err
		}
		defer func() { _ = rows2.Close() }()
		if !rows2.Next() {
			if err := rows2.Err(); err != nil {
				return err
			}
			return service.ErrAffiliateWithdrawalNotFound
		}
		out, err = scanAffiliateWithdrawalRequest(rows2)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func scanAffiliateWithdrawalRequest(rows *sql.Rows) (*service.AffiliateWithdrawalRequest, error) {
	item := &service.AffiliateWithdrawalRequest{}
	var reviewedBy sql.NullInt64
	var reviewedAt sql.NullTime
	var paidAt sql.NullTime
	if err := rows.Scan(
		&item.ID,
		&item.UserID,
		&item.Amount,
		&item.Status,
		&item.ApplicantNote,
		&item.AdminNote,
		&reviewedBy,
		&reviewedAt,
		&paidAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if reviewedBy.Valid {
		item.ReviewedBy = &reviewedBy.Int64
	}
	if reviewedAt.Valid {
		item.ReviewedAt = &reviewedAt.Time
	}
	if paidAt.Valid {
		item.PaidAt = &paidAt.Time
	}
	return item, nil
}

func (r *affiliateRepository) withTx(ctx context.Context, fn func(txCtx context.Context, txClient *dbent.Client) error) error {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return fn(ctx, tx.Client())
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin affiliate transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	if err := fn(txCtx, tx.Client()); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit affiliate transaction: %w", err)
	}
	return nil
}

func ensureUserAffiliateWithClient(ctx context.Context, client affiliateQueryExecer, userID int64) (*service.AffiliateSummary, error) {
	summary, err := queryAffiliateByUserID(ctx, client, userID)
	if err == nil {
		return summary, nil
	}
	if !errors.Is(err, service.ErrAffiliateProfileNotFound) {
		return nil, err
	}

	for i := 0; i < affiliateCodeMaxAttempts; i++ {
		code, codeErr := generateAffiliateCode()
		if codeErr != nil {
			return nil, codeErr
		}
		_, insertErr := client.ExecContext(ctx, `
INSERT INTO user_affiliates (user_id, aff_code, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (user_id) DO NOTHING`, userID, code)
		if insertErr == nil {
			break
		}
		if isAffiliateUniqueViolation(insertErr) {
			continue
		}
		return nil, insertErr
	}

	return queryAffiliateByUserID(ctx, client, userID)
}

func queryAffiliateByUserID(ctx context.Context, client affiliateQueryExecer, userID int64) (*service.AffiliateSummary, error) {
	rows, err := client.QueryContext(ctx, `
SELECT user_id,
       aff_code,
       inviter_id,
       aff_count,
       aff_quota::double precision,
       aff_history_quota::double precision,
       debt_quota::double precision,
       created_at,
       updated_at
FROM user_affiliates
WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrAffiliateProfileNotFound
	}

	var out service.AffiliateSummary
	var inviterID sql.NullInt64
	if err := rows.Scan(
		&out.UserID,
		&out.AffCode,
		&inviterID,
		&out.AffCount,
		&out.AffQuota,
		&out.AffHistoryQuota,
		&out.DebtQuota,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if inviterID.Valid {
		out.InviterID = &inviterID.Int64
	}
	return &out, nil
}

func queryAffiliateByCode(ctx context.Context, client affiliateQueryExecer, code string) (*service.AffiliateSummary, error) {
	rows, err := client.QueryContext(ctx, `
SELECT user_id,
       aff_code,
       inviter_id,
       aff_count,
       aff_quota::double precision,
       aff_history_quota::double precision,
       debt_quota::double precision,
       created_at,
       updated_at
FROM user_affiliates
WHERE aff_code = $1
LIMIT 1`, strings.ToUpper(strings.TrimSpace(code)))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrAffiliateProfileNotFound
	}

	var out service.AffiliateSummary
	var inviterID sql.NullInt64
	if err := rows.Scan(
		&out.UserID,
		&out.AffCode,
		&inviterID,
		&out.AffCount,
		&out.AffQuota,
		&out.AffHistoryQuota,
		&out.DebtQuota,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if inviterID.Valid {
		out.InviterID = &inviterID.Int64
	}
	return &out, nil
}

func queryUserBalance(ctx context.Context, client affiliateQueryExecer, userID int64) (float64, error) {
	rows, err := client.QueryContext(ctx,
		"SELECT balance::double precision FROM users WHERE id = $1 LIMIT 1",
		userID,
	)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return 0, service.ErrUserNotFound
	}
	var balance float64
	if err := rows.Scan(&balance); err != nil {
		return 0, err
	}
	return balance, nil
}

func generateAffiliateCode() (string, error) {
	buf := make([]byte, affiliateCodeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate affiliate code: %w", err)
	}
	for i := range buf {
		buf[i] = affiliateCodeCharset[int(buf[i])%len(affiliateCodeCharset)]
	}
	return string(buf), nil
}

func isAffiliateUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == "23505"
	}
	return false
}
