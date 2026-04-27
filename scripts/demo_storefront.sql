-- 演示商城数据：测试租户 TEST001
-- 可重复执行；若数据已存在则跳过

INSERT INTO product_categories (tenant_id, parent_id, name, cover_image, sort, status)
SELECT 0, 0, v.name, v.cover_image, v.sort, 1
FROM (
    VALUES
        ('鲜食水果', 'https://images.unsplash.com/photo-1619566636858-adf3ef46400b?auto=format&fit=crop&w=800&q=80', 10),
        ('人气零食', 'https://images.unsplash.com/photo-1585238342024-78d387f4a707?auto=format&fit=crop&w=800&q=80', 20),
        ('生活好物', 'https://images.unsplash.com/photo-1511556532299-8f662fc26c06?auto=format&fit=crop&w=800&q=80', 30)
) AS v(name, cover_image, sort)
WHERE NOT EXISTS (
    SELECT 1 FROM product_categories c WHERE c.tenant_id = 0 AND c.name = v.name
);

WITH tenant AS (
    SELECT id FROM tenants WHERE code = 'TEST001'
)
INSERT INTO coupons (
    tenant_id, name, type, threshold_amount, discount_value, max_discount,
    total_count, remain_count, per_limit, receive_start_at, receive_end_at,
    valid_start_at, valid_end_at, valid_days, applicable_type, applicable_ids,
    member_levels, status
)
SELECT tenant.id, v.name, v.type, v.threshold_amount, v.discount_value, v.max_discount,
       v.total_count, v.remain_count, v.per_limit, NOW() - INTERVAL '1 day', NOW() + INTERVAL '30 day',
       NOW() - INTERVAL '1 day', NOW() + INTERVAL '30 day', 30, 'all', '[]'::jsonb,
       '[]'::jsonb, 1
FROM tenant
CROSS JOIN (
    VALUES
        ('新人满99减10', 'cash', 99.00::numeric, 10.00::numeric, NULL::numeric, 500, 500, 1),
        ('周末88折券', 'discount', 199.00::numeric, 8.80::numeric, 50.00::numeric, 300, 300, 1)
) AS v(name, type, threshold_amount, discount_value, max_discount, total_count, remain_count, per_limit)
WHERE NOT EXISTS (
    SELECT 1 FROM coupons c WHERE c.tenant_id = tenant.id AND c.name = v.name
);

