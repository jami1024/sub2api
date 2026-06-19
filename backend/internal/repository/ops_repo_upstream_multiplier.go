package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

const opsUpstreamMultiplierSampleColumns = `
id, account_id, account_name_snapshot, platform, base_url_snapshot, key_prefix_snapshot,
model, status, http_status, standard_cost_delta, actual_cost_delta, multiplier,
balance_before, balance_after, error_message, measured_at, created_at`

func (r *opsRepository) InsertUpstreamMultiplierSample(ctx context.Context, input *service.OpsUpstreamMultiplierSample) (*service.OpsUpstreamMultiplierSample, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}
	if input == nil {
		return nil, fmt.Errorf("nil input")
	}
	if input.MeasuredAt.IsZero() {
		input.MeasuredAt = time.Now().UTC()
	}
	const q = `
INSERT INTO ops_upstream_multiplier_samples (
  account_id, account_name_snapshot, platform, base_url_snapshot, key_prefix_snapshot,
  model, status, http_status, standard_cost_delta, actual_cost_delta, multiplier,
  balance_before, balance_after, error_message, measured_at
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
)
RETURNING ` + opsUpstreamMultiplierSampleColumns

	row := r.db.QueryRowContext(
		ctx,
		q,
		input.AccountID,
		input.AccountNameSnapshot,
		input.Platform,
		input.BaseURLSnapshot,
		input.KeyPrefixSnapshot,
		input.Model,
		input.Status,
		input.HTTPStatus,
		input.StandardCostDelta,
		input.ActualCostDelta,
		input.Multiplier,
		input.BalanceBefore,
		input.BalanceAfter,
		input.ErrorMessage,
		input.MeasuredAt.UTC(),
	)
	return scanUpstreamMultiplierSample(row)
}

func (r *opsRepository) ListUpstreamMultiplierSamples(ctx context.Context, filter *service.OpsUpstreamMultiplierSamplesFilter) ([]*service.OpsUpstreamMultiplierSample, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}
	if filter == nil {
		filter = &service.OpsUpstreamMultiplierSamplesFilter{}
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}
	const q = `
SELECT ` + opsUpstreamMultiplierSampleColumns + `
FROM ops_upstream_multiplier_samples
WHERE model = $1
  AND ($2::BIGINT IS NULL OR account_id = $2)
ORDER BY measured_at DESC, id DESC
LIMIT $3`
	rows, err := r.db.QueryContext(ctx, q, filter.Model, filter.AccountID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*service.OpsUpstreamMultiplierSample, 0, limit)
	for rows.Next() {
		item, err := scanUpstreamMultiplierSample(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *opsRepository) GetLatestUpstreamMultiplierSamples(ctx context.Context, model string, accountIDs []int64) (map[int64]*service.OpsUpstreamMultiplierSample, error) {
	result := make(map[int64]*service.OpsUpstreamMultiplierSample, len(accountIDs))
	if r == nil || r.db == nil {
		return result, fmt.Errorf("nil ops repository")
	}
	if len(accountIDs) == 0 {
		return result, nil
	}
	const q = `
SELECT DISTINCT ON (account_id) ` + opsUpstreamMultiplierSampleColumns + `
FROM ops_upstream_multiplier_samples
WHERE model = $1
  AND account_id = ANY($2)
ORDER BY account_id, measured_at DESC, id DESC`
	rows, err := r.db.QueryContext(ctx, q, model, pq.Array(accountIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		item, err := scanUpstreamMultiplierSample(rows)
		if err != nil {
			return nil, err
		}
		result[item.AccountID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func scanUpstreamMultiplierSample(scanner interface{ Scan(...any) error }) (*service.OpsUpstreamMultiplierSample, error) {
	var httpStatus sql.NullInt64
	var standardCostDelta, actualCostDelta, multiplier sql.NullFloat64
	var balanceBefore, balanceAfter sql.NullFloat64
	item := &service.OpsUpstreamMultiplierSample{}
	if err := scanner.Scan(
		&item.ID,
		&item.AccountID,
		&item.AccountNameSnapshot,
		&item.Platform,
		&item.BaseURLSnapshot,
		&item.KeyPrefixSnapshot,
		&item.Model,
		&item.Status,
		&httpStatus,
		&standardCostDelta,
		&actualCostDelta,
		&multiplier,
		&balanceBefore,
		&balanceAfter,
		&item.ErrorMessage,
		&item.MeasuredAt,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}
	if httpStatus.Valid {
		v := int(httpStatus.Int64)
		item.HTTPStatus = &v
	}
	if standardCostDelta.Valid {
		v := standardCostDelta.Float64
		item.StandardCostDelta = &v
	}
	if actualCostDelta.Valid {
		v := actualCostDelta.Float64
		item.ActualCostDelta = &v
	}
	if multiplier.Valid {
		v := multiplier.Float64
		item.Multiplier = &v
	}
	if balanceBefore.Valid {
		v := balanceBefore.Float64
		item.BalanceBefore = &v
	}
	if balanceAfter.Valid {
		v := balanceAfter.Float64
		item.BalanceAfter = &v
	}
	item.MeasuredAt = item.MeasuredAt.UTC()
	item.CreatedAt = item.CreatedAt.UTC()
	return item, nil
}
