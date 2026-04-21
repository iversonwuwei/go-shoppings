-- 2026-04-21 虚拟商品支持
-- 为 products 与 orders 增加 is_virtual 标识。
--   products.is_virtual = 1 表示虚拟商品（无需发货）
--   orders.is_virtual  = 1 表示订单内全部为虚拟商品（支付成功后直接完成）

ALTER TABLE "products" ADD COLUMN IF NOT EXISTS "is_virtual" SMALLINT NOT NULL DEFAULT 0;
ALTER TABLE "orders"   ADD COLUMN IF NOT EXISTS "is_virtual" SMALLINT NOT NULL DEFAULT 0;
