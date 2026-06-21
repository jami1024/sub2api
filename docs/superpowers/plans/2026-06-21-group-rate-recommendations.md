# Group Rate Recommendations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an admin-only read-only panel that recommends OpenAI group rate multipliers and upstream account traffic weights from package pricing, measured upstream multipliers, and recent group/account usage.

**Architecture:** Add one Ops repository aggregation endpoint that loads package basis, OpenAI groups, group account bindings, latest upstream multiplier samples, and recent usage shares in efficient SQL queries. Keep recommendation math in service layer for testability. Extend the existing Provider Status admin page with a new `GroupRateRecommendationsPanel` below the upstream multiplier monitor.

**Tech Stack:** Go service/repository/handler with PostgreSQL, Gin admin routes, Vue 3 + TypeScript frontend, Vitest frontend tests, Go unit tests.

---

## File Structure

Create:
- `backend/internal/service/ops_group_rate_recommendations.go` — normalizes request params and computes package-based rate/weight recommendations.
- `backend/internal/service/ops_group_rate_recommendations_test.go` — service math and edge-case tests.
- `backend/internal/repository/ops_repo_group_rate_recommendations.go` — repository SQL for package basis, groups/accounts, latest samples, and usage shares.
- `backend/internal/repository/ops_repo_group_rate_recommendations_test.go` — repository integration-style tests using existing repository test helpers if available.
- `frontend/src/views/admin/ops/components/GroupRateRecommendationsPanel.vue` — admin UI panel.
- `frontend/src/views/admin/ops/components/__tests__/GroupRateRecommendationsPanel.spec.ts` — component rendering tests.

Modify:
- `backend/internal/service/ops_dashboard_models.go` — add DTOs for recommendation params, package basis, group rows, account rows, and response.
- `backend/internal/service/ops_port.go` — add repository method signatures.
- `backend/internal/service/ops_repo_mock_test.go` — add mock function fields for new repository methods.
- `backend/internal/handler/admin/ops_dashboard_handler.go` — add handler parsing query parameters.
- `backend/internal/server/routes/admin.go` — register `GET /admin/ops/group-rate-recommendations`.
- `frontend/src/api/admin/ops.ts` — add response types and API method.
- `frontend/src/views/admin/ops/ProviderStatusView.vue` — load recommendations and render the new panel.
- `frontend/src/views/admin/ops/__tests__/ProviderStatusView.spec.ts` — include the new API mock and verify initial loading.

Do not modify:
- Billing calculation.
- Actual group/account priority update APIs.
- Upstream probing behavior.
- User-facing pages.

---

### Task 1: Add backend DTOs and repository contract

**Files:**
- Modify: `backend/internal/service/ops_dashboard_models.go`
- Modify: `backend/internal/service/ops_port.go`
- Modify: `backend/internal/service/ops_repo_mock_test.go`

- [ ] **Step 1: Add DTO definitions**

Append these types after existing upstream multiplier DTOs in `backend/internal/service/ops_dashboard_models.go`:

```go
const (
	OpsGroupRateRecommendationStatusSafe       = "safe"
	OpsGroupRateRecommendationStatusBasicSafe  = "basic_safe"
	OpsGroupRateRecommendationStatusLow        = "low"
	OpsGroupRateRecommendationStatusInsufficient = "insufficient_data"
)

type OpsGroupRateRecommendationFilter struct {
	Model                string
	PackageScope         string
	ProfitMargin         float64
	SafetyFactor         float64
	UsageDays            int
	IncludeUnschedulable bool
	IncludeSelfHosted    bool
}

type OpsGroupRateRecommendationPackageBasis struct {
	PackageID        int64   `json:"package_id"`
	Name             string  `json:"name"`
	Price            float64 `json:"price"`
	CreditAmount     float64 `json:"credit_amount"`
	PackageScope     string  `json:"package_scope"`
	RevenuePerCredit float64 `json:"revenue_per_credit"`
}

type OpsGroupRateRecommendationUsageShare struct {
	RequestCount     int64   `json:"request_count"`
	RequestShare     float64 `json:"request_share"`
	StandardCost     float64 `json:"standard_cost"`
	StandardCostShare float64 `json:"standard_cost_share"`
}

type OpsGroupRateRecommendationAccount struct {
	AccountID             int64   `json:"account_id"`
	AccountName           string  `json:"account_name"`
	BaseURL               string  `json:"base_url"`
	KeyPrefix             string  `json:"key_prefix"`
	Schedulable           bool    `json:"schedulable"`
	Status                string  `json:"status"`
	CurrentPriority       int     `json:"current_priority"`
	BindingPriority       int     `json:"binding_priority"`
	UpstreamMultiplier    *float64 `json:"upstream_multiplier,omitempty"`
	MultiplierStatus      string  `json:"multiplier_status"`
	MultiplierMeasuredAt  *time.Time `json:"multiplier_measured_at,omitempty"`
	RequestCount          int64   `json:"request_count"`
	RequestShare          float64 `json:"request_share"`
	StandardCost          float64 `json:"standard_cost"`
	StandardCostShare     float64 `json:"standard_cost_share"`
	RecommendedWeight     float64 `json:"recommended_weight"`
	RecommendedPriority   int     `json:"recommended_priority"`
	ParticipatesInAdvice  bool    `json:"participates_in_advice"`
	Note                  string  `json:"note"`
}

type OpsGroupRateRecommendationGroup struct {
	GroupID                      int64   `json:"group_id"`
	GroupName                    string  `json:"group_name"`
	CurrentGroupMultiplier       float64 `json:"current_group_multiplier"`
	PackageScope                 string  `json:"package_scope"`
	SchedulableAccountCount      int     `json:"schedulable_account_count"`
	ActualBlendedMultiplier      *float64 `json:"actual_blended_multiplier,omitempty"`
	RecommendedBlendedMultiplier *float64 `json:"recommended_blended_multiplier,omitempty"`
	WorstCaseMultiplier          *float64 `json:"worst_case_multiplier,omitempty"`
	MinimumGroupMultiplier       *float64 `json:"minimum_group_multiplier,omitempty"`
	SafeGroupMultiplier          *float64 `json:"safe_group_multiplier,omitempty"`
	Status                       string  `json:"status"`
	Notes                        []string `json:"notes,omitempty"`
	Accounts                     []*OpsGroupRateRecommendationAccount `json:"accounts"`
}

type OpsGroupRateRecommendationResponse struct {
	Params       OpsGroupRateRecommendationFilter        `json:"params"`
	PackageBasis *OpsGroupRateRecommendationPackageBasis `json:"package_basis,omitempty"`
	Groups       []*OpsGroupRateRecommendationGroup      `json:"groups"`
}

type OpsGroupRateRecommendationSourceData struct {
	Packages []*OpsGroupRateRecommendationPackageBasis
	Groups   []*OpsGroupRateRecommendationSourceGroup
	Usage    map[int64]map[int64]OpsGroupRateRecommendationUsageShare
	Samples  map[int64]*OpsUpstreamMultiplierSample
}

type OpsGroupRateRecommendationSourceGroup struct {
	GroupID                int64
	GroupName              string
	RateMultiplier         float64
	PackageScope           string
	AllowImageGeneration   bool
	Accounts               []*OpsGroupRateRecommendationSourceAccount
}

type OpsGroupRateRecommendationSourceAccount struct {
	AccountID       int64
	AccountName     string
	Platform         string
	Type             string
	Status           string
	Schedulable      bool
	CurrentPriority  int
	BindingPriority  int
	BaseURL          string
	KeyPrefix        string
}
```

