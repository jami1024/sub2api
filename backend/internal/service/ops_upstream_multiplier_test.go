package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestMeasureUpstreamMultipliersSkipsUnsupportedModelAndStoresSample(t *testing.T) {
	var saved []*OpsUpstreamMultiplierSample
	opsRepo := &opsRepoMock{
		InsertUpstreamMultiplierSampleFn: func(ctx context.Context, input *OpsUpstreamMultiplierSample) (*OpsUpstreamMultiplierSample, error) {
			saved = append(saved, input)
			input.ID = int64(len(saved))
			return input, nil
		},
	}
	accountRepo := &opsUpstreamMultiplierAccountRepo{
		accounts: []Account{
			{
				ID:       11,
				Name:     "normal upstream",
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Status:   StatusActive,
				Credentials: map[string]any{
					"api_key":       "sk-test-secret",
					"base_url":      "https://upstream.example.com",
					"model_mapping": map[string]any{"gpt-4.1": "gpt-4.1"},
				},
			},
		},
	}
	svc := NewOpsService(opsRepo, nil, &config.Config{}, accountRepo, nil, nil, nil, nil, nil, nil, nil)

	resp, err := svc.MeasureUpstreamMultipliers(context.Background(), OpsMeasureUpstreamMultiplierRequest{Model: "gpt-5.4"})
	require.NoError(t, err)
	require.Len(t, resp.Samples, 1)
	require.Equal(t, OpsUpstreamMultiplierStatusSkipped, resp.Samples[0].Status)
	require.Contains(t, resp.Samples[0].ErrorMessage, "model_mapping")
	require.Len(t, saved, 1)
	require.Equal(t, "sk-test-", saved[0].KeyPrefixSnapshot)
	require.NotContains(t, saved[0].KeyPrefixSnapshot, "secret")
}

func TestMeasureUpstreamMultipliersCalculatesMultiplierFromUsageDelta(t *testing.T) {
	usageCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer sk-live-secret", r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/v1/usage":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"mode": "unrestricted",
				"usage": map[string]any{
					"total": map[string]any{
						"cost":        []float64{1.00, 1.10}[min(usageCalls, 1)],
						"actual_cost": []float64{0.50, 0.512}[min(usageCalls, 1)],
					},
				},
			})
			usageCalls++
		case "/v1/usage/stats":
			http.NotFound(w, r)
		case "/v1/chat/completions":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":      "chatcmpl_test",
				"object":  "chat.completion",
				"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": "ok"}}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var saved []*OpsUpstreamMultiplierSample
	opsRepo := &opsRepoMock{
		InsertUpstreamMultiplierSampleFn: func(ctx context.Context, input *OpsUpstreamMultiplierSample) (*OpsUpstreamMultiplierSample, error) {
			saved = append(saved, input)
			input.ID = int64(len(saved))
			return input, nil
		},
	}
	accountRepo := &opsUpstreamMultiplierAccountRepo{
		accounts: []Account{
			{
				ID:       12,
				Name:     "priced upstream",
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Status:   StatusActive,
				Credentials: map[string]any{
					"api_key":       "sk-live-secret",
					"base_url":      server.URL,
					"model_mapping": map[string]any{"gpt-5.4": "gpt-5.4"},
				},
			},
		},
	}
	svc := NewOpsService(opsRepo, nil, &config.Config{}, accountRepo, nil, nil, nil, nil, nil, nil, nil)
	svc.upstreamMultiplierHTTPClient = server.Client()
	svc.upstreamMultiplierPollAttempts = 1

	resp, err := svc.MeasureUpstreamMultipliers(context.Background(), OpsMeasureUpstreamMultiplierRequest{Model: "gpt-5.4"})
	require.NoError(t, err)
	require.Len(t, resp.Samples, 1)
	require.Equal(t, OpsUpstreamMultiplierStatusSuccess, resp.Samples[0].Status)
	require.InDelta(t, 0.10, *resp.Samples[0].StandardCostDelta, 0.000001)
	require.InDelta(t, 0.012, *resp.Samples[0].ActualCostDelta, 0.000001)
	require.InDelta(t, 0.12, *resp.Samples[0].Multiplier, 0.000001)
	require.Len(t, saved, 1)
	require.Equal(t, "sk-live-", saved[0].KeyPrefixSnapshot)
}

