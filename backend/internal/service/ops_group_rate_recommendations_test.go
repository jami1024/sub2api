package service

import (
	"context"
	"math"
	"testing"
	"time"
)

func floatPtr(v float64) *float64 { return &v }

func TestGetGroupRateRecommendationsUsesCheapestPackageAndMultipleAccounts(t *testing.T) {
	now := time.Now().UTC()
	repo := &opsRepoMock{
		GetGroupRateRecommendationSourceDataFn: func(ctx context.Context, filter *OpsGroupRateRecommendationFilter) (*OpsGroupRateRecommendationSourceData, error) {
			if filter.Model != "gpt-5.4" {
				t.Fatalf("model = %s, want gpt-5.4", filter.Model)
			}
			if filter.ProfitMargin != 0.2 || filter.SafetyFactor != 1.2 || filter.UsageDays != 7 {
				t.Fatalf("unexpected normalized filter: %#v", filter)
			}
			return &OpsGroupRateRecommendationSourceData{
				Packages: []*OpsGroupRateRecommendationPackageBasis{
					{PackageID: 1, Name: "50", Price: 50, CreditAmount: 110, PackageScope: "codex", RevenuePerCredit: 50.0 / 110.0},
					{PackageID: 2, Name: "100", Price: 100, CreditAmount: 400, PackageScope: "codex", RevenuePerCredit: 0.25},
				},
				Groups: []*OpsGroupRateRecommendationSourceGroup{
					{
						GroupID: 8, GroupName: "gpt pro 高价", RateMultiplier: 1.3, PackageScope: "codex",
						Accounts: []*OpsGroupRateRecommendationSourceAccount{
							{AccountID: 9, AccountName: "天才程序员", Status: StatusActive, Schedulable: true, CurrentPriority: 1, BindingPriority: 1, BaseURL: "https://api.dzzzz.cf", KeyPrefix: "sk-f642"},
							{AccountID: 18, AccountName: "xixi 高速", Status: StatusActive, Schedulable: true, CurrentPriority: 10, BindingPriority: 1, BaseURL: "https://xixiapi.cc", KeyPrefix: "sk-b773"},
						},
					},
				},
				Usage: map[int64]map[int64]OpsGroupRateRecommendationUsageShare{
					8: {
						9:  {RequestCount: 70, RequestShare: 0.7, StandardCost: 70, StandardCostShare: 0.7},
						18: {RequestCount: 30, RequestShare: 0.3, StandardCost: 30, StandardCostShare: 0.3},
					},
				},
				Samples: map[int64]*OpsUpstreamMultiplierSample{
					9:  {AccountID: 9, Status: OpsUpstreamMultiplierStatusSuccess, Multiplier: floatPtr(0.135), MeasuredAt: now},
					18: {AccountID: 18, Status: OpsUpstreamMultiplierStatusSuccess, Multiplier: floatPtr(0.18), MeasuredAt: now},
				},
			}, nil
		},
	}
	svc := NewOpsService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	resp, err := svc.GetGroupRateRecommendations(context.Background(), &OpsGroupRateRecommendationFilter{PackageScope: "codex", ProfitMargin: 0.2, SafetyFactor: 1.2, UsageDays: 7, Model: "gpt-5.4"})
	if err != nil {
		t.Fatalf("GetGroupRateRecommendations error: %v", err)
	}
	if resp.PackageBasis == nil || resp.PackageBasis.PackageID != 2 {
		t.Fatalf("package basis = %#v, want package 2", resp.PackageBasis)
	}
	if len(resp.Groups) != 1 {
		t.Fatalf("groups len = %d, want 1", len(resp.Groups))
	}
	group := resp.Groups[0]
	assertClose(t, derefFloat(group.ActualBlendedMultiplier), 0.1485)
	assertClose(t, derefFloat(group.WorstCaseMultiplier), 0.18)
	assertClose(t, derefFloat(group.MinimumGroupMultiplier), 0.891)
	assertClose(t, derefFloat(group.SafeGroupMultiplier), 1.08)
	if group.Status != OpsGroupRateRecommendationStatusSafe {
		t.Fatalf("status = %s, want safe", group.Status)
	}
	if len(group.Accounts) != 2 {
		t.Fatalf("accounts len = %d, want 2", len(group.Accounts))
	}
	if group.Accounts[0].RecommendedWeight <= group.Accounts[1].RecommendedWeight {
		t.Fatalf("cheaper account should have higher recommended weight: %#v", group.Accounts)
	}
}