If `time` is not currently imported by this file, change its import from:

```go
import "time"
```

to a parenthesized import only if additional packages are added. Keep the file gofmt-clean.

- [ ] **Step 2: Add repository contract methods**

In `backend/internal/service/ops_port.go`, add this method to `OpsRepository` after `GetLatestUpstreamMultiplierSamples`:

```go
	GetGroupRateRecommendationSourceData(ctx context.Context, filter *OpsGroupRateRecommendationFilter) (*OpsGroupRateRecommendationSourceData, error)
```

- [ ] **Step 3: Extend test repo mock**

In `backend/internal/service/ops_repo_mock_test.go`, add a function field:

```go
	GetGroupRateRecommendationSourceDataFn func(ctx context.Context, filter *OpsGroupRateRecommendationFilter) (*OpsGroupRateRecommendationSourceData, error)
```

Add the method implementation near the other OpsRepository mock methods:

```go
func (m *opsRepoMock) GetGroupRateRecommendationSourceData(ctx context.Context, filter *OpsGroupRateRecommendationFilter) (*OpsGroupRateRecommendationSourceData, error) {
	if m.GetGroupRateRecommendationSourceDataFn != nil {
		return m.GetGroupRateRecommendationSourceDataFn(ctx, filter)
	}
	return &OpsGroupRateRecommendationSourceData{
		Packages: []*OpsGroupRateRecommendationPackageBasis{},
		Groups:   []*OpsGroupRateRecommendationSourceGroup{},
		Usage:    map[int64]map[int64]OpsGroupRateRecommendationUsageShare{},
		Samples:  map[int64]*OpsUpstreamMultiplierSample{},
	}, nil
}
```

- [ ] **Step 4: Run backend compile check for service package**

Run:

```bash
go test ./internal/service
```

Expected: it may fail because repository implementation is not yet added. The expected failure should mention that `*opsRepository` does not implement `service.OpsRepository` because `GetGroupRateRecommendationSourceData` is missing. If it fails for syntax errors, fix them before continuing.

---

### Task 2: Implement repository source-data queries

**Files:**
- Create: `backend/internal/repository/ops_repo_group_rate_recommendations.go`

- [ ] **Step 1: Create repository file**

Create `backend/internal/repository/ops_repo_group_rate_recommendations.go` with this implementation:

```go
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *opsRepository) GetGroupRateRecommendationSourceData(ctx context.Context, filter *service.OpsGroupRateRecommendationFilter) (*service.OpsGroupRateRecommendationSourceData, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}
	if filter == nil {
		return nil, fmt.Errorf("nil filter")
	}
	packages, err := r.queryGroupRateRecommendationPackages(ctx, filter.PackageScope)
	if err != nil {
		return nil, err
	}
	groups, accountIDs, err := r.queryGroupRateRecommendationGroups(ctx, filter)
	if err != nil {
		return nil, err
	}
	usage, err := r.queryGroupRateRecommendationUsage(ctx, time.Now().UTC().AddDate(0, 0, -filter.UsageDays), time.Now().UTC())
	if err != nil {
		return nil, err
	}
	samples, err := r.queryLatestGroupRateRecommendationSamples(ctx, filter.Model, accountIDs)
	if err != nil {
		return nil, err
	}
	return &service.OpsGroupRateRecommendationSourceData{
		Packages: packages,
		Groups:   groups,
		Usage:    usage,
		Samples:  samples,
	}, nil
}

func (r *opsRepository) queryGroupRateRecommendationPackages(ctx context.Context, packageScope string) ([]*service.OpsGroupRateRecommendationPackageBasis, error) {
	const q = `
SELECT id, name, price, credit_amount, package_scope,
       CASE WHEN credit_amount > 0 THEN price / credit_amount ELSE 0 END AS revenue_per_credit
FROM balance_packages
WHERE for_sale = TRUE
  AND ($1 = '' OR package_scope = $1)
  AND price > 0
  AND credit_amount > 0
ORDER BY revenue_per_credit ASC, sort_order ASC, id ASC`
	rows, err := r.db.QueryContext(ctx, q, strings.TrimSpace(packageScope))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := []*service.OpsGroupRateRecommendationPackageBasis{}
	for rows.Next() {
		item := &service.OpsGroupRateRecommendationPackageBasis{}
		if err := rows.Scan(&item.PackageID, &item.Name, &item.Price, &item.CreditAmount, &item.PackageScope, &item.RevenuePerCredit); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *opsRepository) queryGroupRateRecommendationGroups(ctx context.Context, filter *service.OpsGroupRateRecommendationFilter) ([]*service.OpsGroupRateRecommendationSourceGroup, []int64, error) {
	const q = `
SELECT
  g.id,
  g.name,
  COALESCE(g.rate_multiplier, 0),
  COALESCE(g.package_scope, ''),
  COALESCE(g.allow_image_generation, FALSE),
  a.id,
  a.name,
  a.platform,
  a.type,
  a.status,
  COALESCE(a.schedulable, FALSE),
  COALESCE(a.priority, 1),
  COALESCE(ag.priority, 1),
  COALESCE(a.extra->>'base_url', ''),
  CASE
    WHEN a.credentials ? 'api_key' THEN left(a.credentials->>'api_key', 7)
    WHEN a.credentials ? 'access_token' THEN left(a.credentials->>'access_token', 7)
    ELSE ''
  END
FROM groups g
JOIN account_groups ag ON ag.group_id = g.id
JOIN accounts a ON a.id = ag.account_id
WHERE g.deleted_at IS NULL
  AND a.deleted_at IS NULL
  AND g.platform = 'openai'
  AND g.status = 'active'
  AND ($1 = '' OR COALESCE(g.package_scope, '') = $1)
ORDER BY g.sort_order ASC, g.id ASC, ag.priority ASC, a.priority ASC, a.id ASC`
	rows, err := r.db.QueryContext(ctx, q, strings.TrimSpace(filter.PackageScope))
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	groupsByID := map[int64]*service.OpsGroupRateRecommendationSourceGroup{}
	ordered := []*service.OpsGroupRateRecommendationSourceGroup{}
	accountIDSet := map[int64]struct{}{}
	for rows.Next() {
		var groupID int64
		var account service.OpsGroupRateRecommendationSourceAccount
		var groupName, packageScope string
		var rateMultiplier float64
		var allowImage bool
		if err := rows.Scan(
			&groupID,
			&groupName,
			&rateMultiplier,
			&packageScope,
			&allowImage,
			&account.AccountID,
			&account.AccountName,
			&account.Platform,
			&account.Type,
			&account.Status,
			&account.Schedulable,
			&account.CurrentPriority,
			&account.BindingPriority,
			&account.BaseURL,
			&account.KeyPrefix,
		); err != nil {
			return nil, nil, err
		}
		group := groupsByID[groupID]
		if group == nil {
			group = &service.OpsGroupRateRecommendationSourceGroup{
				GroupID:              groupID,
				GroupName:            groupName,
				RateMultiplier:       rateMultiplier,
				PackageScope:         packageScope,
				AllowImageGeneration: allowImage,
				Accounts:             []*service.OpsGroupRateRecommendationSourceAccount{},
			}
			groupsByID[groupID] = group
			ordered = append(ordered, group)
		}
		acc := account
		group.Accounts = append(group.Accounts, &acc)
		accountIDSet[account.AccountID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	ids := make([]int64, 0, len(accountIDSet))
	for id := range accountIDSet {
		ids = append(ids, id)
	}
	return ordered, ids, nil
}

func (r *opsRepository) queryGroupRateRecommendationUsage(ctx context.Context, start, end time.Time) (map[int64]map[int64]service.OpsGroupRateRecommendationUsageShare, error) {
	const q = `
