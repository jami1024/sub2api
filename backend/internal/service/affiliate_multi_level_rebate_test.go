package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type affiliateWorkflowSettingRepoStub struct {
	values map[string]string
}

func (s *affiliateWorkflowSettingRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *affiliateWorkflowSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if v, ok := s.values[key]; ok {
		return v, nil
	}
	return "", ErrSettingNotFound
}

func (s *affiliateWorkflowSettingRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *affiliateWorkflowSettingRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if v, ok := s.values[key]; ok {
			out[key] = v
		}
	}
	return out, nil
}

func (s *affiliateWorkflowSettingRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *affiliateWorkflowSettingRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *affiliateWorkflowSettingRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

type affiliateWorkflowRepoStub struct {
	summaries     map[int64]*AffiliateSummary
	records       []AffiliateRebateRecordInput
	rebateRecords []AffiliateRebateRecord
	released      int
	withdrawn     *AffiliateWithdrawalRequest
	withdrawList  []AffiliateWithdrawalRequest
}

func (s *affiliateWorkflowRepoStub) EnsureUserAffiliate(ctx context.Context, userID int64) (*AffiliateSummary, error) {
	if summary, ok := s.summaries[userID]; ok {
		return summary, nil
	}
	return nil, ErrAffiliateProfileNotFound
}
func (s *affiliateWorkflowRepoStub) GetAffiliateByCode(ctx context.Context, code string) (*AffiliateSummary, error) {
	panic("unexpected GetAffiliateByCode call")
}
func (s *affiliateWorkflowRepoStub) BindInviter(ctx context.Context, userID, inviterID int64) (bool, error) {
	panic("unexpected BindInviter call")
}
func (s *affiliateWorkflowRepoStub) AccrueQuota(ctx context.Context, inviterID, inviteeUserID int64, amount float64, freezeHours int) (bool, error) {
	panic("unexpected AccrueQuota call")
}
func (s *affiliateWorkflowRepoStub) TransferQuotaToBalance(ctx context.Context, userID int64) (float64, float64, error) {
	panic("unexpected TransferQuotaToBalance call")
}
func (s *affiliateWorkflowRepoStub) ListInvitees(ctx context.Context, inviterID int64, limit int) ([]AffiliateInvitee, error) {
	panic("unexpected ListInvitees call")
}
func (s *affiliateWorkflowRepoStub) ListAncestors(ctx context.Context, userID int64, maxDepth int) ([]AffiliateAncestor, error) {
	current := userID
	visited := map[int64]struct{}{userID: {}}
	out := make([]AffiliateAncestor, 0, maxDepth)
	for level := 1; level <= maxDepth; level++ {
		summary, ok := s.summaries[current]
		if !ok || summary.InviterID == nil {
			break
		}
		inviterID := *summary.InviterID
		if _, exists := visited[inviterID]; exists {
			break
		}
		visited[inviterID] = struct{}{}
		out = append(out, AffiliateAncestor{UserID: inviterID, Level: level})
		current = inviterID
	}
	return out, nil
}
func (s *affiliateWorkflowRepoStub) CreatePendingRebateRecords(ctx context.Context, records []AffiliateRebateRecordInput) (int, error) {
	s.records = append(s.records, records...)
	return len(records), nil
}
func (s *affiliateWorkflowRepoStub) ReleaseDuePendingRebateRecords(ctx context.Context, now time.Time) (int, error) {
	return s.released, nil
}
func (s *affiliateWorkflowRepoStub) CreateWithdrawalRequest(ctx context.Context, userID int64, amount float64, applicantNote string) (*AffiliateWithdrawalRequest, error) {
	if s.withdrawn != nil {
		return s.withdrawn, nil
	}
	return &AffiliateWithdrawalRequest{ID: 1, UserID: userID, Amount: amount, Status: "pending", ApplicantNote: applicantNote}, nil
}
func (s *affiliateWorkflowRepoStub) ListWithdrawalRequests(ctx context.Context, status string, limit int) ([]AffiliateWithdrawalRequest, error) {
	return s.withdrawList, nil
}
func (s *affiliateWorkflowRepoStub) ListUserWithdrawalRequests(ctx context.Context, userID int64, limit int) ([]AffiliateWithdrawalRequest, error) {
	return s.withdrawList, nil
}
func (s *affiliateWorkflowRepoStub) RejectWithdrawalRequest(ctx context.Context, requestID int64, reviewerID int64, adminNote string) (*AffiliateWithdrawalRequest, error) {
	return &AffiliateWithdrawalRequest{ID: requestID, UserID: 1, Amount: 120, Status: "rejected", AdminNote: adminNote}, nil
}
func (s *affiliateWorkflowRepoStub) MarkWithdrawalPaid(ctx context.Context, requestID int64, reviewerID int64, adminNote string) (*AffiliateWithdrawalRequest, error) {
	return &AffiliateWithdrawalRequest{ID: requestID, UserID: 1, Amount: 120, Status: "paid", AdminNote: adminNote}, nil
}
func (s *affiliateWorkflowRepoStub) ReverseRebatesForOrder(ctx context.Context, sourceOrderID int64) error {
	return nil
}
func (s *affiliateWorkflowRepoStub) SumPendingRebateByUser(ctx context.Context, userID int64) (float64, error) {
	if summary, ok := s.summaries[userID]; ok {
		return summary.PendingQuota, nil
	}
	return 0, nil
}
func (s *affiliateWorkflowRepoStub) ListUserRebateRecords(ctx context.Context, userID int64, limit int) ([]AffiliateRebateRecord, error) {
	return s.rebateRecords, nil
}

func (s *affiliateWorkflowRepoStub) GetAccruedRebateFromInvitee(ctx context.Context, inviterID, inviteeUserID int64) (float64, error) {
	return 0, nil
}
func (s *affiliateWorkflowRepoStub) ThawFrozenQuota(ctx context.Context, userID int64) (float64, error) {
	return 0, nil
}
func (s *affiliateWorkflowRepoStub) UpdateUserAffCode(ctx context.Context, userID int64, newCode string) error {
	return nil
}
func (s *affiliateWorkflowRepoStub) ResetUserAffCode(ctx context.Context, userID int64) (string, error) {
	return "", nil
}
func (s *affiliateWorkflowRepoStub) SetUserRebateRate(ctx context.Context, userID int64, ratePercent *float64) error {
	return nil
}
func (s *affiliateWorkflowRepoStub) BatchSetUserRebateRate(ctx context.Context, userIDs []int64, ratePercent *float64) error {
	return nil
}
func (s *affiliateWorkflowRepoStub) ListUsersWithCustomSettings(ctx context.Context, filter AffiliateAdminFilter) ([]AffiliateAdminEntry, int64, error) {
	return nil, 0, nil
}

func TestCreatePendingRebatesForOrderSkipsWhenAffiliateDisabled(t *testing.T) {
	ctx := context.Background()
	inviterID, buyerID := int64(1), int64(2)
	repo := &affiliateWorkflowRepoStub{
		summaries: map[int64]*AffiliateSummary{
			inviterID: {UserID: inviterID},
			buyerID:   {UserID: buyerID, InviterID: &inviterID},
		},
	}
	settingSvc := NewSettingService(&affiliateWorkflowSettingRepoStub{values: map[string]string{
		SettingKeyAffiliateEnabled: "false",
	}}, &config.Config{})
	svc := &AffiliateService{repo: repo, settingService: settingSvc}

	total, err := svc.CreatePendingRebatesForOrder(ctx, 99, buyerID, 100, time.Now())

	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, repo.records)
}