func TestApplyLatestUpstreamMultiplierUpdatesAccountRateMultiplier(t *testing.T) {
	latestMultiplier := 0.06
	var requestedModel string
	var requestedAccountIDs []int64
	opsRepo := &opsRepoMock{
		GetLatestUpstreamMultiplierSamplesFn: func(ctx context.Context, model string, accountIDs []int64) (map[int64]*OpsUpstreamMultiplierSample, error) {
			requestedModel = model
			requestedAccountIDs = append([]int64(nil), accountIDs...)
			return map[int64]*OpsUpstreamMultiplierSample{
				12: {
					ID:         99,
					AccountID:  12,
					Model:      model,
					Status:     OpsUpstreamMultiplierStatusSuccess,
					Multiplier: &latestMultiplier,
					MeasuredAt: time.Now().UTC(),
				},
			}, nil
		},
	}
	accountRepo := &opsUpstreamMultiplierAccountRepo{
		accounts: []Account{{ID: 12, Name: "priced upstream", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive}},
	}
	svc := NewOpsService(opsRepo, nil, &config.Config{}, accountRepo, nil, nil, nil, nil, nil, nil, nil)

	resp, err := svc.ApplyLatestUpstreamMultiplier(context.Background(), OpsApplyUpstreamMultiplierRequest{Model: "gpt-5.4", AccountID: 12})
	require.NoError(t, err)

	require.Equal(t, "gpt-5.4", requestedModel)
	require.Equal(t, []int64{12}, requestedAccountIDs)
	require.Equal(t, int64(12), resp.AccountID)
	require.InDelta(t, 0.06, resp.RateMultiplier, 0.000001)
	require.NotNil(t, resp.Sample)
	require.Equal(t, []int64{12}, accountRepo.bulkUpdateIDs)
	require.NotNil(t, accountRepo.bulkUpdate.RateMultiplier)
	require.InDelta(t, 0.06, *accountRepo.bulkUpdate.RateMultiplier, 0.000001)
}

func TestApplyLatestUpstreamMultiplierRejectsMissingSuccessfulSample(t *testing.T) {
	opsRepo := &opsRepoMock{
		GetLatestUpstreamMultiplierSamplesFn: func(ctx context.Context, model string, accountIDs []int64) (map[int64]*OpsUpstreamMultiplierSample, error) {
			return map[int64]*OpsUpstreamMultiplierSample{
				12: {AccountID: 12, Model: model, Status: OpsUpstreamMultiplierStatusError},
			}, nil
		},
	}
	accountRepo := &opsUpstreamMultiplierAccountRepo{
		accounts: []Account{{ID: 12, Name: "priced upstream", Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive}},
	}
	svc := NewOpsService(opsRepo, nil, &config.Config{}, accountRepo, nil, nil, nil, nil, nil, nil, nil)

	resp, err := svc.ApplyLatestUpstreamMultiplier(context.Background(), OpsApplyUpstreamMultiplierRequest{Model: "gpt-5.4", AccountID: 12})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Empty(t, accountRepo.bulkUpdateIDs)
}

func TestSanitizeUpstreamMultiplierErrorMessageRedactsSecrets(t *testing.T) {
	msg := sanitizeUpstreamMultiplierErrorMessage("upstream rejected Authorization: Bearer sk-leaky-secret-token-123456 and key-live-secret-abcdef")

	require.NotContains(t, msg, "sk-leaky-secret-token")
	require.NotContains(t, msg, "key-live-secret")
	require.Contains(t, msg, "[redacted]")
}

type opsUpstreamMultiplierAccountRepo struct {
	accounts      []Account
	bulkUpdateIDs []int64
	bulkUpdate    AccountBulkUpdate
}

func (r *opsUpstreamMultiplierAccountRepo) Create(ctx context.Context, account *Account) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) GetByID(ctx context.Context, id int64) (*Account, error) {
	for i := range r.accounts {
		if r.accounts[i].ID == id {
			return &r.accounts[i], nil
		}
	}
	return nil, ErrAccountNotFound
}

func (r *opsUpstreamMultiplierAccountRepo) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	result := make([]*Account, 0, len(ids))
	for _, id := range ids {
		acc, err := r.GetByID(ctx, id)
		if err == nil {
			result = append(result, acc)
		}
	}
	return result, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ExistsByID(ctx context.Context, id int64) (bool, error) {
	_, err := r.GetByID(ctx, id)
	return err == nil, nil
}

func (r *opsUpstreamMultiplierAccountRepo) GetByCRSAccountID(ctx context.Context, crsAccountID string) (*Account, error) {
	return nil, nil
}