WITH base AS (
  SELECT group_id,
         account_id,
         COUNT(*) AS request_count,
         COALESCE(SUM(total_cost), 0) AS standard_cost
  FROM usage_logs
  WHERE created_at >= $1 AND created_at < $2
    AND group_id IS NOT NULL
    AND account_id IS NOT NULL
  GROUP BY group_id, account_id
), with_totals AS (
  SELECT *,
         SUM(request_count) OVER (PARTITION BY group_id) AS group_request_count,
         SUM(standard_cost) OVER (PARTITION BY group_id) AS group_standard_cost
  FROM base
)
SELECT group_id, account_id, request_count,
       CASE WHEN group_request_count > 0 THEN request_count::float8 / group_request_count::float8 ELSE 0 END AS request_share,
       standard_cost,
       CASE WHEN group_standard_cost > 0 THEN standard_cost::float8 / group_standard_cost::float8 ELSE 0 END AS standard_cost_share
FROM with_totals`
	rows, err := r.db.QueryContext(ctx, q, start, end)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := map[int64]map[int64]service.OpsGroupRateRecommendationUsageShare{}
	for rows.Next() {
		var groupID, accountID int64
		var item service.OpsGroupRateRecommendationUsageShare
		if err := rows.Scan(&groupID, &accountID, &item.RequestCount, &item.RequestShare, &item.StandardCost, &item.StandardCostShare); err != nil {
			return nil, err
		}
		if out[groupID] == nil {
			out[groupID] = map[int64]service.OpsGroupRateRecommendationUsageShare{}
		}
		out[groupID][accountID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *opsRepository) queryLatestGroupRateRecommendationSamples(ctx context.Context, model string, accountIDs []int64) (map[int64]*service.OpsUpstreamMultiplierSample, error) {
	if len(accountIDs) == 0 {
		return map[int64]*service.OpsUpstreamMultiplierSample{}, nil
	}
	return r.GetLatestUpstreamMultiplierSamples(ctx, model, accountIDs)
}
```

- [ ] **Step 2: Run repository compile check**

Run:

```bash
go test ./internal/repository ./internal/service
```

Expected: compilation succeeds or test failures point to missing imports/types. Fix compile errors before continuing.

---

### Task 3: Implement service recommendation math

**Files:**
- Create: `backend/internal/service/ops_group_rate_recommendations.go`
- Create: `backend/internal/service/ops_group_rate_recommendations_test.go`

- [ ] **Step 1: Write service tests first**

Create `backend/internal/service/ops_group_rate_recommendations_test.go`:

```go
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
				Groups: []*OpsGroupRateRecommendationSourceGroup{{GroupID: 2, GroupName: "gpt pro", RateMultiplier: 1, PackageScope: "codex", Accounts: []*OpsGroupRateRecommendationSourceAccount{{AccountID: 1, AccountName: "xixi", Status: StatusActive, Schedulable: true, CurrentPriority: 1}}}},
				Usage: map[int64]map[int64]OpsGroupRateRecommendationUsageShare{},
				Samples: map[int64]*OpsUpstreamMultiplierSample{},
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
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/service -run 'TestGetGroupRateRecommendations' -count=1
```

Expected: FAIL because `GetGroupRateRecommendations` is not implemented.

- [ ] **Step 3: Implement service file**

Create `backend/internal/service/ops_group_rate_recommendations.go`:

