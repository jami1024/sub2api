-- Speed up OpenAI channel monitor status reads that derive health from recent real usage logs.
-- These partial indexes keep the hot monitor query on small endpoint/model/time ranges.
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_openai_model_created_id
    ON usage_logs (model, created_at DESC, id DESC)
    WHERE actual_cost > 0
      AND (
          inbound_endpoint IN ('/v1/chat/completions', '/v1/responses', '/v1/responses/compact')
          OR upstream_endpoint IN ('/v1/chat/completions', '/v1/responses', '/v1/responses/compact')
      );

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_openai_requested_model_created_id
    ON usage_logs (requested_model, created_at DESC, id DESC)
    WHERE actual_cost > 0
      AND requested_model IS NOT NULL
      AND (
          inbound_endpoint IN ('/v1/chat/completions', '/v1/responses', '/v1/responses/compact')
          OR upstream_endpoint IN ('/v1/chat/completions', '/v1/responses', '/v1/responses/compact')
      );

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_logs_openai_upstream_model_created_id
    ON usage_logs (upstream_model, created_at DESC, id DESC)
    WHERE actual_cost > 0
      AND upstream_model IS NOT NULL
      AND (
          inbound_endpoint IN ('/v1/chat/completions', '/v1/responses', '/v1/responses/compact')
          OR upstream_endpoint IN ('/v1/chat/completions', '/v1/responses', '/v1/responses/compact')
      );
