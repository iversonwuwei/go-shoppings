-- 拼团功能
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
CREATE INDEX IF NOT EXISTS "idx_groupon_activities_status" ON "groupon_activities" ("tenant_id","status");

CREATE TABLE IF NOT EXISTS "groupons" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL,
    "activity_id"     BIGINT NOT NULL,
    "leader_id"       BIGINT NOT NULL,
    "require_num"     INT NOT NULL,
    "current_num"     INT NOT NULL DEFAULT 1,
    "status"          SMALLINT NOT NULL DEFAULT 1, -- 1进行中 2已成团 3已失败
    "expires_at"      TIMESTAMP NOT NULL,
    "succeed_at"      TIMESTAMP,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_groupons_tenant_activity" ON "groupons" ("tenant_id","activity_id");
CREATE INDEX IF NOT EXISTS "idx_groupons_status" ON "groupons" ("tenant_id","status");

CREATE TABLE IF NOT EXISTS "groupon_members" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "groupon_id"  BIGINT NOT NULL,
    "member_id"   BIGINT NOT NULL,
    "order_id"    BIGINT NOT NULL DEFAULT 0,
    "is_leader"   SMALLINT NOT NULL DEFAULT 0,
    "joined_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_groupon_members_groupon" ON "groupon_members" ("tenant_id","groupon_id");
CREATE INDEX IF NOT EXISTS "idx_groupon_members_member" ON "groupon_members" ("tenant_id","member_id");
