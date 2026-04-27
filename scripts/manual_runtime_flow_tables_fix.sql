-- 手动/本地修复运行态流程缺失表
-- 适用场景：已有本地库只执行过早期 init_db.sql，缺少后续短信、分销、拼团、开放 API、积分规则、配送设置、租户订阅订单等运行态表。
-- 脚本可重复执行；执行后末尾会输出缺失项校验结果。

BEGIN;

SET search_path TO public;

ALTER TABLE "tenants" ADD COLUMN IF NOT EXISTS "billing_cycle" VARCHAR(10) NOT NULL DEFAULT 'yearly';
ALTER TABLE "tenants" ADD COLUMN IF NOT EXISTS "extra_features" JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE TABLE IF NOT EXISTS "tenant_subscription_orders" (
    "id"                 BIGSERIAL PRIMARY KEY,
    "tenant_id"          BIGINT NOT NULL,
    "plan_id"            BIGINT NOT NULL,
    "billing_cycle"      VARCHAR(10) NOT NULL,
    "amount"             NUMERIC(10,2) NOT NULL,
    "status"             SMALLINT NOT NULL DEFAULT 0,
    "order_no"           VARCHAR(64) NOT NULL UNIQUE,
    "pay_transaction_id" VARCHAR(64) NOT NULL DEFAULT '',
    "pay_at"             TIMESTAMP NULL,
    "expire_before"      TIMESTAMP NULL,
    "expire_after"       TIMESTAMP NULL,
    "created_at"         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_tenant_sub_orders_tenant" ON "tenant_subscription_orders" ("tenant_id", "status");
CREATE INDEX IF NOT EXISTS "idx_tenant_sub_orders_created" ON "tenant_subscription_orders" ("created_at" DESC);

CREATE TABLE IF NOT EXISTS "points_settings" (
    "tenant_id"   BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "enabled"     SMALLINT NOT NULL DEFAULT 1,
    "earn_rate"   NUMERIC(10,4) NOT NULL DEFAULT 1,
    "min_amount"  NUMERIC(10,2) NOT NULL DEFAULT 0,
    "redeem_rate" INT NOT NULL DEFAULT 100,
    "remark"      VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "delivery_settings" (
    "tenant_id"            BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "express_enabled"      SMALLINT NOT NULL DEFAULT 1,
    "express_free_amount"  NUMERIC(10,2) NOT NULL DEFAULT 0,
    "express_default_fee"  NUMERIC(10,2) NOT NULL DEFAULT 0,
    "city_enabled"         SMALLINT NOT NULL DEFAULT 0,
    "city_radius_km"       NUMERIC(6,2) NOT NULL DEFAULT 5,
    "city_base_fee"        NUMERIC(10,2) NOT NULL DEFAULT 5,
    "city_per_km_fee"      NUMERIC(10,2) NOT NULL DEFAULT 1,
    "city_min_order"       NUMERIC(10,2) NOT NULL DEFAULT 0,
    "pickup_enabled"       SMALLINT NOT NULL DEFAULT 0,
    "pickup_address"       VARCHAR(255) NOT NULL DEFAULT '',
    "pickup_hours"         VARCHAR(100) NOT NULL DEFAULT '',
    "pickup_phone"         VARCHAR(30) NOT NULL DEFAULT '',
    "remark"               VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "sms_settings" (
    "tenant_id"     BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "enabled"       SMALLINT NOT NULL DEFAULT 0,
    "provider"      VARCHAR(32) NOT NULL DEFAULT 'aliyun',
    "access_key"    VARCHAR(128) NOT NULL DEFAULT '',
    "access_secret" VARCHAR(256) NOT NULL DEFAULT '',
    "sign_name"     VARCHAR(64) NOT NULL DEFAULT '',
    "region"        VARCHAR(32) NOT NULL DEFAULT '',
    "remark"        VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "sms_templates" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "code"        VARCHAR(64) NOT NULL,
    "name"        VARCHAR(100) NOT NULL,
    "template_id" VARCHAR(64) NOT NULL DEFAULT '',
    "content"     VARCHAR(500) NOT NULL DEFAULT '',
    "enabled"     SMALLINT NOT NULL DEFAULT 1,
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS "uk_sms_templates_tenant_code" ON "sms_templates" ("tenant_id", "code");

CREATE TABLE IF NOT EXISTS "sms_logs" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "phone"       VARCHAR(20) NOT NULL,
    "code"        VARCHAR(64) NOT NULL,
    "content"     VARCHAR(500) NOT NULL DEFAULT '',
    "status"      SMALLINT NOT NULL DEFAULT 1,
    "error"       VARCHAR(500) NOT NULL DEFAULT '',
    "biz_id"      VARCHAR(64) NOT NULL DEFAULT '',
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_sms_logs_tenant" ON "sms_logs" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_sms_logs_phone" ON "sms_logs" ("tenant_id", "phone");

CREATE TABLE IF NOT EXISTS "distribution_settings" (
    "tenant_id"     BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "enabled"       SMALLINT NOT NULL DEFAULT 1,
    "level1_rate"   NUMERIC(5,4) NOT NULL DEFAULT 0.10,
    "level2_rate"   NUMERIC(5,4) NOT NULL DEFAULT 0.05,
    "min_withdraw"  NUMERIC(10,2) NOT NULL DEFAULT 10,
    "auto_become"   SMALLINT NOT NULL DEFAULT 0,
    "remark"        VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "distributors" (
    "id"                   BIGSERIAL PRIMARY KEY,
    "tenant_id"            BIGINT NOT NULL,
    "member_id"            BIGINT NOT NULL,
    "parent_id"            BIGINT NOT NULL DEFAULT 0,
    "grandparent_id"       BIGINT NOT NULL DEFAULT 0,
    "status"               SMALLINT NOT NULL DEFAULT 0,
    "total_commission"     NUMERIC(12,2) NOT NULL DEFAULT 0,
    "pending_commission"   NUMERIC(12,2) NOT NULL DEFAULT 0,
    "withdrawn"            NUMERIC(12,2) NOT NULL DEFAULT 0,
    "invite_count"         INT NOT NULL DEFAULT 0,
    "created_at"           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "approved_at"          TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS "uk_distributors_tenant_member" ON "distributors" ("tenant_id", "member_id");
CREATE INDEX IF NOT EXISTS "idx_distributors_parent" ON "distributors" ("tenant_id", "parent_id");
CREATE INDEX IF NOT EXISTS "idx_distributors_status" ON "distributors" ("tenant_id", "status");

CREATE TABLE IF NOT EXISTS "commission_logs" (
    "id"             BIGSERIAL PRIMARY KEY,
    "tenant_id"      BIGINT NOT NULL,
    "distributor_id" BIGINT NOT NULL,
    "member_id"      BIGINT NOT NULL,
    "order_id"       BIGINT NOT NULL,
    "order_no"       VARCHAR(64) NOT NULL,
    "buyer_id"       BIGINT NOT NULL,
    "level"          SMALLINT NOT NULL,
    "amount"         NUMERIC(12,2) NOT NULL,
    "rate"           NUMERIC(5,4) NOT NULL,
    "status"         SMALLINT NOT NULL DEFAULT 1,
    "settled_at"     TIMESTAMP,
    "created_at"     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_commission_logs_tenant" ON "commission_logs" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_commission_logs_distributor" ON "commission_logs" ("tenant_id", "distributor_id");
CREATE INDEX IF NOT EXISTS "idx_commission_logs_order" ON "commission_logs" ("tenant_id", "order_id");

CREATE TABLE IF NOT EXISTS "groupon_activities" (
    "id"             BIGSERIAL PRIMARY KEY,
    "tenant_id"      BIGINT NOT NULL,
    "name"           VARCHAR(100) NOT NULL,
    "product_id"     BIGINT NOT NULL,
    "sku_id"         BIGINT NOT NULL DEFAULT 0,
    "group_price"    NUMERIC(10,2) NOT NULL,
    "original_price" NUMERIC(10,2) NOT NULL,
    "require_num"    INT NOT NULL DEFAULT 2,
    "expire_hours"   INT NOT NULL DEFAULT 24,
    "total_stock"    INT NOT NULL DEFAULT 0,
    "sold_count"     INT NOT NULL DEFAULT 0,
    "start_at"       TIMESTAMP NOT NULL,
    "end_at"         TIMESTAMP NOT NULL,
    "status"         SMALLINT NOT NULL DEFAULT 1,
    "created_at"     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_groupon_activities_tenant" ON "groupon_activities" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_groupon_activities_status" ON "groupon_activities" ("tenant_id", "status");

CREATE TABLE IF NOT EXISTS "groupons" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL,
    "activity_id"     BIGINT NOT NULL,
    "leader_id"       BIGINT NOT NULL,
    "require_num"     INT NOT NULL,
    "current_num"     INT NOT NULL DEFAULT 1,
    "status"          SMALLINT NOT NULL DEFAULT 1,
    "expires_at"      TIMESTAMP NOT NULL,
    "succeed_at"      TIMESTAMP,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_groupons_tenant_activity" ON "groupons" ("tenant_id", "activity_id");
CREATE INDEX IF NOT EXISTS "idx_groupons_status" ON "groupons" ("tenant_id", "status");

CREATE TABLE IF NOT EXISTS "groupon_members" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "groupon_id"  BIGINT NOT NULL,
    "member_id"   BIGINT NOT NULL,
    "order_id"    BIGINT NOT NULL DEFAULT 0,
    "is_leader"   SMALLINT NOT NULL DEFAULT 0,
    "joined_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_groupon_members_groupon" ON "groupon_members" ("tenant_id", "groupon_id");
CREATE INDEX IF NOT EXISTS "idx_groupon_members_member" ON "groupon_members" ("tenant_id", "member_id");

CREATE TABLE IF NOT EXISTS "api_tokens" (
    "id"           BIGSERIAL PRIMARY KEY,
    "tenant_id"    BIGINT NOT NULL,
    "name"         VARCHAR(100) NOT NULL,
    "app_key"      VARCHAR(64) NOT NULL,
    "app_secret"   VARCHAR(128) NOT NULL,
    "scopes"       VARCHAR(500) NOT NULL DEFAULT '',
    "ip_whitelist" VARCHAR(500) NOT NULL DEFAULT '',
    "status"       SMALLINT NOT NULL DEFAULT 1,
    "last_used_at" TIMESTAMP NULL,
    "expires_at"   TIMESTAMP NULL,
    "created_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS "uk_api_tokens_app_key" ON "api_tokens" ("app_key");
CREATE INDEX IF NOT EXISTS "idx_api_tokens_tenant" ON "api_tokens" ("tenant_id");

CREATE TABLE IF NOT EXISTS "api_request_logs" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "token_id"    BIGINT NOT NULL,
    "app_key"     VARCHAR(64) NOT NULL,
    "method"      VARCHAR(10) NOT NULL,
    "path"        VARCHAR(255) NOT NULL,
    "status_code" INT NOT NULL DEFAULT 0,
    "ip"          VARCHAR(64) NOT NULL DEFAULT '',
    "cost_ms"     INT NOT NULL DEFAULT 0,
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_api_logs_tenant" ON "api_request_logs" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_api_logs_token" ON "api_request_logs" ("token_id");

COMMIT;

WITH expected_tables(table_name) AS (
    VALUES
        ('api_request_logs'),
        ('api_tokens'),
        ('commission_logs'),
        ('delivery_settings'),
        ('distribution_settings'),
        ('distributors'),
        ('groupon_activities'),
        ('groupon_members'),
        ('groupons'),
        ('points_settings'),
        ('sms_logs'),
        ('sms_settings'),
        ('sms_templates'),
        ('tenant_subscription_orders')
), expected_columns(table_name, column_name) AS (
    VALUES
        ('tenants', 'billing_cycle'),
        ('tenants', 'extra_features')
)
SELECT 'missing_table' AS kind, table_name AS item
FROM expected_tables e
WHERE NOT EXISTS (
    SELECT 1 FROM information_schema.tables t
    WHERE t.table_schema = 'public' AND t.table_name = e.table_name
)
UNION ALL
SELECT 'missing_column' AS kind, table_name || '.' || column_name AS item
FROM expected_columns e
WHERE NOT EXISTS (
    SELECT 1 FROM information_schema.columns c
    WHERE c.table_schema = 'public' AND c.table_name = e.table_name AND c.column_name = e.column_name
)
ORDER BY kind, item;