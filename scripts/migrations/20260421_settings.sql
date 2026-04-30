-- 2026-04-21 新增：商户收款配置 与 物流承运商
CREATE TABLE IF NOT EXISTS "tenant_payment_configs" (
    "id"                BIGSERIAL PRIMARY KEY,
    "tenant_id"         BIGINT NOT NULL,
    "provider"          VARCHAR(20) NOT NULL DEFAULT 'wechat',
    "mch_id"            VARCHAR(64),
    "app_id"            VARCHAR(64),
    "settlement_account_name" VARCHAR(100) DEFAULT '',
    "settlement_account_no"   VARCHAR(128) DEFAULT '',
    "settlement_bank_name"    VARCHAR(100) DEFAULT '',
    "settlement_remark"       VARCHAR(500) DEFAULT '',
    "api_v3_key"        VARCHAR(128),
    "cert_serial_no"    VARCHAR(64),
    "private_key_pem"   TEXT,
    "cert_pem"          TEXT,
    "notify_url"        VARCHAR(255),
    "enabled"           SMALLINT NOT NULL DEFAULT 0,
    "audit_status"      SMALLINT NOT NULL DEFAULT 0,
    "audit_remark"      VARCHAR(500) DEFAULT '',
    "submitted_at"      TIMESTAMPTZ,
    "audited_at"        TIMESTAMPTZ,
    "created_at"        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS "uniq_tpc_tenant_provider"
    ON "tenant_payment_configs" ("tenant_id", "provider");

CREATE TABLE IF NOT EXISTS "shipping_carriers" (
    "id"               BIGSERIAL PRIMARY KEY,
    "tenant_id"        BIGINT NOT NULL,
    "code"             VARCHAR(30) NOT NULL,
    "name"             VARCHAR(50) NOT NULL,
    "api_provider"     VARCHAR(30) DEFAULT '',
    "api_customer"     VARCHAR(128) DEFAULT '',
    "api_key"          VARCHAR(256) DEFAULT '',
    "api_secret"       VARCHAR(256) DEFAULT '',
    "priority"         INT NOT NULL DEFAULT 0,
    "enabled"          SMALLINT NOT NULL DEFAULT 0,
    "audit_status"     SMALLINT NOT NULL DEFAULT 0,
    "audit_remark"     VARCHAR(500) DEFAULT '',
    "submitted_at"     TIMESTAMPTZ,
    "audited_at"       TIMESTAMPTZ,
    "created_at"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS "idx_shipping_carriers_tenant"
    ON "shipping_carriers" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_shipping_carriers_tenant_code"
    ON "shipping_carriers" ("tenant_id", "code");
