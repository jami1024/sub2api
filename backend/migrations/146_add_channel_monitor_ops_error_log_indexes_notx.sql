-- Speed up channel status reads that count recent OpenAI SLA errors from ops_error_logs.
-- Keep the predicate focused on SLA errors so the indexes stay smaller on high-write tables.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ops_error_logs_channel_monitor_model_created_id
    ON ops_error_logs (model, created_at DESC, id DESC)
    WHERE COALESCE(status_code, 0) >= 400
      AND NOT is_business_limited
      AND is_count_tokens = FALSE
      AND model IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ops_error_logs_channel_monitor_requested_model_created_id
    ON ops_error_logs (requested_model, created_at DESC, id DESC)
    WHERE COALESCE(status_code, 0) >= 400
      AND NOT is_business_limited
      AND is_count_tokens = FALSE
      AND requested_model IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ops_error_logs_channel_monitor_upstream_model_created_id
    ON ops_error_logs (upstream_model, created_at DESC, id DESC)
    WHERE COALESCE(status_code, 0) >= 400
      AND NOT is_business_limited
      AND is_count_tokens = FALSE
      AND upstream_model IS NOT NULL;
