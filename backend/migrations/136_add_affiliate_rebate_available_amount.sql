ALTER TABLE affiliate_rebate_records
    ADD COLUMN IF NOT EXISTS available_amount DECIMAL(20,8) NOT NULL DEFAULT 0;

COMMENT ON COLUMN affiliate_rebate_records.available_amount IS '返利明细当前仍可提现的剩余金额，已扣除负债抵扣、退款冲正和提现占用';

WITH withdrawal_allocations AS (
    SELECT rebate_record_id,
           COALESCE(SUM(amount), 0) AS allocated_amount
    FROM affiliate_withdrawal_request_items
    GROUP BY rebate_record_id
)
UPDATE affiliate_rebate_records r
SET available_amount = GREATEST(
        r.rebate_amount
        - r.debt_amount
        - r.reversed_amount
        - COALESCE(w.allocated_amount, 0),
        0
    ),
    updated_at = NOW()
FROM withdrawal_allocations w
WHERE r.id = w.rebate_record_id
  AND r.status IN ('available', 'withdraw_requested')
  AND r.available_amount = 0;

UPDATE affiliate_rebate_records r
SET available_amount = GREATEST(r.rebate_amount - r.debt_amount - r.reversed_amount, 0),
    updated_at = NOW()
WHERE r.status = 'available'
  AND r.available_amount = 0;
