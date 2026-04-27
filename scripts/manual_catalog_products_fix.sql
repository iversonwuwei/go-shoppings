-- 手动修复小程序“全部商品”无数据的 SQL 汇总
-- 适用场景：本地/演示库使用默认租户 TEST001，小程序分类页需要展示已上架商品。
-- 执行前建议先备份数据库；如果目标租户不是 TEST001，请先把下方所有 TEST001 替换为目标租户 code。

BEGIN;

SET search_path TO public;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM tenants WHERE code = 'TEST001') THEN
        RAISE EXCEPTION '租户 TEST001 不存在，请先创建租户或替换脚本中的租户 code';
    END IF;
END $$;

-- 0) 补齐平台全局设置表。后端订阅/支付配置会读取 platform_settings(id=1)。
CREATE TABLE IF NOT EXISTS "platform_settings" (
    "id"                  BIGSERIAL PRIMARY KEY,
    "platform_name"       TEXT DEFAULT '',
    "platform_logo"       TEXT DEFAULT '',
    "support_phone"       TEXT DEFAULT '',
    "support_email"       TEXT DEFAULT '',
    "wxpay_app_id"        TEXT DEFAULT '',
    "wxpay_mch_id"        TEXT DEFAULT '',
    "wxpay_apiv3_key"     TEXT DEFAULT '',
    "wxpay_cert_serial"   TEXT DEFAULT '',
    "wxpay_notify_url"    TEXT DEFAULT '',
    "updated_at"          TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO "platform_settings" ("id") VALUES (1) ON CONFLICT ("id") DO NOTHING;

-- 0.1) 补齐管理员租户绑定字段，并创建本地演示商户管理员。
ALTER TABLE "admins" ADD COLUMN IF NOT EXISTS "tenant_id" BIGINT NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS "idx_admins_tenant_phone" ON "admins" ("tenant_id", "phone");
CREATE INDEX IF NOT EXISTS "idx_admins_phone" ON "admins" ("phone");

INSERT INTO "admins" (
    "tenant_id", "username", "password", "real_name", "phone", "email", "role", "status"
)
SELECT t.id,
       'smokeadmin22',
    '$2a$10$.laeckOx4u7cZPzuSihBReyLGYuM65e7qhw9I3gshd6GWMs6EXD/C',
       '演示商户管理员',
       '13900000001',
       'smokeadmin22@example.com',
       'admin',
       1
FROM "tenants" t
WHERE t.code = 'TEST001'
ON CONFLICT ("username") DO UPDATE SET
    "tenant_id" = EXCLUDED."tenant_id",
    "password" = EXCLUDED."password",
    "real_name" = EXCLUDED."real_name",
    "phone" = EXCLUDED."phone",
    "email" = EXCLUDED."email",
    "role" = EXCLUDED."role",
    "status" = EXCLUDED."status",
    "updated_at" = CURRENT_TIMESTAMP;

-- 1) 对齐运行时代码：商品分类由平台统一管理，tenant_id = 0 表示全局分类。
ALTER TABLE "product_categories" DROP CONSTRAINT IF EXISTS "product_categories_tenant_id_fkey";
ALTER TABLE "product_categories" ALTER COLUMN "tenant_id" SET DEFAULT 0;
UPDATE "product_categories" SET "tenant_id" = 0 WHERE "tenant_id" <> 0;

-- 2) 补齐平台默认分类，保证 /api/v1/member/categories 有可用分类。
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
    ('其他', 10)
) AS v(name, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM "product_categories" pc WHERE pc.tenant_id = 0 AND pc.name = v.name
);

-- 3) 补齐演示分类，用于下面的 TEST001 演示商品。
INSERT INTO "product_categories" ("tenant_id", "parent_id", "name", "cover_image", "sort", "status")
SELECT 0, 0, v.name, v.cover_image, v.sort, 1
FROM (VALUES
    ('鲜食水果', 'https://images.unsplash.com/photo-1619566636858-adf3ef46400b?auto=format&fit=crop&w=800&q=80', 30),
    ('人气零食', 'https://images.unsplash.com/photo-1585238342024-78d387f4a707?auto=format&fit=crop&w=800&q=80', 20),
    ('生活好物', 'https://images.unsplash.com/photo-1511556532299-8f662fc26c06?auto=format&fit=crop&w=800&q=80', 10)
) AS v(name, cover_image, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM "product_categories" pc WHERE pc.tenant_id = 0 AND pc.name = v.name
);

