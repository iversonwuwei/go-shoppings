-- 20260421_admin_tenant.sql
-- 为 admins 表补充 tenant_id 字段，用于区分平台管理员（tenant_id=0）与租户管理员（tenant_id>0）
-- 以及根据手机号进行跨租户的唯一定位。

ALTER TABLE "admins" ADD COLUMN IF NOT EXISTS "tenant_id" BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS "idx_admins_tenant_phone" ON "admins" ("tenant_id", "phone");
CREATE INDEX IF NOT EXISTS "idx_admins_phone" ON "admins" ("phone");
