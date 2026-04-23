CREATE TABLE IF NOT EXISTS "tenant_category_assets" (
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "category_id"     BIGINT NOT NULL REFERENCES "product_categories"("id"),
    "icon"            VARCHAR(255),
    "cover_image"     VARCHAR(255),
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("tenant_id", "category_id")
);

CREATE INDEX IF NOT EXISTS "idx_tenant_category_assets_category"
    ON "tenant_category_assets" ("category_id");
