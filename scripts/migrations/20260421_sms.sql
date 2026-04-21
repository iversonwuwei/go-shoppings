-- 短信通知功能
CREATE TABLE IF NOT EXISTS "sms_settings" (
    "tenant_id"    BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "enabled"      SMALLINT NOT NULL DEFAULT 0,
    "provider"     VARCHAR(32) NOT NULL DEFAULT 'aliyun', -- aliyun / tencent
    "access_key"   VARCHAR(128) NOT NULL DEFAULT '',
    "access_secret" VARCHAR(256) NOT NULL DEFAULT '',
    "sign_name"    VARCHAR(64) NOT NULL DEFAULT '',
    "region"       VARCHAR(32) NOT NULL DEFAULT '',
    "remark"       VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "sms_templates" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "code"        VARCHAR(64) NOT NULL, -- 业务码：order_paid, order_shipped, verify_code...
    "name"        VARCHAR(100) NOT NULL,
    "template_id" VARCHAR(64) NOT NULL DEFAULT '', -- 运营商模板ID
    "content"     VARCHAR(500) NOT NULL DEFAULT '',
    "enabled"     SMALLINT NOT NULL DEFAULT 1,
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS "uk_sms_templates_tenant_code" ON "sms_templates" ("tenant_id","code");

CREATE TABLE IF NOT EXISTS "sms_logs" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "phone"       VARCHAR(20) NOT NULL,
    "code"        VARCHAR(64) NOT NULL,
    "content"     VARCHAR(500) NOT NULL DEFAULT '',
    "status"      SMALLINT NOT NULL DEFAULT 1, -- 1成功 2失败
    "error"       VARCHAR(500) NOT NULL DEFAULT '',
    "biz_id"      VARCHAR(64) NOT NULL DEFAULT '',
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_sms_logs_tenant" ON "sms_logs" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_sms_logs_phone" ON "sms_logs" ("tenant_id","phone");
