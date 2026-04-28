-- 2026-04-28 售后单表
-- 目标：补齐订单级整单退款/退货退款状态闭环；真实微信退款后续接入支付退款 API。

CREATE TABLE IF NOT EXISTS "after_sale_orders" (
    "id" BIGSERIAL PRIMARY KEY,
    "tenant_id" BIGINT NOT NULL REFERENCES "tenants"("id"),
    "after_sale_no" VARCHAR(32) NOT NULL UNIQUE,
    "order_id" BIGINT NOT NULL REFERENCES "orders"("id"),
    "order_no" VARCHAR(32) NOT NULL,
    "order_item_id" BIGINT NOT NULL DEFAULT 0,
    "member_id" BIGINT NOT NULL,
    "type" VARCHAR(20) NOT NULL,
    "status" VARCHAR(20) NOT NULL DEFAULT 'pending',
    "amount" NUMERIC(10,2) NOT NULL,
    "reason" VARCHAR(120) NOT NULL,
    "description" VARCHAR(500) DEFAULT '',
    "images" JSONB NOT NULL DEFAULT '[]',
    "order_status_before" VARCHAR(20) NOT NULL,
    "audit_remark" VARCHAR(500) DEFAULT '',
    "refund_remark" VARCHAR(500) DEFAULT '',
    "return_express_company" VARCHAR(80) DEFAULT '',
    "return_express_no" VARCHAR(80) DEFAULT '',
    "refund_no" VARCHAR(64) DEFAULT '',
    "applied_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "audited_at" TIMESTAMPTZ,
    "returned_at" TIMESTAMPTZ,
    "received_at" TIMESTAMPTZ,
    "refunded_at" TIMESTAMPTZ,
    "cancelled_at" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_after_sale_tenant" ON "after_sale_orders" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_after_sale_order" ON "after_sale_orders" ("order_id");
CREATE INDEX IF NOT EXISTS "idx_after_sale_member" ON "after_sale_orders" ("member_id");
CREATE INDEX IF NOT EXISTS "idx_after_sale_status" ON "after_sale_orders" ("status");

-- Verification SQL:
-- SELECT column_name
-- FROM information_schema.columns
-- WHERE table_name = 'after_sale_orders'
-- ORDER BY ordinal_position;