```go
package service

import (
	"context"
	"math"
	"sort"
	"strings"
)

const (
	opsGroupRateDefaultModel        = "gpt-5.4"
	opsGroupRateDefaultPackageScope = "codex"
	opsGroupRateDefaultProfitMargin = 0.20
	opsGroupRateDefaultSafetyFactor = 1.20
	opsGroupRateDefaultUsageDays    = 7
	opsGroupRateMaxUsageDays        = 30
)

func (s *OpsService) GetGroupRateRecommendations(ctx context.Context, filter *OpsGroupRateRecommendationFilter) (*OpsGroupRateRecommendationResponse, error) {
	normalized := normalizeOpsGroupRateRecommendationFilter(filter)
	if s == nil || s.opsRepo == nil {
		return &OpsGroupRateRecommendationResponse{Params: normalized, Groups: []*OpsGroupRateRecommendationGroup{}}, nil
	}
	source, err := s.opsRepo.GetGroupRateRecommendationSourceData(ctx, &normalized)
	if err != nil {
		return nil, err
	}
	basis := selectGroupRatePackageBasis(source.Packages, normalized.PackageScope)
	groups := buildGroupRateRecommendations(source, basis, normalized)
	return &OpsGroupRateRecommendationResponse{Params: normalized, PackageBasis: basis, Groups: groups}, nil
}

func normalizeOpsGroupRateRecommendationFilter(filter *OpsGroupRateRecommendationFilter) OpsGroupRateRecommendationFilter {
	out := OpsGroupRateRecommendationFilter{}
	if filter != nil {
		out = *filter
	}
	out.Model = normalizeOpsUpstreamMultiplierModel(out.Model)
	if strings.TrimSpace(out.Model) == "" {
		out.Model = opsGroupRateDefaultModel
	}
	out.PackageScope = strings.TrimSpace(out.PackageScope)
	if out.PackageScope == "" {
		out.PackageScope = opsGroupRateDefaultPackageScope
	}
	if out.ProfitMargin <= 0 || out.ProfitMargin >= 0.95 {
		out.ProfitMargin = opsGroupRateDefaultProfitMargin
	}
	if out.SafetyFactor <= 0 {
		out.SafetyFactor = opsGroupRateDefaultSafetyFactor
	}
	if out.UsageDays <= 0 {
		out.UsageDays = opsGroupRateDefaultUsageDays
	}
	if out.UsageDays > opsGroupRateMaxUsageDays {
		out.UsageDays = opsGroupRateMaxUsageDays
	}
	return out
}

func selectGroupRatePackageBasis(packages []*OpsGroupRateRecommendationPackageBasis, packageScope string) *OpsGroupRateRecommendationPackageBasis {
	var best *OpsGroupRateRecommendationPackageBasis
	for _, pkg := range packages {
		if pkg == nil || pkg.Price <= 0 || pkg.CreditAmount <= 0 || pkg.RevenuePerCredit <= 0 {
			continue
		}
		if packageScope != "" && pkg.PackageScope != packageScope {
			continue
		}
		if best == nil || pkg.RevenuePerCredit < best.RevenuePerCredit || (pkg.RevenuePerCredit == best.RevenuePerCredit && pkg.PackageID < best.PackageID) {
			copyPkg := *pkg
			best = &copyPkg
		}
	}
	return best
}

func buildGroupRateRecommendations(source *OpsGroupRateRecommendationSourceData, basis *OpsGroupRateRecommendationPackageBasis, filter OpsGroupRateRecommendationFilter) []*OpsGroupRateRecommendationGroup {
	if source == nil {
		return []*OpsGroupRateRecommendationGroup{}
	}
	out := make([]*OpsGroupRateRecommendationGroup, 0, len(source.Groups))
	for _, group := range source.Groups {
		if group == nil || group.AllowImageGeneration {
			continue
		}
		item := buildOneGroupRateRecommendation(group, source, basis, filter)
		out = append(out, item)
	}
	return out
}

func buildOneGroupRateRecommendation(group *OpsGroupRateRecommendationSourceGroup, source *OpsGroupRateRecommendationSourceData, basis *OpsGroupRateRecommendationPackageBasis, filter OpsGroupRateRecommendationFilter) *OpsGroupRateRecommendationGroup {
	item := &OpsGroupRateRecommendationGroup{
		GroupID:                group.GroupID,
		GroupName:              group.GroupName,
		CurrentGroupMultiplier: group.RateMultiplier,
		PackageScope:           group.PackageScope,
		Status:                 OpsGroupRateRecommendationStatusInsufficient,
		Accounts:               []*OpsGroupRateRecommendationAccount{},
	}
	if basis == nil || basis.RevenuePerCredit <= 0 {
		item.Notes = append(item.Notes, "缺少可用套餐口径")
	}

	usageByAccount := source.Usage[group.GroupID]
	participants := []*OpsGroupRateRecommendationAccount{}
	for _, acc := range group.Accounts {
		if acc == nil {
			continue
		}
		usage := usageByAccount[acc.AccountID]
		sample := source.Samples[acc.AccountID]
		account := &OpsGroupRateRecommendationAccount{
			AccountID:        acc.AccountID,
			AccountName:      acc.AccountName,
			BaseURL:          acc.BaseURL,
			KeyPrefix:        acc.KeyPrefix,
			Schedulable:      acc.Schedulable,
			Status:           acc.Status,
			CurrentPriority:  acc.CurrentPriority,
			BindingPriority:  acc.BindingPriority,
			RequestCount:     usage.RequestCount,
			RequestShare:     roundFloat(usage.RequestShare, 6),
			StandardCost:     roundFloat(usage.StandardCost, 8),
			StandardCostShare: roundFloat(usage.StandardCostShare, 6),
		}
		if sample != nil {
			account.MultiplierStatus = sample.Status
			account.UpstreamMultiplier = sample.Multiplier
			account.MultiplierMeasuredAt = &sample.MeasuredAt
		}
		if isGroupRateRecommendationParticipant(acc, account, filter) {
			account.ParticipatesInAdvice = true
			participants = append(participants, account)
			if acc.Schedulable {
				item.SchedulableAccountCount++
			}
		} else {
			account.Note = groupRateAccountSkipNote(acc, account)
		}
		item.Accounts = append(item.Accounts, account)
	}

	assignRecommendedWeights(participants)
	actual := blendedMultiplierFromUsage(participants)
	recommended := blendedMultiplierFromRecommendedWeights(participants)
	worst := worstCaseMultiplier(participants)
	item.ActualBlendedMultiplier = actual
	item.RecommendedBlendedMultiplier = recommended
	item.WorstCaseMultiplier = worst
	if basis != nil && recommended != nil {
		min := recommendationRequiredGroupMultiplier(*recommended, basis.RevenuePerCredit, filter.ProfitMargin, filter.SafetyFactor)
		item.MinimumGroupMultiplier = &min
	}
	if basis != nil && worst != nil {
		safe := recommendationRequiredGroupMultiplier(*worst, basis.RevenuePerCredit, filter.ProfitMargin, filter.SafetyFactor)
		item.SafeGroupMultiplier = &safe
	}
	item.Status = classifyGroupRateRecommendation(item)
	for _, acc := range item.Accounts {
		if acc.ParticipatesInAdvice && acc.Note == "" {
			acc.Note = groupRateParticipantNote(acc)
		}
	}
	if len(participants) == 0 {
		item.Notes = append(item.Notes, "没有可参与建议的上游账号")
	}
	return item
}

func isGroupRateRecommendationParticipant(source *OpsGroupRateRecommendationSourceAccount, account *OpsGroupRateRecommendationAccount, filter OpsGroupRateRecommendationFilter) bool {
	if source == nil || account == nil {
		return false
	}
	if source.Status != StatusActive {
		return false
	}
	if !filter.IncludeUnschedulable && !source.Schedulable {
		return false
	}
	if !filter.IncludeSelfHosted && strings.Contains(strings.ToLower(source.AccountName), "自建") {
		return false
	}
	if account.UpstreamMultiplier == nil || *account.UpstreamMultiplier <= 0 || account.MultiplierStatus != OpsUpstreamMultiplierStatusSuccess {
		return false
	}
	return true
}

func groupRateAccountSkipNote(source *OpsGroupRateRecommendationSourceAccount, account *OpsGroupRateRecommendationAccount) string {
	if source == nil || account == nil {
		return "数据异常"
	}
	if source.Status != StatusActive {
		return "账号未启用，不参与建议"
	}
	if !source.Schedulable {
		return "不可调度，不参与建议"
	}
	if strings.Contains(strings.ToLower(source.AccountName), "自建") {
		return "自建账号默认不参与建议"
	}
	if account.UpstreamMultiplier == nil || account.MultiplierStatus != OpsUpstreamMultiplierStatusSuccess {
		return "缺少成功倍率样本，请先检测"
	}
	return "不参与建议"
}

func assignRecommendedWeights(accounts []*OpsGroupRateRecommendationAccount) {
	if len(accounts) == 0 {
		return
	}
	if len(accounts) == 1 {
		accounts[0].RecommendedWeight = 1
		accounts[0].RecommendedPriority = 1
		return
	}
	inv := make([]float64, len(accounts))
	total := 0.0
	for i, account := range accounts {
		m := 0.0
		if account.UpstreamMultiplier != nil {
			m = *account.UpstreamMultiplier
		}
		if m <= 0 {
			continue
		}
		inv[i] = 1 / m
		total += inv[i]
	}
	if total <= 0 {
		return
	}
	for i, account := range accounts {
		account.RecommendedWeight = inv[i] / total
	}
	capRecommendedWeights(accounts, 0.5)
	for _, account := range accounts {
		account.RecommendedWeight = roundFloat(account.RecommendedWeight, 4)
		account.RecommendedPriority = recommendedPriorityFromWeight(account.RecommendedWeight)
	}
}

func capRecommendedWeights(accounts []*OpsGroupRateRecommendationAccount, capValue float64) {
	for iter := 0; iter < 5; iter++ {
		over := 0.0
		underTotal := 0.0
		for _, account := range accounts {
			if account.RecommendedWeight > capValue {
				over += account.RecommendedWeight - capValue
				account.RecommendedWeight = capValue
			} else {
				underTotal += account.RecommendedWeight
			}
		}
		if over <= 0 || underTotal <= 0 {
			break
		}
		for _, account := range accounts {
			if account.RecommendedWeight < capValue {
				account.RecommendedWeight += over * (account.RecommendedWeight / underTotal)
			}
		}
	}
	total := 0.0
	for _, account := range accounts {
		total += account.RecommendedWeight
	}
	if total > 0 {
		for _, account := range accounts {
			account.RecommendedWeight /= total
		}
	}
}

func recommendedPriorityFromWeight(weight float64) int {
	switch {
	case weight >= 0.40:
		return 1
	case weight >= 0.25:
		return 2
	case weight >= 0.15:
		return 4
	case weight >= 0.05:
		return 6
	default:
		return 8
	}
}

func blendedMultiplierFromUsage(accounts []*OpsGroupRateRecommendationAccount) *float64 {
	totalUsage := 0.0
	value := 0.0
	for _, account := range accounts {
		if account.UpstreamMultiplier == nil || account.StandardCostShare <= 0 {
			continue
		}
		totalUsage += account.StandardCostShare
		value += *account.UpstreamMultiplier * account.StandardCostShare
	}
	if totalUsage <= 0 {
		return blendedMultiplierFromRecommendedWeights(accounts)
	}
	result := roundFloat(value/totalUsage, 6)
	return &result
}

func blendedMultiplierFromRecommendedWeights(accounts []*OpsGroupRateRecommendationAccount) *float64 {
	value := 0.0
	total := 0.0
	for _, account := range accounts {
		if account.UpstreamMultiplier == nil || account.RecommendedWeight <= 0 {
			continue
		}
		value += *account.UpstreamMultiplier * account.RecommendedWeight
		total += account.RecommendedWeight
	}
	if total <= 0 {
		return nil
	}
	result := roundFloat(value/total, 6)
	return &result
}

func worstCaseMultiplier(accounts []*OpsGroupRateRecommendationAccount) *float64 {
	maxValue := 0.0
	for _, account := range accounts {
		if account.UpstreamMultiplier != nil && *account.UpstreamMultiplier > maxValue {
			maxValue = *account.UpstreamMultiplier
		}
	}
	if maxValue <= 0 {
		return nil
	}
	result := roundFloat(maxValue, 6)
	return &result
}

func recommendationRequiredGroupMultiplier(upstreamMultiplier, revenuePerCredit, profitMargin, safetyFactor float64) float64 {
	if upstreamMultiplier <= 0 || revenuePerCredit <= 0 || profitMargin >= 1 {
		return 0
	}
	return roundFloat(upstreamMultiplier*safetyFactor/(1-profitMargin)/revenuePerCredit, 6)
}

func classifyGroupRateRecommendation(group *OpsGroupRateRecommendationGroup) string {
	if group == nil || group.MinimumGroupMultiplier == nil {
		return OpsGroupRateRecommendationStatusInsufficient
	}
	if group.SafeGroupMultiplier != nil && group.CurrentGroupMultiplier >= *group.SafeGroupMultiplier {
		return OpsGroupRateRecommendationStatusSafe
	}
	if group.CurrentGroupMultiplier >= *group.MinimumGroupMultiplier {
		return OpsGroupRateRecommendationStatusBasicSafe
	}
	return OpsGroupRateRecommendationStatusLow
}

func groupRateParticipantNote(account *OpsGroupRateRecommendationAccount) string {
	if account == nil || account.UpstreamMultiplier == nil {
		return ""
	}
	if account.RecommendedWeight >= 0.35 {
		return "成本较低，建议主力"
	}
	if account.RecommendedWeight >= 0.15 {
		return "建议保留补充权重"
	}
	return "成本较高，建议热备"
}

func sortGroupRateAccounts(accounts []*OpsGroupRateRecommendationAccount) {
	sort.SliceStable(accounts, func(i, j int) bool {
		if accounts[i].ParticipatesInAdvice != accounts[j].ParticipatesInAdvice {
			return accounts[i].ParticipatesInAdvice
		}
		if accounts[i].RecommendedWeight != accounts[j].RecommendedWeight {
			return accounts[i].RecommendedWeight > accounts[j].RecommendedWeight
		}
		return accounts[i].AccountID < accounts[j].AccountID
	})
}

func roundFloat(value float64, precision int) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	pow := math.Pow10(precision)
	return math.Round(value*pow) / pow
}
```

