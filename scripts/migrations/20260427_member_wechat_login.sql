-- 微信会员登录以 tenant_id + openid 识别用户；手机号不是注册前置条件。
-- 允许多个未绑定手机号的微信会员存在，但已绑定手机号仍需在租户内唯一。

UPDATE "members"
SET "phone" = NULL
WHERE "phone" = '';

ALTER TABLE "members" DROP CONSTRAINT IF EXISTS "members_tenant_id_phone_key";

CREATE UNIQUE INDEX IF NOT EXISTS "uk_members_tenant_phone_present"
    ON "members" ("tenant_id", "phone")
    WHERE "phone" IS NOT NULL AND "phone" <> '';