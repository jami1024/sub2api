CREATE TABLE IF NOT EXISTS ops_upstream_multiplier_samples (
  id BIGSERIAL PRIMARY KEY,
  account_id BIGINT NOT NULL,
  account_name_snapshot TEXT NOT NULL DEFAULT '',
  platform VARCHAR(32) NOT NULL DEFAULT '',
  base_url_snapshot TEXT NOT NULL DEFAULT '',
  key_prefix_snapshot VARCHAR(16) NOT NULL DEFAULT '',
  model TEXT NOT NULL DEFAULT '',
  status VARCHAR(32) NOT NULL DEFAULT '',
  http_status INT,
  standard_cost_delta NUMERIC,
  actual_cost_delta NUMERIC,
  multiplier NUMERIC,
  balance_before NUMERIC,
  balance_after NUMERIC,
  error_message TEXT NOT NULL DEFAULT '',
  measured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ops_upstream_multiplier_samples_account_time
  ON ops_upstream_multiplier_samples (account_id, measured_at DESC);

CREATE INDEX IF NOT EXISTS idx_ops_upstream_multiplier_samples_model_time
  ON ops_upstream_multiplier_samples (model, measured_at DESC);
