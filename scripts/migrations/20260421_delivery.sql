-- 配送设置（快递/同城/自提合一，按 feature 分段启用）
CREATE TABLE IF NOT EXISTS "delivery_settings" (
    "tenant_id"          BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    -- 快递
    "express_enabled"    SMALLINT NOT NULL DEFAULT 1,
    "express_free_amount" NUMERIC(10,2) NOT NULL DEFAULT 0, -- 满额包邮，0 为关闭
    "express_default_fee" NUMERIC(10,2) NOT NULL DEFAULT 0,
    -- 同城配送
    "city_enabled"       SMALLINT NOT NULL DEFAULT 0,
    "city_radius_km"     NUMERIC(6,2) NOT NULL DEFAULT 5,
    "city_base_fee"      NUMERIC(10,2) NOT NULL DEFAULT 5,
    "city_per_km_fee"    NUMERIC(10,2) NOT NULL DEFAULT 1,
    "city_min_order"     NUMERIC(10,2) NOT NULL DEFAULT 0,
    -- 自提
    "pickup_enabled"     SMALLINT NOT NULL DEFAULT 0,
    "pickup_address"     VARCHAR(255) NOT NULL DEFAULT '',
    "pickup_hours"       VARCHAR(100) NOT NULL DEFAULT '',
    "pickup_phone"       VARCHAR(30) NOT NULL DEFAULT '',
    "remark"             VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
