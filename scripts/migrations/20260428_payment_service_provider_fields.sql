-- 2026-04-28 服务商支付字段
-- 目标：平台订阅付款和顾客订单付款账务隔离；顾客订单生产支付使用服务商 + 子商户。

ALTER TABLE IF EXISTS "platform_settings"
    ADD COLUMN IF NOT EXISTS "sp_appid" TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS "sp_mchid" TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS "sp_apiv3_key" TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS "sp_cert_serial" TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS "partner_notify_url" TEXT DEFAULT '';

ALTER TABLE IF EXISTS "tenant_payment_configs"
    ADD COLUMN IF NOT EXISTS "sp_appid" VARCHAR(64) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "sp_mchid" VARCHAR(64) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "sub_appid" VARCHAR(64) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "sub_mchid" VARCHAR(64) DEFAULT '';

ALTER TABLE IF EXISTS "payments"
    ADD COLUMN IF NOT EXISTS "pay_scene" VARCHAR(32) NOT NULL DEFAULT 'member_order',
    ADD COLUMN IF NOT EXISTS "sp_appid" VARCHAR(64) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "sp_mchid" VARCHAR(64) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "sub_appid" VARCHAR(64) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "sub_mchid" VARCHAR(64) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "settlement_tenant_id" BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS "idx_tpc_sub_mchid"
    ON "tenant_payment_configs" ("provider", "sub_mchid");

CREATE INDEX IF NOT EXISTS "idx_payments_pay_scene"
    ON "payments" ("pay_scene");

CREATE INDEX IF NOT EXISTS "idx_payments_settlement_tenant"
    ON "payments" ("settlement_tenant_id");

-- Verification SQL:
-- SELECT table_name, column_name
-- FROM information_schema.columns
-- WHERE table_name IN ('platform_settings', 'tenant_payment_configs', 'payments')
--   AND column_name IN (
--     'sp_appid', 'sp_mchid', 'sp_apiv3_key', 'sp_cert_serial', 'partner_notify_url',
--     'sub_appid', 'sub_mchid', 'pay_scene', 'settlement_tenant_id'
--   )
-- ORDER BY table_name, column_name;
