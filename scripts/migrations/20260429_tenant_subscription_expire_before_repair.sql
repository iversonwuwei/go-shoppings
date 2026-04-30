-- 20260429_tenant_subscription_expire_before_repair.sql
-- 修复历史订阅订单 expire_before：按已支付订单的 expire_after 和计费周期倒推实际订阅周期起点。

UPDATE "tenant_subscription_orders"
SET "expire_before" = CASE
    WHEN "billing_cycle" = 'monthly' THEN "expire_after" - INTERVAL '1 month'
    WHEN "billing_cycle" = 'yearly' THEN "expire_after" - INTERVAL '1 year'
    ELSE "expire_before"
END
WHERE "status" = 1
  AND "expire_after" IS NOT NULL
  AND "billing_cycle" IN ('monthly', 'yearly');
