ALTER TABLE user_affiliates
    ADD COLUMN IF NOT EXISTS debt_quota DECIMAL(20,8) NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_user_affiliates_debt_quota
    ON user_affiliates(debt_quota);

COMMENT ON COLUMN user_affiliates.debt_quota IS '已打款返利被退款后形成的待抵扣负债';
