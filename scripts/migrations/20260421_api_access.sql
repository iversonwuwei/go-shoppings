-- 开放 API 功能
CREATE TABLE IF NOT EXISTS "api_tokens" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "name"        VARCHAR(100) NOT NULL,
    "app_key"     VARCHAR(64) NOT NULL,
    "app_secret"  VARCHAR(128) NOT NULL,
    "scopes"      VARCHAR(500) NOT NULL DEFAULT '',
    "ip_whitelist" VARCHAR(500) NOT NULL DEFAULT '',
    "status"      SMALLINT NOT NULL DEFAULT 1, -- 1启用 2禁用
    "last_used_at" TIMESTAMP NULL,
    "expires_at"  TIMESTAMP NULL,
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
