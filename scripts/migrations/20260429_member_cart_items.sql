CREATE TABLE IF NOT EXISTS "member_cart_items" (
    "id"         BIGSERIAL PRIMARY KEY,
    "tenant_id"  BIGINT NOT NULL,
    "member_id"  BIGINT NOT NULL,
    "product_id" BIGINT NOT NULL,
    "sku_id"     BIGINT NOT NULL DEFAULT 0,
    "quantity"   INT NOT NULL DEFAULT 1,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS "uniq_member_cart_item"
    ON "member_cart_items" ("tenant_id", "member_id", "product_id", "sku_id");
CREATE INDEX IF NOT EXISTS "idx_member_cart_items_member"
    ON "member_cart_items" ("tenant_id", "member_id", "updated_at" DESC);
CREATE INDEX IF NOT EXISTS "idx_member_cart_items_product"
    ON "member_cart_items" ("tenant_id", "product_id");
