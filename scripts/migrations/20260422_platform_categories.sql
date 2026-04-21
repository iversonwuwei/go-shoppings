-- 将商品分类改为平台统一管理（tenant_id = 0 表示平台全局）
-- 0) 去掉 product_categories.tenant_id 对 tenants 的外键（tenant_id=0 表示平台）
ALTER TABLE "product_categories" DROP CONSTRAINT IF EXISTS "product_categories_tenant_id_fkey";

-- 1) 历史数据：把所有已有分类统一收编到平台（如不希望保留旧租户数据，可改为 DELETE）
UPDATE "product_categories" SET "tenant_id" = 0 WHERE "tenant_id" <> 0;

-- 2) 初始几个默认分类（幂等：存在同名则跳过）
INSERT INTO "product_categories" ("tenant_id", "parent_id", "name", "sort", "status")
SELECT 0, 0, v.name, v.sort, 1
FROM (VALUES
    ('食品饮料', 100),
    ('美妆护肤', 90),
    ('数码家电', 80),
    ('服饰鞋包', 70),
    ('家居日用', 60),
    ('母婴玩具', 50),
    ('图书文娱', 40),
    ('运动户外', 30),
    ('虚拟服务', 20),
    ('其他',   10)
) AS v(name, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM "product_categories" pc WHERE pc.tenant_id = 0 AND pc.name = v.name
);
