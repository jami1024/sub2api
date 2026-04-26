CREATE TABLE IF NOT EXISTS affiliate_rebate_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_order_id BIGINT NOT NULL REFERENCES payment_orders(id) ON DELETE CASCADE,
    level SMALLINT NOT NULL,
    rate DECIMAL(10,4) NOT NULL,
    base_amount DECIMAL(20,8) NOT NULL,
    rebate_amount DECIMAL(20,8) NOT NULL,
    status VARCHAR(32) NOT NULL,
    available_at TIMESTAMPTZ NOT NULL,
    reversed_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    debt_amount DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_affiliate_rebate_records_user_order_level UNIQUE (user_id, source_order_id, level)
);

CREATE INDEX IF NOT EXISTS idx_affiliate_rebate_records_user_status
    ON affiliate_rebate_records(user_id, status);

CREATE INDEX IF NOT EXISTS idx_affiliate_rebate_records_source_order
    ON affiliate_rebate_records(source_order_id);

COMMENT ON TABLE affiliate_rebate_records IS '多层邀请返利明细账本';
COMMENT ON COLUMN affiliate_rebate_records.status IS 'pending|available|withdraw_requested|withdraw_paid|cancelled|reversed|debt_offset';
