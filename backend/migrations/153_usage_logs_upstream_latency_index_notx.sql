CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_account_created_upstream_latency
    ON usage_logs (account_id, created_at DESC)
    WHERE upstream_latency_ms IS NOT NULL;