If `roundFloat` already exists in `ops_upstream_multiplier.go`, do not duplicate it. Remove the `roundFloat` definition and unused `math` import from the new file.

Add this line before `return item` in `buildOneGroupRateRecommendation`:

```go
	sortGroupRateAccounts(item.Accounts)
```

- [ ] **Step 4: Run service tests**

Run:

```bash
go test ./internal/service -run 'TestGetGroupRateRecommendations' -count=1
```

Expected: PASS.

- [ ] **Step 5: Run full service package tests**

Run:

```bash
go test ./internal/service
```

Expected: PASS.

---

### Task 4: Add admin handler and route

**Files:**
- Modify: `backend/internal/handler/admin/ops_dashboard_handler.go`
- Modify: `backend/internal/server/routes/admin.go`

- [ ] **Step 1: Add handler method**

In `backend/internal/handler/admin/ops_dashboard_handler.go`, add this method after `GetUpstreamMultiplierSamples`:

```go
// GetGroupRateRecommendations returns read-only group multiplier and upstream weight suggestions.
// GET /api/v1/admin/ops/group-rate-recommendations
func (h *OpsHandler) GetGroupRateRecommendations(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	filter := &service.OpsGroupRateRecommendationFilter{
		Model:        strings.TrimSpace(c.Query("model")),
		PackageScope: strings.TrimSpace(c.Query("package_scope")),
	}
	if v := strings.TrimSpace(c.Query("profit_margin")); v != "" {
		value, err := strconv.ParseFloat(v, 64)
		if err != nil || value <= 0 || value >= 0.95 {
			response.BadRequest(c, "Invalid profit_margin")
			return
		}
		filter.ProfitMargin = value
	}
	if v := strings.TrimSpace(c.Query("safety_factor")); v != "" {
		value, err := strconv.ParseFloat(v, 64)
		if err != nil || value <= 0 {
			response.BadRequest(c, "Invalid safety_factor")
			return
		}
		filter.SafetyFactor = value
	}
	if v := strings.TrimSpace(c.Query("usage_days")); v != "" {
		value, err := strconv.Atoi(v)
		if err != nil || value <= 0 {
			response.BadRequest(c, "Invalid usage_days")
			return
		}
		filter.UsageDays = value
	}
	if v := strings.TrimSpace(c.Query("include_unschedulable")); v != "" {
		filter.IncludeUnschedulable = v == "true" || v == "1"
	}
	if v := strings.TrimSpace(c.Query("include_self_hosted")); v != "" {
		filter.IncludeSelfHosted = v == "true" || v == "1"
	}
	data, err := h.opsService.GetGroupRateRecommendations(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}
```

Imports `net/http`, `strconv`, and `strings` are already used in this file. If one is not currently used after edits, gofmt/go test will reveal it.

- [ ] **Step 2: Register route**

In `backend/internal/server/routes/admin.go`, after upstream multiplier routes, add:

```go
		ops.GET("/group-rate-recommendations", h.Admin.Ops.GetGroupRateRecommendations)
```

- [ ] **Step 3: Run backend tests**

Run:

```bash
go test ./internal/handler/admin ./internal/server ./internal/service ./internal/repository
```

Expected: PASS.

---

### Task 5: Add frontend API types and method

**Files:**
- Modify: `frontend/src/api/admin/ops.ts`

- [ ] **Step 1: Add TypeScript types**

