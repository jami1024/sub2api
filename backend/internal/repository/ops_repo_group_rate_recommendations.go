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
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -filter.UsageDays)
	usage, err := r.queryGroupRateRecommendationUsage(ctx, start, end)
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
  CASE
    WHEN a.credentials ? 'base_url' THEN a.credentials->>'base_url'
    WHEN a.extra ? 'base_url' THEN a.extra->>'base_url'
    ELSE ''
  END,
  CASE
    WHEN a.credentials ? 'api_key' THEN left(a.credentials->>'api_key', 8)
    WHEN a.credentials ? 'access_token' THEN left(a.credentials->>'access_token', 8)
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
  SELECT ul.group_id,
         ul.account_id,
         COUNT(*) AS request_count,
         COALESCE(SUM(ul.total_cost), 0) AS standard_cost
  FROM usage_logs ul
  JOIN groups g ON g.id = ul.group_id
  WHERE ul.created_at >= $1 AND ul.created_at < $2
    AND ul.group_id IS NOT NULL
    AND ul.account_id IS NOT NULL
    AND g.platform = 'openai'
  GROUP BY ul.group_id, ul.account_id
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
