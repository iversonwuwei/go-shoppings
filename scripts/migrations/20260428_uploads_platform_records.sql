-- 平台侧上传记录使用 tenant_id=0；上传记录本身需要同时支持平台与租户作用域。
ALTER TABLE "uploads"
    DROP CONSTRAINT IF EXISTS "uploads_tenant_id_fkey";

ALTER TABLE "uploads"
    ALTER COLUMN "tenant_id" SET DEFAULT 0;
