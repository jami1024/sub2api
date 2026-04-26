CREATE TABLE IF NOT EXISTS affiliate_withdrawal_requests (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20,8) NOT NULL,
    status VARCHAR(32) NOT NULL,
    applicant_note TEXT NOT NULL DEFAULT '',
    admin_note TEXT NOT NULL DEFAULT '',
    reviewed_by BIGINT,
    reviewed_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_affiliate_withdrawal_requests_user_status
    ON affiliate_withdrawal_requests(user_id, status);

CREATE INDEX IF NOT EXISTS idx_affiliate_withdrawal_requests_status_created
    ON affiliate_withdrawal_requests(status, created_at DESC);

COMMENT ON TABLE affiliate_withdrawal_requests IS '邀请返利提现申请（人工打款）';
COMMENT ON COLUMN affiliate_withdrawal_requests.status IS 'pending|approved|rejected|paid';