Near upstream multiplier types in `frontend/src/api/admin/ops.ts`, add:

```ts
export type OpsGroupRateRecommendationStatus = 'safe' | 'basic_safe' | 'low' | 'insufficient_data'

export interface OpsGroupRateRecommendationParams {
  model?: string
  package_scope?: string
  profit_margin?: number
  safety_factor?: number
  usage_days?: number
  include_unschedulable?: boolean
  include_self_hosted?: boolean
}

export interface OpsGroupRateRecommendationPackageBasis {
  package_id: number
  name: string
  price: number
  credit_amount: number
  package_scope: string
  revenue_per_credit: number
}

export interface OpsGroupRateRecommendationAccount {
  account_id: number
  account_name: string
  base_url: string
  key_prefix: string
  schedulable: boolean
  status: string
  current_priority: number
  binding_priority: number
  upstream_multiplier?: number | null
  multiplier_status?: string
  multiplier_measured_at?: string | null
  request_count: number
  request_share: number
  standard_cost: number
  standard_cost_share: number
  recommended_weight: number
  recommended_priority: number
  participates_in_advice: boolean
  note?: string
}

export interface OpsGroupRateRecommendationGroup {
  group_id: number
  group_name: string
  current_group_multiplier: number
  package_scope: string
  schedulable_account_count: number
  actual_blended_multiplier?: number | null
  recommended_blended_multiplier?: number | null
  worst_case_multiplier?: number | null
  minimum_group_multiplier?: number | null
  safe_group_multiplier?: number | null
  status: OpsGroupRateRecommendationStatus
  notes?: string[]
  accounts: OpsGroupRateRecommendationAccount[]
}

export interface OpsGroupRateRecommendationsResponse {
  params: OpsGroupRateRecommendationParams
  package_basis?: OpsGroupRateRecommendationPackageBasis | null
  groups: OpsGroupRateRecommendationGroup[]
}
```

- [ ] **Step 2: Add API function**

Near `getUpstreamMultiplierSamples`, add:

```ts
async function getGroupRateRecommendations(params: OpsGroupRateRecommendationParams): Promise<OpsGroupRateRecommendationsResponse> {
  const { data } = await apiClient.get<OpsGroupRateRecommendationsResponse>('/admin/ops/group-rate-recommendations', { params })
  return data
}
```

- [ ] **Step 3: Export in `opsAPI` object**

Find the `opsAPI` export near the bottom of the file and add:

```ts
  getGroupRateRecommendations,
```

- [ ] **Step 4: Run frontend type check**

Run:

```bash
pnpm --dir frontend exec vue-tsc -b
```

Expected: PASS or unrelated existing type errors. Fix errors caused by the new types.

---

### Task 6: Build `GroupRateRecommendationsPanel.vue`

**Files:**
- Create: `frontend/src/views/admin/ops/components/GroupRateRecommendationsPanel.vue`
- Create: `frontend/src/views/admin/ops/components/__tests__/GroupRateRecommendationsPanel.spec.ts`

- [ ] **Step 1: Write component test first**

Create `frontend/src/views/admin/ops/components/__tests__/GroupRateRecommendationsPanel.spec.ts`:

```ts
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import GroupRateRecommendationsPanel from '../GroupRateRecommendationsPanel.vue'
import type { OpsGroupRateRecommendationsResponse } from '@/api/admin/ops'

const sample: OpsGroupRateRecommendationsResponse = {
  params: { model: 'gpt-5.4', package_scope: 'codex', profit_margin: 0.2, safety_factor: 1.2, usage_days: 7 },
  package_basis: { package_id: 2, name: '专属包-进阶级', price: 100, credit_amount: 400, package_scope: 'codex', revenue_per_credit: 0.25 },
  groups: [
    {
      group_id: 8,
      group_name: 'gpt pro 高价',
      current_group_multiplier: 1.3,
      package_scope: 'codex',
      schedulable_account_count: 2,
      actual_blended_multiplier: 0.148,
      recommended_blended_multiplier: 0.1485,
      worst_case_multiplier: 0.18,
      minimum_group_multiplier: 0.891,
      safe_group_multiplier: 1.08,
      status: 'safe',
      accounts: [
        {
          account_id: 9,
          account_name: '天才程序员',
          base_url: 'https://api.dzzzz.cf',
          key_prefix: 'sk-f642',
          schedulable: true,
          status: 'active',
          current_priority: 1,
          binding_priority: 1,
          upstream_multiplier: 0.135,
          multiplier_status: 'success',
          request_count: 70,
          request_share: 0.7,
          standard_cost: 70,
          standard_cost_share: 0.7,
          recommended_weight: 0.7,
          recommended_priority: 1,
          participates_in_advice: true,
          note: '成本较低，建议主力',
        },
      ],
    },
  ],
}

describe('GroupRateRecommendationsPanel', () => {
  it('renders package basis, group recommendation, and account weight advice', () => {
    const wrapper = mount(GroupRateRecommendationsPanel, {
      props: {
        model: 'gpt-5.4',
        data: sample,
        loading: false,
        profitMargin: 0.2,
        safetyFactor: 1.2,
        usageDays: 7,
      },
    })

    expect(wrapper.text()).toContain('分组倍率与权重建议')
    expect(wrapper.text()).toContain('专属包-进阶级')
    expect(wrapper.text()).toContain('gpt pro 高价')
    expect(wrapper.text()).toContain('1.3x')
    expect(wrapper.text()).toContain('1.08x')
    expect(wrapper.text()).toContain('天才程序员')
    expect(wrapper.text()).toContain('70%')
  })

  it('emits refresh when parameters change', async () => {
    const wrapper = mount(GroupRateRecommendationsPanel, {
      props: {
        model: 'gpt-5.4',
        data: sample,
        loading: false,
        profitMargin: 0.2,
        safetyFactor: 1.2,
        usageDays: 7,
      },
    })

    await wrapper.get('[data-testid="profit-margin-input"]').setValue('0.25')
    expect(wrapper.emitted('update:profitMargin')?.[0]).toEqual([0.25])
    expect(wrapper.emitted('refresh')).toBeTruthy()
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
pnpm --dir frontend run test:run -- src/views/admin/ops/components/__tests__/GroupRateRecommendationsPanel.spec.ts
```

Expected: FAIL because component does not exist.

- [ ] **Step 3: Create component**

Create `frontend/src/views/admin/ops/components/GroupRateRecommendationsPanel.vue` with:

