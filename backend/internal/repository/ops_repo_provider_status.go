package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

const (
	providerStatusDefaultLimit = 50
	providerStatusMaxLimit     = 100
)

func (r *opsRepository) GetProviderStatus(ctx context.Context, filter *service.OpsProviderStatusFilter) (*service.OpsProviderStatusResponse, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}
	if filter == nil {
		return nil, fmt.Errorf("nil filter")
	}
	if filter.StartTime.IsZero() || filter.EndTime.IsZero() || !filter.StartTime.Before(filter.EndTime) {
		return nil, fmt.Errorf("valid start_time/end_time required")
	}

	start := filter.StartTime.UTC()
	end := filter.EndTime.UTC()
	bucketSeconds := filter.BucketSeconds
	if bucketSeconds <= 0 {
		bucketSeconds = 60
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = providerStatusDefaultLimit
	}
	if limit > providerStatusMaxLimit {
		limit = providerStatusMaxLimit
	}

	items, err := r.queryProviderStatusSummary(ctx, start, end, limit)
	if err != nil {
		return nil, err
	}
	timeline, err := r.queryProviderStatusTimeline(ctx, start, end, bucketSeconds, providerNames(items))
	if err != nil {
		return nil, err
	}
	attachProviderTimeline(items, timeline)

	return &service.OpsProviderStatusResponse{
		StartTime:     start,
		EndTime:       end,
		BucketSeconds: bucketSeconds,
		Items:         items,
		Timeline:      timeline,
	}, nil
}