func TestCreatePendingRebatesForOrderCreatesThreeLevels(t *testing.T) {
	ctx := context.Background()
	aID, bID, cID, dID := int64(1), int64(2), int64(3), int64(4)
	repo := &affiliateWorkflowRepoStub{
		summaries: map[int64]*AffiliateSummary{
			aID: {UserID: aID},
			bID: {UserID: bID, InviterID: &aID},
			cID: {UserID: cID, InviterID: &bID},
			dID: {UserID: dID, InviterID: &cID},
		},
	}
	svc := &AffiliateService{repo: repo}
	paidAt := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

	total, err := svc.CreatePendingRebatesForOrder(ctx, 99, dID, 100, paidAt)
	require.NoError(t, err)
	require.InDelta(t, 10.0, total, 1e-9)
	require.Len(t, repo.records, 3)

	require.Equal(t, cID, repo.records[0].UserID)
	require.Equal(t, 1, repo.records[0].Level)
	require.InDelta(t, 6.0, repo.records[0].RebateAmount, 1e-9)
	require.Equal(t, AffiliateRebateStatusPending, repo.records[0].Status)
	require.Equal(t, paidAt.Add(7*24*time.Hour), repo.records[0].AvailableAt)

	require.Equal(t, bID, repo.records[1].UserID)
	require.Equal(t, 2, repo.records[1].Level)
	require.InDelta(t, 3.0, repo.records[1].RebateAmount, 1e-9)

	require.Equal(t, aID, repo.records[2].UserID)
	require.Equal(t, 3, repo.records[2].Level)
	require.InDelta(t, 1.0, repo.records[2].RebateAmount, 1e-9)
}