```vue
<template>
  <section class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-dark-800">
    <div class="flex flex-col gap-4 border-b border-gray-100 px-4 py-4 dark:border-gray-700 lg:flex-row lg:items-start lg:justify-between">
      <div>
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">分组倍率与权重建议</h3>
        <p class="mt-1 max-w-3xl text-xs leading-5 text-gray-500 dark:text-gray-400">
          按余额包折扣、上游真实倍率和最近使用占比估算分组成本；这里只给建议，不会自动修改分组或账号配置。
        </p>
      </div>
      <div class="grid gap-2 sm:grid-cols-3 lg:min-w-[520px]">
        <label class="text-xs font-medium text-gray-600 dark:text-gray-300">
          目标利润率
          <input
            data-testid="profit-margin-input"
            :value="profitMargin"
            type="number"
            step="0.01"
            min="0.01"
            max="0.9"
            class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-gray-700 dark:bg-dark-900 dark:text-white"
            @change="emitNumber('update:profitMargin', ($event.target as HTMLInputElement).value)"
          />
        </label>
        <label class="text-xs font-medium text-gray-600 dark:text-gray-300">
          安全系数
          <input
            :value="safetyFactor"
            type="number"
            step="0.1"
            min="1"
            class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-gray-700 dark:bg-dark-900 dark:text-white"
            @change="emitNumber('update:safetyFactor', ($event.target as HTMLInputElement).value)"
          />
        </label>
        <label class="text-xs font-medium text-gray-600 dark:text-gray-300">
          使用天数
          <input
            :value="usageDays"
            type="number"
            step="1"
            min="1"
            max="30"
            class="mt-1 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-gray-700 dark:bg-dark-900 dark:text-white"
            @change="emitInteger('update:usageDays', ($event.target as HTMLInputElement).value)"
          />
        </label>
      </div>
    </div>

    <div v-if="loading" class="flex h-40 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      加载分组建议中…
    </div>

    <div v-else class="space-y-4 p-4">
      <div class="rounded-xl border border-blue-100 bg-blue-50 px-4 py-3 text-xs text-blue-800 dark:border-blue-500/20 dark:bg-blue-500/10 dark:text-blue-200">
        <template v-if="data?.package_basis">
          当前套餐口径：{{ data.package_basis.name }}，{{ formatMoney(data.package_basis.price) }} 元买 {{ trimNumber(data.package_basis.credit_amount) }} 额度，单额度收入 {{ trimNumber(data.package_basis.revenue_per_credit) }}。
        </template>
        <template v-else>
          没有找到可用套餐口径，建议结果会显示为数据不足。
        </template>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-100 text-sm dark:divide-gray-700">
          <thead class="bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:bg-dark-700/60 dark:text-gray-400">
            <tr>
              <th class="px-3 py-3 text-left">分组</th>
              <th class="px-3 py-3 text-right">当前倍率</th>
              <th class="px-3 py-3 text-right">综合成本</th>
              <th class="px-3 py-3 text-right">最坏成本</th>
              <th class="px-3 py-3 text-right">建议倍率</th>
              <th class="px-3 py-3 text-left">状态</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
            <tr v-if="groups.length === 0">
              <td colspan="6" class="px-3 py-8 text-center text-sm text-gray-500 dark:text-gray-400">暂无分组建议</td>
            </tr>
            <template v-for="group in groups" :key="group.group_id">
              <tr class="bg-white dark:bg-dark-800">
                <td class="px-3 py-3 font-semibold text-gray-900 dark:text-white">{{ group.group_name }}</td>
                <td class="px-3 py-3 text-right font-semibold text-gray-900 dark:text-white">{{ formatMultiplier(group.current_group_multiplier) }}</td>
                <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-200">{{ formatMultiplier(group.recommended_blended_multiplier ?? group.actual_blended_multiplier) }}</td>
                <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-200">{{ formatMultiplier(group.worst_case_multiplier) }}</td>
                <td class="px-3 py-3 text-right text-gray-700 dark:text-gray-200">
                  <div>最低 {{ formatMultiplier(group.minimum_group_multiplier) }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">稳妥 {{ formatMultiplier(group.safe_group_multiplier) }}</div>
                </td>
                <td class="px-3 py-3">
                  <span class="inline-flex rounded-full px-2 py-1 text-xs font-semibold" :class="statusClass(group.status)">{{ statusText(group.status) }}</span>
                </td>
              </tr>
              <tr>
                <td colspan="6" class="bg-gray-50 px-3 py-3 dark:bg-dark-900/50">
                  <div class="grid gap-2 lg:grid-cols-2 xl:grid-cols-3">
                    <div v-for="account in group.accounts" :key="account.account_id" class="rounded-lg border border-gray-100 bg-white p-3 text-xs dark:border-gray-700 dark:bg-dark-800">
                      <div class="flex items-start justify-between gap-2">
                        <div>
                          <div class="font-semibold text-gray-900 dark:text-white">{{ account.account_name }}</div>
                          <div class="text-gray-500 dark:text-gray-400">{{ hostOf(account.base_url) }}</div>
                        </div>
                        <div class="text-right font-semibold text-gray-900 dark:text-white">{{ formatMultiplier(account.upstream_multiplier) }}</div>
                      </div>
                      <div class="mt-3 grid grid-cols-2 gap-2 text-gray-600 dark:text-gray-300">
                        <div>当前占比 {{ formatPercent(account.standard_cost_share) }}</div>
                        <div>建议权重 {{ formatPercent(account.recommended_weight) }}</div>
                        <div>当前 priority {{ account.current_priority }}</div>
                        <div>建议 priority {{ account.recommended_priority || '—' }}</div>
                      </div>
                      <div class="mt-2 text-gray-500 dark:text-gray-400">{{ account.note || '—' }}</div>
                    </div>
                  </div>
                  <div v-if="group.notes?.length" class="mt-2 text-xs text-amber-700 dark:text-amber-300">{{ group.notes.join('；') }}</div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { OpsGroupRateRecommendationStatus, OpsGroupRateRecommendationsResponse } from '@/api/admin/ops'

const props = defineProps<{
  model: string
  data?: OpsGroupRateRecommendationsResponse | null
  loading?: boolean
  profitMargin: number
  safetyFactor: number
  usageDays: number
}>()

const emit = defineEmits<{
  'update:profitMargin': [value: number]
  'update:safetyFactor': [value: number]
  'update:usageDays': [value: number]
  refresh: []
}>()

const groups = computed(() => props.data?.groups || [])

function emitNumber(event: 'update:profitMargin' | 'update:safetyFactor', raw: string) {
  const value = Number(raw)
  if (!Number.isFinite(value)) return
  emit(event, value)
  emit('refresh')
}

function emitInteger(event: 'update:usageDays', raw: string) {
  const value = Number.parseInt(raw, 10)
  if (!Number.isFinite(value)) return
  emit(event, value)
  emit('refresh')
}

function formatMultiplier(value?: number | null): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return `${trimNumber(value)}x`
}

function formatPercent(value?: number | null): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return `${trimNumber(value * 100)}%`
}

function formatMoney(value?: number | null): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) return '—'
  return trimNumber(value)
}

function trimNumber(value: number): string {
  return value.toFixed(4).replace(/0+$/, '').replace(/\.$/, '')
}

function hostOf(raw: string): string {
  try {
    return new URL(raw).host
  } catch {
    return raw || '—'
  }
}

function statusText(status: OpsGroupRateRecommendationStatus): string {
  switch (status) {
    case 'safe':
      return '安全'
    case 'basic_safe':
      return '基本安全'
    case 'low':
      return '偏低'
    default:
      return '数据不足'
  }
}

function statusClass(status: OpsGroupRateRecommendationStatus): string {
  switch (status) {
    case 'safe':
      return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'basic_safe':
      return 'bg-blue-50 text-blue-700 dark:bg-blue-500/15 dark:text-blue-300'
    case 'low':
      return 'bg-rose-50 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-200'
  }
}
</script>
```

- [ ] **Step 4: Run component test**

Run:

```bash
pnpm --dir frontend run test:run -- src/views/admin/ops/components/__tests__/GroupRateRecommendationsPanel.spec.ts
```

Expected: PASS.

---

### Task 7: Integrate panel into Provider Status page

