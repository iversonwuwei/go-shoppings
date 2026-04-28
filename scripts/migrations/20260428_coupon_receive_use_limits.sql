-- 优惠券领取限制与使用限制拆分
ALTER TABLE "coupons"
    ADD COLUMN IF NOT EXISTS "receive_limit_type" VARCHAR(20) NOT NULL DEFAULT 'limited';

ALTER TABLE "coupons"
    ADD COLUMN IF NOT EXISTS "use_limit" INT NOT NULL DEFAULT 1;

ALTER TABLE "member_coupons"
    ADD COLUMN IF NOT EXISTS "use_limit" INT NOT NULL DEFAULT 1;

UPDATE "coupons"
SET "receive_limit_type" = 'limited'
WHERE "receive_limit_type" IS NULL OR "receive_limit_type" = '';

UPDATE "coupons"
SET "use_limit" = 1
WHERE "use_limit" IS NULL;

UPDATE "member_coupons"
SET "use_limit" = 1
WHERE "use_limit" IS NULL;
