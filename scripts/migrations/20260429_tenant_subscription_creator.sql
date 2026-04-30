-- 20260429_tenant_subscription_creator.sql
-- 订阅订单创建人快照，便于多管理员场景下审计追溯。

ALTER TABLE "tenant_subscription_orders"
    ADD COLUMN IF NOT EXISTS "created_by_admin_id" BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS "created_by_admin_username" VARCHAR(50) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS "idx_tenant_sub_orders_creator"
    ON "tenant_subscription_orders" ("tenant_id", "created_by_admin_id", "created_at" DESC);
