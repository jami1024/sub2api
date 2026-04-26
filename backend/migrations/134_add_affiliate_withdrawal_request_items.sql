CREATE TABLE IF NOT EXISTS affiliate_withdrawal_request_items (
    id BIGSERIAL PRIMARY KEY,
    withdrawal_request_id BIGINT NOT NULL REFERENCES affiliate_withdrawal_requests(id) ON DELETE CASCADE,
    rebate_record_id BIGINT NOT NULL REFERENCES affiliate_rebate_records(id) ON DELETE CASCADE,
    amount DECIMAL(20,8) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (withdrawal_request_id, rebate_record_id)
);

CREATE INDEX IF NOT EXISTS idx_affiliate_withdrawal_request_items_request
    ON affiliate_withdrawal_request_items(withdrawal_request_id);

CREATE INDEX IF NOT EXISTS idx_affiliate_withdrawal_request_items_rebate
    ON affiliate_withdrawal_request_items(rebate_record_id);

COMMENT ON TABLE affiliate_withdrawal_request_items IS '提现申请与返利明细的分配关系';
