-- ================================================
-- SaaS 微信商城 - 初始数据（PostgreSQL）
-- ================================================

-- ----------------------------
-- 插入套餐数据
-- ----------------------------
INSERT INTO "plans" ("name", "code", "monthly_fee", "yearly_fee", "product_limit", "order_limit", "user_limit", "features", "is_default", "status") VALUES
(
    '基础版',
    'basic',
    299.00,
    2990.00,
    100,
    500,
    1000,
    '["multi_sku", "virtual_product", "coupon", "points", "express_delivery", "city_delivery", "self_pickup"]'::jsonb,
    1,
    1
),
(
    '专业版',
    'professional',
    799.00,
    7990.00,
    2000,
    10000,
    50000,
    '["multi_sku", "virtual_product", "seckill", "group_buy", "distribution", "coupon", "points", "member_level", "express_delivery", "city_delivery", "self_pickup", "custom_domain", "api_access"]'::jsonb,
    0,
    1
),
(
    '旗舰版',
    'enterprise',
    1999.00,
    19990.00,
    0,
    0,
    0,
    '["multi_sku", "virtual_product", "seckill", "group_buy", "distribution", "coupon", "points", "member_level", "express_delivery", "city_delivery", "self_pickup", "custom_domain", "api_access", "white_label", "sms_notification", "private_deployment"]'::jsonb,
    0,
    1
);

-- ----------------------------
-- 插入测试租户（基础版）
-- ----------------------------
INSERT INTO "tenants" (
    "code", "company_name", "contact_name", "contact_phone", "contact_email",
    "plan_id", "plan_expire_at", "brand_name", "brand_theme", "status"
) VALUES (
    'TEST001',
    '测试商户有限公司',
    '张三',
    '13800138000',
    'test@example.com',
    1,
    NOW() + INTERVAL '1 year',
    '测试商城',
    '#1677ff',
    1
);

-- ----------------------------
-- 插入测试租户管理员（商户登录：TEST001 / smokeadmin22 / admin123）
-- ----------------------------
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

-- ----------------------------
-- 插入平台统一商品分类（tenant_id=0，所有租户共享）
-- ----------------------------
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

-- ----------------------------
-- 插入平台物流承运商（快递100 com 编码）
-- ----------------------------
INSERT INTO "shipping_carriers" (
    "tenant_id", "code", "name", "api_provider", "priority",
    "enabled", "audit_status", "submitted_at", "audited_at"
)
SELECT
    0, v.code, v.name, 'kuaidi100', v.priority,
    1, 1, NOW(), NOW()
FROM (VALUES
    ('shunfeng', '顺丰速运', 100),
    ('zhongtong', '中通快递', 95),
    ('yuantong', '圆通速递', 90),
    ('yunda', '韵达快递', 85),
    ('shentong', '申通快递', 80),
    ('ems', 'EMS', 75),
    ('jd', '京东物流', 70),
    ('jtexpress', '极兔速递', 65),
    ('debangwuliu', '德邦快递', 60),
    ('danniao', '丹鸟物流', 55),
    ('annengwuliu', '安能物流', 50),
    ('zhaijisong', '宅急送', 45),
    ('huitongkuaidi', '百世快递', 40),
    ('baishiwuliu', '百世快运', 35),
    ('youshuwuliu', '优速快递', 30),
    ('tiantian', '天天快递', 25),
    ('kuayue', '跨越速运', 20)
) AS v(code, name, priority)
WHERE NOT EXISTS (
    SELECT 1
    FROM "shipping_carriers" sc
    WHERE sc.tenant_id = 0 AND sc.code = v.code
);