func (r *opsRepository) queryProviderStatusSummary(ctx context.Context, start, end time.Time, limit int) ([]*service.OpsProviderStatusSummaryItem, error) {
	const q = `
WITH success_rows AS (
  SELECT
    COALESCE(NULLIF(a.name, ''), NULLIF(a.platform, ''), NULLIF(g.platform, ''), 'unknown') AS provider,
    ul.created_at,
    ul.duration_ms
  FROM usage_logs ul
  LEFT JOIN groups g ON g.id = ul.group_id
  LEFT JOIN accounts a ON a.id = ul.account_id
  WHERE ul.created_at >= $1 AND ul.created_at < $2
),
error_rows AS (
  SELECT
    COALESCE(NULLIF(a.name, ''), NULLIF(a.platform, ''), NULLIF(g.platform, ''), 'unknown') AS provider,
    oel.created_at,
    oel.time_to_first_token_ms,
    oel.is_business_limited,
    COALESCE(oel.status_code, 0) AS status_code
  FROM ops_error_logs oel
  LEFT JOIN groups g ON g.id = oel.group_id
  LEFT JOIN accounts a ON a.id = oel.account_id
  WHERE oel.created_at >= $1 AND oel.created_at < $2
    AND oel.account_id IS NOT NULL
    AND oel.is_count_tokens = FALSE
),
providers AS (
  SELECT provider FROM success_rows
  UNION
  SELECT provider FROM error_rows WHERE status_code >= 400
),
stats AS (
  SELECT
    p.provider,
    COALESCE(s.success_count, 0) AS success_count,
    COALESCE(e.failure_count, 0) AS failure_count,
    COALESCE(e.business_limited_count, 0) AS business_limited_count,
    GREATEST(s.last_seen, e.last_seen) AS last_seen,
    s.p50_ms,
    s.p95_ms,
    s.p99_ms
  FROM providers p
  LEFT JOIN (
    SELECT provider,
           COUNT(*) AS success_count,
           MAX(created_at) AS last_seen,
           percentile_cont(0.50) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL) AS p50_ms,
           percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL) AS p95_ms,
           percentile_cont(0.99) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL) AS p99_ms
    FROM success_rows
    GROUP BY provider
  ) s ON s.provider = p.provider
  LEFT JOIN (
    SELECT provider,
           COUNT(*) FILTER (WHERE status_code >= 400 AND NOT is_business_limited) AS failure_count,
           COUNT(*) FILTER (WHERE status_code >= 400 AND is_business_limited) AS business_limited_count,
           MAX(created_at) AS last_seen
    FROM error_rows
    WHERE status_code >= 400
    GROUP BY provider
  ) e ON e.provider = p.provider
)
SELECT provider,
       (success_count + failure_count) AS request_count,
       success_count,
       failure_count,
       business_limited_count,
       p50_ms,
       p95_ms,
       p99_ms,
       last_seen
FROM stats
WHERE (success_count + failure_count + business_limited_count) > 0
ORDER BY (success_count + failure_count) DESC, provider ASC
LIMIT $3`

	rows, err := r.db.QueryContext(ctx, q, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*service.OpsProviderStatusSummaryItem, 0, limit)
	for rows.Next() {
		item, err := scanProviderStatusSummary(rows)
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

func (r *opsRepository) queryProviderStatusTimeline(ctx context.Context, start, end time.Time, bucketSeconds int, providers []string) ([]*service.OpsProviderStatusTimelinePoint, error) {
	if len(providers) == 0 {
		return []*service.OpsProviderStatusTimelinePoint{}, nil
	}
	const q = `
WITH selected_providers AS (
  SELECT unnest($4::text[]) AS provider
),
buckets AS (
  SELECT sp.provider,
         gs.bucket_idx,
         $1::timestamptz + (gs.bucket_idx * ($3::int * interval '1 second')) AS bucket_start,
         $1::timestamptz + ((gs.bucket_idx + 1) * ($3::int * interval '1 second')) AS bucket_end
  FROM selected_providers sp
  CROSS JOIN generate_series(0, CEIL(EXTRACT(EPOCH FROM ($2::timestamptz - $1::timestamptz)) / $3::int)::int - 1) AS gs(bucket_idx)
),
success_rows AS (
  SELECT
    COALESCE(NULLIF(a.name, ''), NULLIF(a.platform, ''), NULLIF(g.platform, ''), 'unknown') AS provider,
    ul.created_at,
    ul.duration_ms
  FROM usage_logs ul
  LEFT JOIN groups g ON g.id = ul.group_id
  LEFT JOIN accounts a ON a.id = ul.account_id
  WHERE ul.created_at >= $1 AND ul.created_at < $2
),
error_rows AS (
  SELECT
    COALESCE(NULLIF(a.name, ''), NULLIF(a.platform, ''), NULLIF(g.platform, ''), 'unknown') AS provider,
    oel.created_at,
    COALESCE(oel.status_code, 0) AS status_code,
    oel.is_business_limited
  FROM ops_error_logs oel
  LEFT JOIN groups g ON g.id = oel.group_id
  LEFT JOIN accounts a ON a.id = oel.account_id
  WHERE oel.created_at >= $1 AND oel.created_at < $2
    AND oel.account_id IS NOT NULL
    AND oel.is_count_tokens = FALSE
),
bucket_start_expr AS (
  SELECT
    provider,
    created_at,
    duration_ms,
    $1::timestamptz + (FLOOR(EXTRACT(EPOCH FROM (created_at - $1::timestamptz)) / $3::int) * ($3::int * interval '1 second')) AS bucket_start
  FROM success_rows
  WHERE provider = ANY($4::text[])
),
bucketed_success AS (
  SELECT
    provider,
    bucket_start,
    COUNT(*) AS success_count,
    percentile_cont(0.50) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL) AS p50_ms,
    percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL) AS p95_ms,
    percentile_cont(0.99) WITHIN GROUP (ORDER BY duration_ms) FILTER (WHERE duration_ms IS NOT NULL) AS p99_ms
  FROM bucket_start_expr
  GROUP BY provider, bucket_start
),
bucketed_error AS (
  SELECT
    provider,
    $1::timestamptz + (FLOOR(EXTRACT(EPOCH FROM (created_at - $1::timestamptz)) / $3::int) * ($3::int * interval '1 second')) AS bucket_start,
    COUNT(*) AS failure_count
  FROM error_rows
  WHERE provider = ANY($4::text[])
    AND status_code >= 400
    AND NOT is_business_limited
  GROUP BY provider, bucket_start
),
bucket_stats AS (
  SELECT
    b.provider,
    b.bucket_start,
    COALESCE(s.success_count, 0) AS success_count,
    COALESCE(e.failure_count, 0) AS failure_count,
    s.p50_ms,
    s.p95_ms,
    s.p99_ms
  FROM buckets b
  LEFT JOIN bucketed_success s ON s.provider = b.provider AND s.bucket_start = b.bucket_start
  LEFT JOIN bucketed_error e ON e.provider = b.provider AND e.bucket_start = b.bucket_start
)
SELECT provider, bucket_start, (success_count + failure_count) AS request_count, success_count, failure_count, p50_ms, p95_ms, p99_ms
FROM bucket_stats
ORDER BY provider ASC, bucket_start ASC`

	rows, err := r.db.QueryContext(ctx, q, start, end, bucketSeconds, pq.Array(providers))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	points := make([]*service.OpsProviderStatusTimelinePoint, 0)
	for rows.Next() {
		point, err := scanProviderStatusTimelinePoint(rows)
		if err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return points, nil
}

func scanProviderStatusSummary(rows interface{ Scan(...any) error }) (*service.OpsProviderStatusSummaryItem, error) {
	var p50, p95, p99 sql.NullFloat64
	var lastSeen sql.NullTime
	item := &service.OpsProviderStatusSummaryItem{}
	if err := rows.Scan(
		&item.Provider,
		&item.RequestCount,
		&item.SuccessCount,
		&item.FailureCount,
		&item.BusinessLimitedCount,
		&p50,
		&p95,
		&p99,
		&lastSeen,
	); err != nil {
		return nil, err
	}
	item.P50Ms = floatToIntPtr(p50)
	item.P95Ms = floatToIntPtr(p95)
	item.P99Ms = floatToIntPtr(p99)
	if lastSeen.Valid {
		v := lastSeen.Time.UTC()
		item.LastSeen = &v
	}
	denom := item.SuccessCount + item.FailureCount
	item.Availability = roundTo4DP(safeDivideFloat64(float64(item.SuccessCount), float64(denom)) * 100)
	item.ErrorRate = roundTo4DP(safeDivideFloat64(float64(item.FailureCount), float64(denom)) * 100)
	return item, nil
}

func scanProviderStatusTimelinePoint(rows interface{ Scan(...any) error }) (*service.OpsProviderStatusTimelinePoint, error) {
	var p50, p95, p99 sql.NullFloat64
	point := &service.OpsProviderStatusTimelinePoint{}
	if err := rows.Scan(
		&point.Provider,
		&point.BucketStart,
		&point.RequestCount,
		&point.SuccessCount,
		&point.FailureCount,
		&p50,
		&p95,
		&p99,
	); err != nil {
		return nil, err
	}
	point.BucketStart = point.BucketStart.UTC()
	point.P50Ms = floatToIntPtr(p50)
	point.P95Ms = floatToIntPtr(p95)
	point.P99Ms = floatToIntPtr(p99)
	denom := point.SuccessCount + point.FailureCount
	point.Availability = roundTo4DP(safeDivideFloat64(float64(point.SuccessCount), float64(denom)) * 100)
	return point, nil
}

func providerNames(items []*service.OpsProviderStatusSummaryItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item != nil && item.Provider != "" {
			out = append(out, item.Provider)
		}
	}
	return out
}

func attachProviderTimeline(items []*service.OpsProviderStatusSummaryItem, points []*service.OpsProviderStatusTimelinePoint) {
	byProvider := make(map[string][]*service.OpsProviderStatusTimelinePoint)
	for _, point := range points {
		if point == nil {
			continue
		}
		byProvider[point.Provider] = append(byProvider[point.Provider], point)
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		item.Timeline = byProvider[item.Provider]
	}
}
