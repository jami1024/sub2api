package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	opsUpstreamMultiplierDefaultModel        = "gpt-5.4"
	opsUpstreamMultiplierDefaultLimit        = 200
	opsUpstreamMultiplierMaxHistoryLimit     = 500
	opsUpstreamMultiplierDefaultPollAttempts = 3
	opsUpstreamMultiplierDefaultPollInterval = 2 * time.Second
	opsUpstreamMultiplierHTTPTimeout         = 90 * time.Second
)

func (s *OpsService) ListUpstreamMultiplierAccounts(ctx context.Context, model string) (*OpsUpstreamMultiplierAccountsResponse, error) {
	model = normalizeOpsUpstreamMultiplierModel(model)
	accounts, err := s.loadUpstreamMultiplierCandidateAccounts(ctx, nil)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, account.ID)
	}
	latest := map[int64]*OpsUpstreamMultiplierSample{}
	if s.opsRepo != nil && len(ids) > 0 {
		latest, err = s.opsRepo.GetLatestUpstreamMultiplierSamples(ctx, model, ids)
		if err != nil {
			return nil, err
		}
	}

	items := make([]*OpsUpstreamMultiplierAccount, 0, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		supported, reason := isAccountSupportedForUpstreamMultiplier(account, model)
		items = append(items, &OpsUpstreamMultiplierAccount{
			AccountID:    account.ID,
			AccountName:  account.Name,
			Platform:     account.Platform,
			BaseURL:      account.GetOpenAIBaseURL(),
			KeyPrefix:    keyPrefix(account.GetOpenAIApiKey()),
			Supported:    supported,
			SkipReason:   reason,
			LatestSample: latest[account.ID],
		})
	}
	return &OpsUpstreamMultiplierAccountsResponse{Model: model, Accounts: items}, nil
}

func (s *OpsService) ListUpstreamMultiplierSamples(ctx context.Context, filter *OpsUpstreamMultiplierSamplesFilter) (*OpsUpstreamMultiplierSamplesResponse, error) {
	if s == nil || s.opsRepo == nil {
		return &OpsUpstreamMultiplierSamplesResponse{Model: normalizeOpsUpstreamMultiplierModel("")}, nil
	}
	if filter == nil {
		filter = &OpsUpstreamMultiplierSamplesFilter{}
	}
	normalized := *filter
	normalized.Model = normalizeOpsUpstreamMultiplierModel(filter.Model)
	if normalized.Limit <= 0 {
		normalized.Limit = opsUpstreamMultiplierDefaultLimit
	}
	if normalized.Limit > opsUpstreamMultiplierMaxHistoryLimit {
		normalized.Limit = opsUpstreamMultiplierMaxHistoryLimit
	}
	samples, err := s.opsRepo.ListUpstreamMultiplierSamples(ctx, &normalized)
	if err != nil {
		return nil, err
	}
	return &OpsUpstreamMultiplierSamplesResponse{Model: normalized.Model, Samples: samples}, nil
}

func (s *OpsService) MeasureUpstreamMultipliers(ctx context.Context, req OpsMeasureUpstreamMultiplierRequest) (*OpsMeasureUpstreamMultiplierResponse, error) {
	model := normalizeOpsUpstreamMultiplierModel(req.Model)
	if s == nil || s.opsRepo == nil || s.accountRepo == nil {
		return &OpsMeasureUpstreamMultiplierResponse{Model: model}, nil
	}

	accounts, err := s.loadUpstreamMultiplierCandidateAccounts(ctx, req.AccountIDs)
	if err != nil {
		return nil, err
	}
	samples := make([]*OpsUpstreamMultiplierSample, 0, len(accounts))
	for i := range accounts {
		account := &accounts[i]
		sample := s.measureOneUpstreamMultiplier(ctx, account, model)
		saved, err := s.opsRepo.InsertUpstreamMultiplierSample(ctx, sample)
		if err != nil {
			return nil, err
		}
		samples = append(samples, saved)
	}
	return &OpsMeasureUpstreamMultiplierResponse{Model: model, Samples: samples}, nil
}