-- 4) 给 TEST001 补齐可上架展示的演示商品。
WITH tenant AS (
    SELECT id FROM tenants WHERE code = 'TEST001'
),
cats AS (
    SELECT DISTINCT ON (name) id, name
    FROM product_categories
    WHERE tenant_id = 0
    ORDER BY name, sort DESC, id ASC
)
INSERT INTO products (
    tenant_id, category_id, name, subtitle, cover_image, images, description,
    price, stock, stock_warning, has_sku, is_virtual, delivery_type, delivery_fee,
    status, is_recommend, is_hot, sort, sold_count, view_count
)
SELECT tenant.id, cats.id, v.name, v.subtitle, v.cover_image, v.images, v.description,
       v.price, v.stock, 10, 0, v.is_virtual, v.delivery_type, v.delivery_fee,
       1, v.is_recommend, v.is_hot, v.sort, v.sold_count, v.view_count
FROM tenant
CROSS JOIN (VALUES
    (
        '鲜食水果', '云南蓝莓鲜果礼盒', '当季鲜摘，酸甜平衡，适合家庭分享',
        'https://images.unsplash.com/photo-1498557850523-fd3d118b962e?auto=format&fit=crop&w=1200&q=80',
        '["https://images.unsplash.com/photo-1498557850523-fd3d118b962e?auto=format&fit=crop&w=1200&q=80","https://images.unsplash.com/photo-1464965911861-746a04b4bca6?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
        '来自高原果园的蓝莓礼盒，果径饱满，适合作为早餐和下午茶搭配。',
        59.90::numeric, 120, 0, '["express","self_pickup"]'::jsonb, 6.00::numeric, 1, 1, 100, 256, 1280
    ),
    (
        '鲜食水果', '混合水果轻享箱', '苹果、橙子、梨组合搭配，日常囤货首选',
        'https://images.unsplash.com/photo-1610832958506-aa56368176cf?auto=format&fit=crop&w=1200&q=80',
        '["https://images.unsplash.com/photo-1610832958506-aa56368176cf?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
        '精选时令水果组合，一箱满足家庭一周水果需求。',
        89.00::numeric, 80, 0, '["express","city"]'::jsonb, 0.00::numeric, 1, 0, 90, 188, 860
    ),
    (
        '人气零食', '坚果谷物能量包', '低温烘焙，办公零食，补充能量',
        'https://images.unsplash.com/photo-1509440159596-0249088772ff?auto=format&fit=crop&w=1200&q=80',
        '["https://images.unsplash.com/photo-1509440159596-0249088772ff?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
        '核桃、腰果、巴旦木与燕麦混合装，适合办公和差旅。',
        39.90::numeric, 200, 0, '["express","self_pickup"]'::jsonb, 5.00::numeric, 1, 1, 80, 512, 1620
    ),
    (
        '生活好物', '手冲咖啡挂耳组合', '浅烘与中烘双拼，适合居家办公',
        'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=1200&q=80',
        '["https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
        '精选产地豆风味组合，12 包独立包装，适合作为礼物。',
        69.00::numeric, 64, 0, '["express","self_pickup"]'::jsonb, 0.00::numeric, 1, 1, 70, 96, 430
    ),
    (
        '生活好物', '商城会员月卡', '虚拟权益，购买后即时到账',
        'https://images.unsplash.com/photo-1556740749-887f6717d7e4?auto=format&fit=crop&w=1200&q=80',
        '["https://images.unsplash.com/photo-1556740749-887f6717d7e4?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
        '开通后可享专属活动提醒与会员价展示，适合作为虚拟商品演示。',
        19.90::numeric, 9999, 1, '["self_pickup"]'::jsonb, 0.00::numeric, 1, 0, 60, 48, 210
    )
) AS v(category_name, name, subtitle, cover_image, images, description, price, stock, is_virtual, delivery_type, delivery_fee, is_recommend, is_hot, sort, sold_count, view_count)
JOIN cats ON cats.name = v.category_name
WHERE NOT EXISTS (
    SELECT 1 FROM products p WHERE p.tenant_id = tenant.id AND p.name = v.name
);

-- 5) 执行后校验：小程序“全部商品”接口应能看到 on_shelf_products > 0。
SELECT t.id,
       t.code,
       COUNT(p.id) FILTER (WHERE p.status = 1) AS on_shelf_products,
       COUNT(p.id) AS total_products
FROM tenants t
LEFT JOIN products p ON p.tenant_id = t.id
WHERE t.code = 'TEST001'
GROUP BY t.id, t.code;

SELECT tenant_id, COUNT(*) AS category_count
FROM product_categories
GROUP BY tenant_id
ORDER BY tenant_id;

SELECT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'platform_settings'
) AS has_platform_settings;

COMMIT;