-- 平台短信配置使用 tenant_id=0 表示平台自身，不应受租户外键约束。
ALTER TABLE "sms_settings" DROP CONSTRAINT IF EXISTS "sms_settings_tenant_id_fkey";
ALTER TABLE "sms_settings" ALTER COLUMN "tenant_id" SET DEFAULT 0;

COMMENT ON COLUMN "sms_settings"."tenant_id" IS '0 = platform SMS settings; >0 = tenant SMS settings';