func (s *OpsService) loadUpstreamMultiplierCandidateAccounts(ctx context.Context, accountIDs []int64) ([]Account, error) {
	if s == nil || s.accountRepo == nil {
		return nil, nil
	}
	var accounts []Account
	if len(accountIDs) > 0 {
		got, err := s.accountRepo.GetByIDs(ctx, accountIDs)
		if err != nil {
			return nil, err
		}
		accounts = make([]Account, 0, len(got))
		for _, acc := range got {
			if acc != nil {
				accounts = append(accounts, *acc)
			}
		}
	} else {
		got, _, err := s.accountRepo.ListWithFilters(
			ctx,
			pagination.PaginationParams{Page: 1, PageSize: 10000},
			PlatformOpenAI,
			AccountTypeAPIKey,
			StatusActive,
			"",
			0,
			"",
		)
		if err != nil {
			return nil, err
		}
		accounts = got
	}

	filtered := make([]Account, 0, len(accounts))
	for _, account := range accounts {
		if account.Platform != PlatformOpenAI || account.Type != AccountTypeAPIKey || account.Status != StatusActive {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(account.Name))
		if strings.Contains(name, "生图") || strings.Contains(name, "自建") {
			continue
		}
		filtered = append(filtered, account)
	}
	return filtered, nil
}

func (s *OpsService) measureOneUpstreamMultiplier(ctx context.Context, account *Account, model string) *OpsUpstreamMultiplierSample {
	now := time.Now().UTC()
	sample := &OpsUpstreamMultiplierSample{
		AccountID:           account.ID,
		AccountNameSnapshot: account.Name,
		Platform:            account.Platform,
		BaseURLSnapshot:     account.GetOpenAIBaseURL(),
		KeyPrefixSnapshot:   keyPrefix(account.GetOpenAIApiKey()),
		Model:               model,
		MeasuredAt:          now,
		CreatedAt:           now,
	}

	if ok, reason := isAccountSupportedForUpstreamMultiplier(account, model); !ok {
		sample.Status = OpsUpstreamMultiplierStatusSkipped
		sample.ErrorMessage = reason
		return sample
	}

	before, status, err := s.fetchUpstreamUsageTotals(ctx, account)
	if status != nil {
		sample.HTTPStatus = status
	}
	if err != nil {
		sample.Status = OpsUpstreamMultiplierStatusError
		sample.ErrorMessage = truncateString(err.Error(), 500)
		return sample
	}
	sample.BalanceBefore = before.Balance

	if status, err := s.sendMultiplierProbeRequest(ctx, account, model); err != nil {
		if status != nil {
			sample.HTTPStatus = status
		}
		sample.Status = OpsUpstreamMultiplierStatusError
		sample.ErrorMessage = truncateString(err.Error(), 500)
		return sample
	}

	after, status, err := s.pollUpstreamUsageAfterProbe(ctx, account, before)
	if status != nil {
		sample.HTTPStatus = status
	}
	if err != nil {
		sample.Status = OpsUpstreamMultiplierStatusError
		sample.ErrorMessage = truncateString(err.Error(), 500)
		return sample
	}
	sample.BalanceAfter = after.Balance

	standardDelta := roundFloat(after.TotalCost-before.TotalCost, 12)
	actualDelta := roundFloat(after.TotalActualCost-before.TotalActualCost, 12)
	sample.StandardCostDelta = &standardDelta
	sample.ActualCostDelta = &actualDelta
	if standardDelta <= 0 {
		sample.Status = OpsUpstreamMultiplierStatusError
		sample.ErrorMessage = "usage delta is zero; upstream usage may be delayed"
		return sample
	}
	multiplier := roundFloat(actualDelta/standardDelta, 12)
	sample.Multiplier = &multiplier
	sample.Status = OpsUpstreamMultiplierStatusSuccess
	return sample
}

func isAccountSupportedForUpstreamMultiplier(account *Account, model string) (bool, string) {
	if account == nil {
		return false, "account is empty"
	}
	if account.Platform != PlatformOpenAI || account.Type != AccountTypeAPIKey || account.Status != StatusActive {
		return false, "only active OpenAI API Key accounts are supported"
	}
	if strings.TrimSpace(account.GetOpenAIApiKey()) == "" {
		return false, "api_key is empty"
	}
	if strings.TrimSpace(account.GetOpenAIBaseURL()) == "" {
		return false, "base_url is empty"
	}
	mapping := account.GetModelMapping()
	if len(mapping) == 0 {
		return false, "model_mapping does not declare target model"
	}
	if mappingSupportsRequestedModel(mapping, model) {
		return true, ""
	}
	for _, upstreamModel := range mapping {
		if strings.TrimSpace(upstreamModel) == model {
			return true, ""
		}
	}
	return false, fmt.Sprintf("model_mapping does not include %s", model)
}

type upstreamUsageTotals struct {
	TotalCost       float64
	TotalActualCost float64
	Balance         *float64
}

