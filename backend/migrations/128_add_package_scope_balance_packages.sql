-- Package-scope balance packages
-- 现状兼容前提：现有用户与现有标准余额线路均属于 codex。
-- 因此历史标准分组与有余额用户回填到 codex；余额为 0 的历史用户保留 NULL。

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS package_scope VARCHAR(20);

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS package_scope VARCHAR(20);

CREATE TABLE IF NOT EXISTS balance_packages (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price DECIMAL(20,2) NOT NULL,
    credit_amount DECIMAL(20,8) NOT NULL,
    package_scope VARCHAR(20) NOT NULL,
    product_name VARCHAR(100) NOT NULL DEFAULT '',
    for_sale BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_balance_packages_scope_sale
    ON balance_packages(package_scope, for_sale);

CREATE INDEX IF NOT EXISTS idx_balance_packages_sort_order
    ON balance_packages(sort_order);

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS balance_package_id BIGINT;

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS package_scope_snapshot VARCHAR(20) NOT NULL DEFAULT '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_payment_orders_balance_package'
    ) THEN
        ALTER TABLE payment_orders
            ADD CONSTRAINT fk_payment_orders_balance_package
            FOREIGN KEY (balance_package_id)
            REFERENCES balance_packages(id)
            ON DELETE RESTRICT;
    END IF;
END $$;

UPDATE groups
SET package_scope = 'codex'
WHERE package_scope IS NULL
  AND subscription_type = 'standard';

UPDATE users
SET package_scope = 'codex'
WHERE package_scope IS NULL
  AND balance > 0;
