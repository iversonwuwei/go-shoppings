-- 给租户增加平台单独授予的功能列，便于 super 不调整套餐也能开通特定模块
ALTER TABLE "tenants"
    ADD COLUMN IF NOT EXISTS "extra_features" jsonb NOT NULL DEFAULT '[]'::jsonb;
