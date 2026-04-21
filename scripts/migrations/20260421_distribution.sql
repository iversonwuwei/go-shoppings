-- 分销功能
CREATE TABLE IF NOT EXISTS "distribution_settings" (
    "tenant_id"    BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "enabled"      SMALLINT NOT NULL DEFAULT 1,
    "level1_rate"  NUMERIC(5,4) NOT NULL DEFAULT 0.10,
    "level2_rate"  NUMERIC(5,4) NOT NULL DEFAULT 0.05,
    "min_withdraw" NUMERIC(10,2) NOT NULL DEFAULT 10,
    "auto_become"  SMALLINT NOT NULL DEFAULT 0, -- 0 需审核 1 下单后自动成为
    "remark"       VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "distributors" (
    "id"             BIGSERIAL PRIMARY KEY,
    "tenant_id"      BIGINT NOT NULL,
    "member_id"      BIGINT NOT NULL,
    "parent_id"      BIGINT NOT NULL DEFAULT 0, -- 上级分销员 member_id
    "grandparent_id" BIGINT NOT NULL DEFAULT 0,
    "status"         SMALLINT NOT NULL DEFAULT 0, -- 0待审核 1正常 2冻结
    "total_commission" NUMERIC(12,2) NOT NULL DEFAULT 0,
    "pending_commission" NUMERIC(12,2) NOT NULL DEFAULT 0,
    "withdrawn"      NUMERIC(12,2) NOT NULL DEFAULT 0,
    "invite_count"   INT NOT NULL DEFAULT 0,
    "created_at"     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "approved_at"    TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS "uk_distributors_tenant_member" ON "distributors" ("tenant_id","member_id");
CREATE INDEX IF NOT EXISTS "idx_distributors_parent" ON "distributors" ("tenant_id","parent_id");
CREATE INDEX IF NOT EXISTS "idx_distributors_status" ON "distributors" ("tenant_id","status");

CREATE TABLE IF NOT EXISTS "commission_logs" (
    "id"             BIGSERIAL PRIMARY KEY,
    "tenant_id"      BIGINT NOT NULL,
    "distributor_id" BIGINT NOT NULL,
    "member_id"      BIGINT NOT NULL,
    "order_id"       BIGINT NOT NULL,
    "order_no"       VARCHAR(64) NOT NULL,
    "buyer_id"       BIGINT NOT NULL,
    "level"          SMALLINT NOT NULL, -- 1 直推 2 间推
    "amount"         NUMERIC(12,2) NOT NULL,
    "rate"           NUMERIC(5,4) NOT NULL,
    "status"         SMALLINT NOT NULL DEFAULT 1, -- 1待结算 2已结算 3已取消
    "settled_at"     TIMESTAMP,
    "created_at"     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_commission_logs_tenant" ON "commission_logs" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_commission_logs_distributor" ON "commission_logs" ("tenant_id","distributor_id");
CREATE INDEX IF NOT EXISTS "idx_commission_logs_order" ON "commission_logs" ("tenant_id","order_id");