func TestGetGroupRateRecommendationsMarksInsufficientWhenMissingSamples(t *testing.T) {
	repo := &opsRepoMock{
		GetGroupRateRecommendationSourceDataFn: func(ctx context.Context, filter *OpsGroupRateRecommendationFilter) (*OpsGroupRateRecommendationSourceData, error) {
			return &OpsGroupRateRecommendationSourceData{
				Packages: []*OpsGroupRateRecommendationPackageBasis{{PackageID: 2, Name: "100", Price: 100, CreditAmount: 400, PackageScope: "codex", RevenuePerCredit: 0.25}},
				Groups:   []*OpsGroupRateRecommendationSourceGroup{{GroupID: 2, GroupName: "gpt pro", RateMultiplier: 1, PackageScope: "codex", Accounts: []*OpsGroupRateRecommendationSourceAccount{{AccountID: 1, AccountName: "xixi", Status: StatusActive, Schedulable: true, CurrentPriority: 1}}}},
				Usage:    map[int64]map[int64]OpsGroupRateRecommendationUsageShare{},
				Samples:  map[int64]*OpsUpstreamMultiplierSample{},
			}, nil
		},
	}
	svc := NewOpsService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	resp, err := svc.GetGroupRateRecommendations(context.Background(), &OpsGroupRateRecommendationFilter{PackageScope: "codex"})
	if err != nil {
		t.Fatalf("GetGroupRateRecommendations error: %v", err)
	}
	if got := resp.Groups[0].Status; got != OpsGroupRateRecommendationStatusInsufficient {
		t.Fatalf("status = %s, want insufficient_data", got)
	}
	if len(resp.Groups[0].Notes) == 0 {
		t.Fatalf("expected missing sample note")
	}
}

func TestGetGroupRateRecommendationsSkipsImageGroupsAndSelfHostedByDefault(t *testing.T) {
	repo := &opsRepoMock{
		GetGroupRateRecommendationSourceDataFn: func(ctx context.Context, filter *OpsGroupRateRecommendationFilter) (*OpsGroupRateRecommendationSourceData, error) {
			return &OpsGroupRateRecommendationSourceData{
				Packages: []*OpsGroupRateRecommendationPackageBasis{{PackageID: 1, Name: "pkg", Price: 100, CreditAmount: 400, PackageScope: "codex", RevenuePerCredit: 0.25}},
				Groups: []*OpsGroupRateRecommendationSourceGroup{
					{GroupID: 1, GroupName: "gpt 生图", AllowImageGeneration: true, Accounts: []*OpsGroupRateRecommendationSourceAccount{{AccountID: 1, AccountName: "xixi", Status: StatusActive, Schedulable: true}}},
					{GroupID: 2, GroupName: "gpt pro", RateMultiplier: 1, Accounts: []*OpsGroupRateRecommendationSourceAccount{{AccountID: 2, AccountName: "自建线路", Status: StatusActive, Schedulable: true}}},
				},
				Usage:   map[int64]map[int64]OpsGroupRateRecommendationUsageShare{},
				Samples: map[int64]*OpsUpstreamMultiplierSample{2: {AccountID: 2, Status: OpsUpstreamMultiplierStatusSuccess, Multiplier: floatPtr(0.01), MeasuredAt: time.Now().UTC()}},
			}, nil
		},
	}
	svc := NewOpsService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	resp, err := svc.GetGroupRateRecommendations(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetGroupRateRecommendations error: %v", err)
	}
	if len(resp.Groups) != 1 || resp.Groups[0].GroupName != "gpt pro" {
		t.Fatalf("groups = %#v, want only gpt pro", resp.Groups)
	}
	if resp.Groups[0].Accounts[0].ParticipatesInAdvice {
		t.Fatalf("self-hosted account should not participate by default")
	}
}

func assertClose(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.000001 {
		t.Fatalf("got %.12f, want %.12f", got, want)
	}
}

func derefFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
