package service

import (
	"context"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const AffiliateRebateStatusPending = "pending"

var affiliateLevelRates = []float64{6, 3, 1}

func (s *AffiliateService) ResolveAffiliateAncestors(ctx context.Context, userID int64, maxDepth int) ([]AffiliateAncestor, error) {
	if s == nil || s.repo == nil || userID <= 0 || maxDepth <= 0 {
		return nil, nil
	}
	return s.repo.ListAncestors(ctx, userID, maxDepth)
}

func (s *AffiliateService) CreatePendingRebatesForOrder(ctx context.Context, sourceOrderID, sourceUserID int64, baseAmount float64, paidAt time.Time) (float64, error) {
	if s == nil || s.repo == nil || sourceOrderID <= 0 || sourceUserID <= 0 || baseAmount <= 0 {
		return 0, nil
	}

	ancestors, err := s.ResolveAffiliateAncestors(ctx, sourceUserID, len(affiliateLevelRates))
	if err != nil || len(ancestors) == 0 {
		return 0, err
	}

	availableAt := paidAt.Add(7 * 24 * time.Hour)
	records := make([]AffiliateRebateRecordInput, 0, len(ancestors))
	total := 0.0
	for _, ancestor := range ancestors {
		if ancestor.Level <= 0 || ancestor.Level > len(affiliateLevelRates) || ancestor.UserID <= 0 {
			continue
		}
		rate := affiliateLevelRates[ancestor.Level-1]
		rebateAmount := roundTo(baseAmount*(rate/100), 8)
		if rebateAmount <= 0 {
			continue
		}
		total += rebateAmount
		records = append(records, AffiliateRebateRecordInput{
			UserID:        ancestor.UserID,
			SourceUserID:  sourceUserID,
			SourceOrderID: sourceOrderID,
			Level:         ancestor.Level,
			Rate:          rate,
			BaseAmount:    baseAmount,
			RebateAmount:  rebateAmount,
			Status:        AffiliateRebateStatusPending,
			AvailableAt:   availableAt,
		})
	}
	if len(records) == 0 {
		return 0, nil
	}
	if _, err := s.repo.CreatePendingRebateRecords(ctx, records); err != nil {
		return 0, err
	}
	return roundTo(total, 8), nil
}

func (s *AffiliateService) ReleaseDueRebates(ctx context.Context, now time.Time) (int, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	return s.repo.ReleaseDuePendingRebateRecords(ctx, now)
}

func (s *AffiliateService) CreateWithdrawalRequest(ctx context.Context, userID int64, amount float64, applicantNote string) (*AffiliateWithdrawalRequest, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	if userID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_USER", "invalid user")
	}
	if amount <= 0 {
		return nil, ErrAffiliateWithdrawAmount
	}
	summary, err := s.EnsureUserAffiliate(ctx, userID)
	if err != nil {
		return nil, err
	}
	if summary.DebtQuota > 0 {
		return nil, ErrAffiliateDebtOutstanding
	}
	if summary.AffQuota < 100 {
		return nil, ErrAffiliateWithdrawThreshold
	}
	if amount < 100 || amount > summary.AffQuota {
		return nil, ErrAffiliateWithdrawAmount
	}
	return s.repo.CreateWithdrawalRequest(ctx, userID, amount, strings.TrimSpace(applicantNote))
}

func (s *AffiliateService) ListWithdrawalRequests(ctx context.Context, status string, limit int) ([]AffiliateWithdrawalRequest, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	return s.repo.ListWithdrawalRequests(ctx, strings.TrimSpace(status), limit)
}

func (s *AffiliateService) ListUserWithdrawalRequests(ctx context.Context, userID int64, limit int) ([]AffiliateWithdrawalRequest, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	if userID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_USER", "invalid user")
	}
	return s.repo.ListUserWithdrawalRequests(ctx, userID, limit)
}

func (s *AffiliateService) RejectWithdrawalRequest(ctx context.Context, requestID, reviewerID int64, adminNote string) (*AffiliateWithdrawalRequest, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	if requestID <= 0 || reviewerID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "invalid withdrawal review input")
	}
	return s.repo.RejectWithdrawalRequest(ctx, requestID, reviewerID, strings.TrimSpace(adminNote))
}

func (s *AffiliateService) MarkWithdrawalPaid(ctx context.Context, requestID, reviewerID int64, adminNote string) (*AffiliateWithdrawalRequest, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.ServiceUnavailable("SERVICE_UNAVAILABLE", "affiliate service unavailable")
	}
	if requestID <= 0 || reviewerID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "invalid withdrawal review input")
	}
	return s.repo.MarkWithdrawalPaid(ctx, requestID, reviewerID, strings.TrimSpace(adminNote))
}

func (s *AffiliateService) ReverseRebatesForOrder(ctx context.Context, sourceOrderID int64) error {
	if s == nil || s.repo == nil || sourceOrderID <= 0 {
		return nil
	}
	return s.repo.ReverseRebatesForOrder(ctx, sourceOrderID)
}