**Files:**
- Modify: `frontend/src/views/admin/ops/ProviderStatusView.vue`
- Modify: `frontend/src/views/admin/ops/__tests__/ProviderStatusView.spec.ts`

- [ ] **Step 1: Update ProviderStatusView imports and state**

In `ProviderStatusView.vue`, import the new component and types:

```ts
  type OpsGroupRateRecommendationsResponse,
```

Add component import:

```ts
import GroupRateRecommendationsPanel from './components/GroupRateRecommendationsPanel.vue'
```

Add refs:

```ts
const recommendationLoading = ref(false)
const recommendationData = ref<OpsGroupRateRecommendationsResponse | null>(null)
const recommendationProfitMargin = ref(0.2)
const recommendationSafetyFactor = ref(1.2)
const recommendationUsageDays = ref(7)
```

- [ ] **Step 2: Add loader function**

Add after `loadUpstreamMultipliers`:

```ts
async function loadGroupRateRecommendations() {
  recommendationLoading.value = true
  try {
    const model = multiplierModel.value.trim() || 'gpt-5.4'
    recommendationData.value = await opsAPI.getGroupRateRecommendations({
      model,
      package_scope: 'codex',
      profit_margin: recommendationProfitMargin.value,
      safety_factor: recommendationSafetyFactor.value,
      usage_days: recommendationUsageDays.value,
    })
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '加载分组倍率建议失败'))
  } finally {
    recommendationLoading.value = false
  }
}
```

- [ ] **Step 3: Render panel**

In the template, below `UpstreamMultiplierPanel`, add:

```vue
      <GroupRateRecommendationsPanel
        :model="multiplierModel"
        :data="recommendationData"
        :loading="recommendationLoading"
        v-model:profit-margin="recommendationProfitMargin"
        v-model:safety-factor="recommendationSafetyFactor"
        v-model:usage-days="recommendationUsageDays"
        @refresh="loadGroupRateRecommendations"
      />
```

- [ ] **Step 4: Load recommendations on mount and model change**

In `watch(multiplierModel, ...)`, change to:

```ts
watch(multiplierModel, () => {
  void loadUpstreamMultipliers()
  void loadGroupRateRecommendations()
})
```

In `onMounted`, add:

```ts
  void loadGroupRateRecommendations()
```

- [ ] **Step 5: Update ProviderStatusView test mock**

In `frontend/src/views/admin/ops/__tests__/ProviderStatusView.spec.ts`, add:

```ts
const mockGetGroupRateRecommendations = vi.hoisted(() => vi.fn())
```

Add it to `opsAPI` mock:

```ts
    getGroupRateRecommendations: mockGetGroupRateRecommendations,
```

In `beforeEach`, add:

```ts
    mockGetGroupRateRecommendations.mockReset().mockResolvedValue({
      params: { model: 'gpt-5.4', package_scope: 'codex', profit_margin: 0.2, safety_factor: 1.2, usage_days: 7 },
      package_basis: { package_id: 2, name: '专属包-进阶级', price: 100, credit_amount: 400, package_scope: 'codex', revenue_per_credit: 0.25 },
      groups: [
        {
          group_id: 8,
          group_name: 'gpt pro 高价',
          current_group_multiplier: 1.3,
          package_scope: 'codex',
          schedulable_account_count: 2,
          actual_blended_multiplier: 0.148,
          recommended_blended_multiplier: 0.1485,
          worst_case_multiplier: 0.18,
          minimum_group_multiplier: 0.891,
          safe_group_multiplier: 1.08,
          status: 'safe',
          accounts: [],
        },
      ],
    })
```

Add an assertion to the multiplier test:

```ts
    expect(mockGetGroupRateRecommendations).toHaveBeenCalledWith(expect.objectContaining({ model: 'gpt-5.4', profit_margin: 0.2, safety_factor: 1.2, usage_days: 7 }))
    expect(wrapper.text()).toContain('分组倍率与权重建议')
    expect(wrapper.text()).toContain('gpt pro 高价')
```

- [ ] **Step 6: Run frontend tests**

Run:

```bash
pnpm --dir frontend run test:run -- src/views/admin/ops/__tests__/ProviderStatusView.spec.ts src/views/admin/ops/components/__tests__/GroupRateRecommendationsPanel.spec.ts
```

Expected: PASS.

---

### Task 8: Final verification

**Files:**
- No new files unless previous tasks revealed required fixes.

- [ ] **Step 1: Run backend focused tests**

Run:

```bash
go test ./internal/service ./internal/repository ./internal/handler/admin ./internal/server
```

Expected: PASS.

- [ ] **Step 2: Run frontend build/type/test checks**

Run:

```bash
pnpm --dir frontend exec vue-tsc -b
pnpm --dir frontend run test:run -- src/views/admin/ops/__tests__/ProviderStatusView.spec.ts src/views/admin/ops/components/__tests__/GroupRateRecommendationsPanel.spec.ts
```

Expected: PASS.

- [ ] **Step 3: Run full project tests if practical**

Run:

```bash
go test ./...
pnpm --dir frontend run build
```

Expected: PASS. If `pnpm --dir frontend run build` fails due to known Docker context docs path issues, verify local frontend build context includes `docs/legal/*` or document the exact failure.

- [ ] **Step 4: Manual API smoke test**

With the backend running locally or on a test environment, call:

```bash
curl -sS -H "Authorization: Bearer <admin-token>" \
  "http://localhost:8080/api/v1/admin/ops/group-rate-recommendations?model=gpt-5.4&package_scope=codex&profit_margin=0.2&safety_factor=1.2&usage_days=7" | jq '.data.groups[] | {group_name,current_group_multiplier,recommended_blended_multiplier,safe_group_multiplier,status}'
```

Expected: returns OpenAI group recommendations without exposing full API keys.

- [ ] **Step 5: Commit implementation**

Run:

```bash
git add backend/internal/service/ops_dashboard_models.go \
  backend/internal/service/ops_port.go \
  backend/internal/service/ops_repo_mock_test.go \
  backend/internal/service/ops_group_rate_recommendations.go \
  backend/internal/service/ops_group_rate_recommendations_test.go \
  backend/internal/repository/ops_repo_group_rate_recommendations.go \
  backend/internal/handler/admin/ops_dashboard_handler.go \
  backend/internal/server/routes/admin.go \
  frontend/src/api/admin/ops.ts \
  frontend/src/views/admin/ops/ProviderStatusView.vue \
  frontend/src/views/admin/ops/__tests__/ProviderStatusView.spec.ts \
  frontend/src/views/admin/ops/components/GroupRateRecommendationsPanel.vue \
  frontend/src/views/admin/ops/components/__tests__/GroupRateRecommendationsPanel.spec.ts

git commit -m "feat: recommend group rates from upstream costs"
```

Expected: commit succeeds.

---

## Self-Review Checklist

- Spec coverage: The plan covers 20% margin, safety factor 1.2, package basis, multiple upstream accounts per group, actual/recommended/worst costs, suggested account weights, priority advice, admin-only page, and read-only behavior.
- No placeholders: All tasks include explicit file paths, code snippets, commands, and expected results.
- Type consistency: Backend DTO JSON names match frontend TypeScript names. Status values are `safe`, `basic_safe`, `low`, and `insufficient_data` throughout.
- Scope: No task modifies billing behavior or auto-applies recommendations.
