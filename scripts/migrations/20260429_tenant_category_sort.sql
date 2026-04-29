ALTER TABLE "tenant_category_assets"
    ADD COLUMN IF NOT EXISTS "sort" INT;

CREATE INDEX IF NOT EXISTS "idx_tenant_category_assets_sort"
    ON "tenant_category_assets" ("tenant_id", "sort" DESC);