func (s *OpsService) fetchUpstreamUsageTotals(ctx context.Context, account *Account) (*upstreamUsageTotals, *int, error) {
	endpoint := buildOpenAIEndpointURL(account.GetOpenAIBaseURL(), "/v1/usage")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+account.GetOpenAIApiKey())
	resp, err := s.upstreamMultiplierClient().Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	status := resp.StatusCode
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &status, fmt.Errorf("usage endpoint returned %d", resp.StatusCode)
	}
	totals, err := parseUpstreamUsageTotals(body)
	if err != nil {
		return nil, &status, err
	}
	return totals, &status, nil
}

func (s *OpsService) sendMultiplierProbeRequest(ctx context.Context, account *Account, model string) (*int, error) {
	endpoint := buildOpenAIEndpointURL(account.GetOpenAIBaseURL(), "/v1/chat/completions")
	payload := map[string]any{
		"model":      model,
		"messages":   []map[string]string{{"role": "user", "content": "hi"}},
		"max_tokens": 1,
		"stream":     false,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+account.GetOpenAIApiKey())
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.upstreamMultiplierClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	status := resp.StatusCode
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &status, fmt.Errorf("probe request returned %d", resp.StatusCode)
	}
	return &status, nil
}

func (s *OpsService) pollUpstreamUsageAfterProbe(ctx context.Context, account *Account, before *upstreamUsageTotals) (*upstreamUsageTotals, *int, error) {
	attempts := s.upstreamMultiplierPollAttempts
	if attempts <= 0 {
		attempts = opsUpstreamMultiplierDefaultPollAttempts
	}
	interval := s.upstreamMultiplierPollInterval
	if interval <= 0 {
		interval = opsUpstreamMultiplierDefaultPollInterval
	}
	var last *upstreamUsageTotals
	var lastStatus *int
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, lastStatus, ctx.Err()
			case <-time.After(interval):
			}
		}
		after, status, err := s.fetchUpstreamUsageTotals(ctx, account)
		last = after
		lastStatus = status
		lastErr = err
		if err == nil && after.TotalCost > before.TotalCost {
			return after, status, nil
		}
	}
	if lastErr != nil {
		return nil, lastStatus, lastErr
	}
	if last == nil {
		return nil, lastStatus, fmt.Errorf("usage endpoint returned no data")
	}
	return last, lastStatus, nil
}

func (s *OpsService) upstreamMultiplierClient() *http.Client {
	if s != nil && s.upstreamMultiplierHTTPClient != nil {
		return s.upstreamMultiplierHTTPClient
	}
	return &http.Client{Timeout: opsUpstreamMultiplierHTTPTimeout}
}

func parseUpstreamUsageTotals(body []byte) (*upstreamUsageTotals, error) {
	var payload struct {
		TotalCost       *float64 `json:"total_cost"`
		TotalActualCost *float64 `json:"total_actual_cost"`
		Balance         *float64 `json:"balance"`
		Usage           *struct {
			Total *struct {
				Cost       *float64 `json:"cost"`
				ActualCost *float64 `json:"actual_cost"`
			} `json:"total"`
		} `json:"usage"`
		Data *struct {
			TotalCost       *float64 `json:"total_cost"`
			TotalActualCost *float64 `json:"total_actual_cost"`
			Balance         *float64 `json:"balance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode usage response: %w", err)
	}
	if payload.Data != nil {
		payload.TotalCost = payload.Data.TotalCost
		payload.TotalActualCost = payload.Data.TotalActualCost
		payload.Balance = payload.Data.Balance
	}
	if payload.Usage != nil && payload.Usage.Total != nil {
		if payload.TotalCost == nil {
			payload.TotalCost = payload.Usage.Total.Cost
		}
		if payload.TotalActualCost == nil {
			payload.TotalActualCost = payload.Usage.Total.ActualCost
		}
	}
	if payload.TotalCost == nil || payload.TotalActualCost == nil {
		return nil, fmt.Errorf("usage response missing total_cost or total_actual_cost")
	}
	return &upstreamUsageTotals{
		TotalCost:       *payload.TotalCost,
		TotalActualCost: *payload.TotalActualCost,
		Balance:         payload.Balance,
	}, nil
}

func normalizeOpsUpstreamMultiplierModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return opsUpstreamMultiplierDefaultModel
	}
	return model
}

func keyPrefix(key string) string {
	key = strings.TrimSpace(key)
	if len(key) <= 8 {
		return key
	}
	return key[:8]
}

func roundFloat(v float64, places int) float64 {
	if places <= 0 {
		return math.Round(v)
	}
	pow := math.Pow10(places)
	return math.Round(v*pow) / pow
}