func TestCreatePendingRebatesForOrderEveryPurchaseGeneratesNewRecords(t *testing.T) {
	ctx := context.Background()
	aID, bID := int64(1), int64(2)
	repo := &affiliateWorkflowRepoStub{
		summaries: map[int64]*AffiliateSummary{
			aID: {UserID: aID},
			bID: {UserID: bID, InviterID: &aID},
		},
	}
	svc := &AffiliateService{repo: repo}

	_, err := svc.CreatePendingRebatesForOrder(ctx, 1001, bID, 30, time.Now())
	require.NoError(t, err)
	_, err = svc.CreatePendingRebatesForOrder(ctx, 1002, bID, 30, time.Now())
	require.NoError(t, err)

	require.Len(t, repo.records, 2)
	require.NotEqual(t, repo.records[0].SourceOrderID, repo.records[1].SourceOrderID)
}

func TestReleaseDueAffiliateRebates(t *testing.T) {
	ctx := context.Background()
	repo := &affiliateWorkflowRepoStub{released: 2}
	svc := &AffiliateService{repo: repo}

	count, err := svc.ReleaseDueRebates(ctx, time.Now())
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestCreateAffiliateWithdrawalRequestRequiresMin100(t *testing.T) {
	ctx := context.Background()
	userID := int64(11)
	repo := &affiliateWorkflowRepoStub{
		summaries: map[int64]*AffiliateSummary{
			userID: {UserID: userID, AffQuota: 99},
		},
	}
	svc := &AffiliateService{repo: repo}

	_, err := svc.CreateWithdrawalRequest(ctx, userID, 99, "withdraw")
	require.ErrorIs(t, err, ErrAffiliateWithdrawThreshold)

	repo.summaries[userID].AffQuota = 180
	item, err := svc.CreateWithdrawalRequest(ctx, userID, 120, "withdraw")
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, 120.0, item.Amount)
	require.Equal(t, "pending", item.Status)

	repo.summaries[userID].DebtQuota = 1
	_, err = svc.CreateWithdrawalRequest(ctx, userID, 120, "withdraw")
	require.ErrorIs(t, err, ErrAffiliateDebtOutstanding)
}

func TestReviewAffiliateWithdrawalRequest(t *testing.T) {
	ctx := context.Background()
	repo := &affiliateWorkflowRepoStub{}
	svc := &AffiliateService{repo: repo}

	rejected, err := svc.RejectWithdrawalRequest(ctx, 7, 99, "reject")
	require.NoError(t, err)
	require.Equal(t, "rejected", rejected.Status)

	paid, err := svc.MarkWithdrawalPaid(ctx, 8, 99, "paid")
	require.NoError(t, err)
	require.Equal(t, "paid", paid.Status)
}
