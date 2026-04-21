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