func (r *opsUpstreamMultiplierAccountRepo) FindByExtraField(ctx context.Context, key string, value any) ([]Account, error) {
	return nil, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListCRSAccountIDs(ctx context.Context) (map[string]int64, error) {
	return nil, nil
}

func (r *opsUpstreamMultiplierAccountRepo) Update(ctx context.Context, account *Account) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) List(ctx context.Context, params pagination.PaginationParams) ([]Account, *pagination.PaginationResult, error) {
	return r.accounts, &pagination.PaginationResult{Total: int64(len(r.accounts)), Page: 1, PageSize: len(r.accounts)}, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListWithFilters(ctx context.Context, params pagination.PaginationParams, platform, accountType, status, search string, groupID int64, privacyMode string) ([]Account, *pagination.PaginationResult, error) {
	result := make([]Account, 0, len(r.accounts))
	for _, acc := range r.accounts {
		if platform != "" && acc.Platform != platform {
			continue
		}
		if accountType != "" && acc.Type != accountType {
			continue
		}
		if status != "" && acc.Status != status {
			continue
		}
		result = append(result, acc)
	}
	return result, &pagination.PaginationResult{Total: int64(len(result)), Page: 1, PageSize: len(result)}, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListByGroup(ctx context.Context, groupID int64) ([]Account, error) {
	return nil, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListActive(ctx context.Context) ([]Account, error) {
	result := make([]Account, 0, len(r.accounts))
	for _, acc := range r.accounts {
		if acc.Status == StatusActive {
			result = append(result, acc)
		}
	}
	return result, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListByPlatform(ctx context.Context, platform string) ([]Account, error) {
	result := make([]Account, 0, len(r.accounts))
	for _, acc := range r.accounts {
		if acc.Platform == platform {
			result = append(result, acc)
		}
	}
	return result, nil
}

func (r *opsUpstreamMultiplierAccountRepo) UpdateLastUsed(ctx context.Context, id int64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) BatchUpdateLastUsed(ctx context.Context, updates map[int64]time.Time) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) SetError(ctx context.Context, id int64, errorMsg string) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) ClearError(ctx context.Context, id int64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) SetSchedulable(ctx context.Context, id int64, schedulable bool) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) AutoPauseExpiredAccounts(ctx context.Context, now time.Time) (int64, error) {
	return 0, nil
}

func (r *opsUpstreamMultiplierAccountRepo) BindGroups(ctx context.Context, accountID int64, groupIDs []int64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListSchedulable(ctx context.Context) ([]Account, error) {
	return r.accounts, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListSchedulableByGroupID(ctx context.Context, groupID int64) ([]Account, error) {
	return r.accounts, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListSchedulableByPlatform(ctx context.Context, platform string) ([]Account, error) {
	return r.ListByPlatform(ctx, platform)
}

func (r *opsUpstreamMultiplierAccountRepo) ListSchedulableByGroupIDAndPlatform(ctx context.Context, groupID int64, platform string) ([]Account, error) {
	return r.ListByPlatform(ctx, platform)
}

func (r *opsUpstreamMultiplierAccountRepo) ListSchedulableByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	return r.accounts, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListSchedulableByGroupIDAndPlatforms(ctx context.Context, groupID int64, platforms []string) ([]Account, error) {
	return r.accounts, nil
}

func (r *opsUpstreamMultiplierAccountRepo) ListSchedulableUngroupedByPlatform(ctx context.Context, platform string) ([]Account, error) {
	return r.ListByPlatform(ctx, platform)
}

func (r *opsUpstreamMultiplierAccountRepo) ListSchedulableUngroupedByPlatforms(ctx context.Context, platforms []string) ([]Account, error) {
	return r.accounts, nil
}

func (r *opsUpstreamMultiplierAccountRepo) SetRateLimited(ctx context.Context, id int64, resetAt time.Time) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) SetModelRateLimit(ctx context.Context, id int64, scope string, resetAt time.Time, reason ...string) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) SetOverloaded(ctx context.Context, id int64, until time.Time) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) SetTempUnschedulable(ctx context.Context, id int64, until time.Time, reason string) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) ClearTempUnschedulable(ctx context.Context, id int64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) ClearRateLimit(ctx context.Context, id int64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) ClearAntigravityQuotaScopes(ctx context.Context, id int64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) ClearModelRateLimits(ctx context.Context, id int64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) UpdateSessionWindow(ctx context.Context, id int64, start, end *time.Time, status string) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) UpdateSessionWindowEnd(ctx context.Context, id int64, end time.Time) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) UpdateExtra(ctx context.Context, id int64, updates map[string]any) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) BulkUpdate(ctx context.Context, ids []int64, updates AccountBulkUpdate) (int64, error) {
	r.bulkUpdateIDs = append([]int64(nil), ids...)
	r.bulkUpdate = updates
	for i := range r.accounts {
		for _, id := range ids {
			if r.accounts[i].ID == id && updates.RateMultiplier != nil {
				v := *updates.RateMultiplier
				r.accounts[i].RateMultiplier = &v
			}
		}
	}
	return int64(len(ids)), nil
}

func (r *opsUpstreamMultiplierAccountRepo) IncrementQuotaUsed(ctx context.Context, id int64, amount float64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) ResetQuotaUsed(ctx context.Context, id int64) error {
	return nil
}

func (r *opsUpstreamMultiplierAccountRepo) RevertProxyFallback(ctx context.Context, accountID int64) error {
	return nil
}

var _ AccountRepository = (*opsUpstreamMultiplierAccountRepo)(nil)
