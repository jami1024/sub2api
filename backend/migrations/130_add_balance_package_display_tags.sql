ALTER TABLE balance_packages
    ADD COLUMN IF NOT EXISTS display_tags JSONB NOT NULL DEFAULT '[]'::jsonb;
