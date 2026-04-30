-- 20260421_tenant_subscription.sql
-- 租户套餐订阅订单（向平台统一商户号付款）

CREATE TABLE IF NOT EXISTS "tenant_subscription_orders" (
    "id"                BIGSERIAL PRIMARY KEY,
    "tenant_id"         BIGINT NOT NULL,
    "plan_id"           BIGINT NOT NULL,
    "billing_cycle"     VARCHAR(10) NOT NULL,           -- monthly / yearly
    "amount"            NUMERIC(10,2) NOT NULL,          -- 订单金额（元）
    "status"            SMALLINT NOT NULL DEFAULT 0,    -- 0待支付 1已支付 2已取消 3已退款
    "order_no"          VARCHAR(64) NOT NULL UNIQUE,    -- 本地订单号（同时作为 wxpay out_trade_no）
    "created_by_admin_id" BIGINT NOT NULL DEFAULT 0,     -- 创建订单的商户管理员 ID
    "created_by_admin_username" VARCHAR(50) NOT NULL DEFAULT '', -- 创建订单的用户名快照
    "pay_transaction_id" VARCHAR(64) NOT NULL DEFAULT '',
    "pay_at"            TIMESTAMP NULL,
    "expire_before"     TIMESTAMP NULL,                 -- 支付前记录的到期时间
    "expire_after"      TIMESTAMP NULL,                 -- 支付后延长到的到期时间
    "created_at"        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS "idx_tenant_sub_orders_tenant" ON "tenant_subscription_orders" ("tenant_id", "status");
CREATE INDEX IF NOT EXISTS "idx_tenant_sub_orders_created" ON "tenant_subscription_orders" ("created_at" DESC);
CREATE INDEX IF NOT EXISTS "idx_tenant_sub_orders_creator" ON "tenant_subscription_orders" ("tenant_id", "created_by_admin_id", "created_at" DESC);
