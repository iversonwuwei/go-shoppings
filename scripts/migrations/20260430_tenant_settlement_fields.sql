-- 20260430_tenant_settlement_fields.sql
-- 前期支付改为平台统一收款，租户侧收款配置降级为账期结算资料。

ALTER TABLE "tenant_payment_configs"
    ADD COLUMN IF NOT EXISTS "settlement_account_name" VARCHAR(100) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "settlement_account_no" VARCHAR(128) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "settlement_bank_name" VARCHAR(100) DEFAULT '',
    ADD COLUMN IF NOT EXISTS "settlement_remark" VARCHAR(500) DEFAULT '';