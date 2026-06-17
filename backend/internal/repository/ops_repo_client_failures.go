package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	clientFailureStatsDefaultLimit = 50
	clientFailureStatsMaxLimit     = 100
)

func (r *opsRepository) GetClientFailureStats(ctx context.Context, filter *service.OpsClientFailureStatsFilter) (*service.OpsClientFailureStatsResponse, error) {
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
	limit := filter.Limit
	if limit <= 0 {
		limit = clientFailureStatsDefaultLimit
	}
	if limit > clientFailureStatsMaxLimit {
		limit = clientFailureStatsMaxLimit
	}

	items, err := r.queryClientFailureStats(ctx, start, end, limit)
	if err != nil {
		return nil, err
	}
	return &service.OpsClientFailureStatsResponse{
		StartTime: start,
		EndTime:   end,
		Items:     items,
	}, nil
}

func (r *opsRepository) queryClientFailureStats(ctx context.Context, start, end time.Time, limit int) ([]*service.OpsClientFailureStatsItem, error) {
	const q = `
WITH client_failures AS (
  SELECT
    e.user_id,
    COALESCE(u.email, '') AS user_email,
    e.api_key_id,
    COALESCE(NULLIF(e.error_message, ''), NULLIF(e.error_type, ''), 'unknown') AS error_message,
    COALESCE(NULLIF(e.inbound_endpoint, ''), NULLIF(e.request_path, ''), 'unknown') AS inbound_endpoint,
    COALESCE(NULLIF(e.platform, ''), 'unknown') AS platform,
    e.created_at
  FROM ops_error_logs e
  LEFT JOIN users u ON u.id = e.user_id
  WHERE e.created_at >= $1 AND e.created_at < $2
    AND (
      e.error_owner = 'client'
      OR e.error_source = 'client_request'
      OR e.error_message IN ('Failed to read request body', 'Request body is empty', 'Failed to parse request body')
      OR e.error_phase IN ('auth', 'request', 'validation')
    )
),
user_stats AS (
  SELECT
    user_id,
    user_email,
    COUNT(*) AS failure_count,
    COUNT(DISTINCT api_key_id) AS affected_key_count,
    MAX(created_at) AS last_seen
  FROM client_failures
  GROUP BY user_id, user_email
),
error_rank AS (
  SELECT user_id, user_email, error_message, COUNT(*) AS error_count,
         ROW_NUMBER() OVER (PARTITION BY user_id, user_email ORDER BY COUNT(*) DESC, error_message ASC) AS rn
  FROM client_failures
  GROUP BY user_id, user_email, error_message
),
endpoint_rank AS (
  SELECT user_id, user_email, inbound_endpoint, COUNT(*) AS endpoint_count,
         ROW_NUMBER() OVER (PARTITION BY user_id, user_email ORDER BY COUNT(*) DESC, inbound_endpoint ASC) AS rn
  FROM client_failures
  GROUP BY user_id, user_email, inbound_endpoint
),
platform_rank AS (
  SELECT user_id, user_email, platform, COUNT(*) AS platform_count,
         ROW_NUMBER() OVER (PARTITION BY user_id, user_email ORDER BY COUNT(*) DESC, platform ASC) AS rn
  FROM client_failures
  GROUP BY user_id, user_email, platform
)
SELECT
  s.user_id,
  s.user_email,
  s.failure_count,
  s.affected_key_count,
  er.error_message AS top_error_message,
  er.error_count AS top_error_count,
  s.last_seen,
  ep.inbound_endpoint AS top_inbound_endpoint,
  pr.platform AS top_platform
FROM user_stats s
LEFT JOIN error_rank er ON er.user_id IS NOT DISTINCT FROM s.user_id AND er.user_email = s.user_email AND er.rn = 1
LEFT JOIN endpoint_rank ep ON ep.user_id IS NOT DISTINCT FROM s.user_id AND ep.user_email = s.user_email AND ep.rn = 1
LEFT JOIN platform_rank pr ON pr.user_id IS NOT DISTINCT FROM s.user_id AND pr.user_email = s.user_email AND pr.rn = 1
ORDER BY s.failure_count DESC, s.last_seen DESC
LIMIT $3`

	rows, err := r.db.QueryContext(ctx, q, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*service.OpsClientFailureStatsItem, 0, limit)
	for rows.Next() {
		item, err := scanClientFailureStatsItem(rows)
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

func scanClientFailureStatsItem(rows interface{ Scan(...any) error }) (*service.OpsClientFailureStatsItem, error) {
	var userID sql.NullInt64
	var lastSeen sql.NullTime
	item := &service.OpsClientFailureStatsItem{}
	if err := rows.Scan(
		&userID,
		&item.UserEmail,
		&item.FailureCount,
		&item.AffectedKeyCount,
		&item.TopErrorMessage,
		&item.TopErrorCount,
		&lastSeen,
		&item.TopInboundEndpoint,
		&item.TopPlatform,
	); err != nil {
		return nil, err
	}
	if userID.Valid {
		v := userID.Int64
		item.UserID = &v
	}
	if lastSeen.Valid {
		v := lastSeen.Time.UTC()
		item.LastSeen = &v
	}
	return item, nil
}