WITH tenant AS (
    SELECT id FROM tenants WHERE code = 'TEST001'
),
cats AS (
    SELECT id, name FROM product_categories WHERE tenant_id = 0
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
CROSS JOIN (
    VALUES
        (
            '鲜食水果', '云南蓝莓鲜果礼盒', '当季鲜摘，酸甜平衡，适合家庭分享',
            'https://images.unsplash.com/photo-1498557850523-fd3d118b962e?auto=format&fit=crop&w=1200&q=80',
            '["https://images.unsplash.com/photo-1498557850523-fd3d118b962e?auto=format&fit=crop&w=1200&q=80","https://images.unsplash.com/photo-1464965911861-746a04b4bca6?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
            '来自高原果园的蓝莓礼盒，果径饱满，适合作为早餐和下午茶搭配。', 59.90::numeric, 120, 0, '["express","self_pickup"]'::jsonb, 6.00::numeric, 1, 1, 100, 256, 1280
        ),
        (
            '鲜食水果', '混合水果轻享箱', '苹果、橙子、梨组合搭配，日常囤货首选',
            'https://images.unsplash.com/photo-1610832958506-aa56368176cf?auto=format&fit=crop&w=1200&q=80',
            '["https://images.unsplash.com/photo-1610832958506-aa56368176cf?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
            '精选时令水果组合，一箱满足家庭一周水果需求。', 89.00::numeric, 80, 0, '["express","city"]'::jsonb, 0.00::numeric, 1, 0, 90, 188, 860
        ),
        (
            '人气零食', '坚果谷物能量包', '低温烘焙，办公零食，补充能量',
            'https://images.unsplash.com/photo-1509440159596-0249088772ff?auto=format&fit=crop&w=1200&q=80',
            '["https://images.unsplash.com/photo-1509440159596-0249088772ff?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
            '核桃、腰果、巴旦木与燕麦混合装，适合办公和差旅。', 39.90::numeric, 200, 0, '["express","self_pickup"]'::jsonb, 5.00::numeric, 1, 1, 80, 512, 1620
        ),
        (
            '生活好物', '手冲咖啡挂耳组合', '浅烘与中烘双拼，适合居家办公',
            'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=1200&q=80',
            '["https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
            '精选产地豆风味组合，12 包独立包装，适合作为礼物。', 69.00::numeric, 64, 0, '["express","self_pickup"]'::jsonb, 0.00::numeric, 1, 1, 70, 96, 430
        ),
        (
            '生活好物', '商城会员月卡', '虚拟权益，购买后即时到账',
            'https://images.unsplash.com/photo-1556740749-887f6717d7e4?auto=format&fit=crop&w=1200&q=80',
            '["https://images.unsplash.com/photo-1556740749-887f6717d7e4?auto=format&fit=crop&w=1200&q=80"]'::jsonb,
            '开通后可享专属活动提醒与会员价展示，适合作为虚拟商品演示。', 19.90::numeric, 9999, 1, '["self_pickup"]'::jsonb, 0.00::numeric, 1, 0, 60, 48, 210
        )
) AS v(category_name, name, subtitle, cover_image, images, description, price, stock, is_virtual, delivery_type, delivery_fee, is_recommend, is_hot, sort, sold_count, view_count)
JOIN cats ON cats.name = v.category_name
WHERE NOT EXISTS (
    SELECT 1 FROM products p WHERE p.tenant_id = tenant.id AND p.name = v.name
);

WITH tenant AS (
    SELECT id FROM tenants WHERE code = 'TEST001'
)
INSERT INTO members (
    tenant_id, openid, unionid, session_key, nickname, avatar, gender, phone,
    points, growth_value, status
)
SELECT tenant.id, 'dev-demo-member', 'dev-demo-member', 'dev-session',
       '演示会员', 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?auto=format&fit=crop&w=300&q=80',
       1, '13800000001', 128, 36, 1
FROM tenant
WHERE NOT EXISTS (
    SELECT 1 FROM members m WHERE m.tenant_id = tenant.id AND m.phone = '13800000001'
);

WITH tenant AS (
    SELECT id FROM tenants WHERE code = 'TEST001'
),
member_row AS (
    SELECT id, tenant_id FROM members WHERE tenant_id = (SELECT id FROM tenant) AND phone = '13800000001'
)
INSERT INTO member_addresses (
    tenant_id, member_id, receiver_name, receiver_phone, province, city,
    district, address, postcode, is_default
)
SELECT member_row.tenant_id, member_row.id, '演示会员', '13800000001', '上海市', '上海市',
       '浦东新区', '张江高科技园区测试路 88 号 8 楼', '200120', 1
FROM member_row
WHERE NOT EXISTS (
    SELECT 1 FROM member_addresses a WHERE a.member_id = member_row.id AND a.is_default = 1
);

WITH tenant AS (
    SELECT id FROM tenants WHERE code = 'TEST001'
),
member_row AS (
    SELECT id, tenant_id FROM members WHERE tenant_id = (SELECT id FROM tenant) AND phone = '13800000001'
)
INSERT INTO points_logs (
    tenant_id, member_id, change_type, change_value, balance_before, balance_after, source_desc, remark
)
SELECT member_row.tenant_id, member_row.id, 'system_grant', 128, 0, 128, '商城演示初始化', '用于前端积分页面演示'
FROM member_row
WHERE NOT EXISTS (
    SELECT 1 FROM points_logs l WHERE l.member_id = member_row.id AND l.source_desc = '商城演示初始化'
);
