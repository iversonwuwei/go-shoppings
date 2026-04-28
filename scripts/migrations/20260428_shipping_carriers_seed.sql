-- 2026-04-28 新增：平台物流承运商初始化数据（快递100 com 编码）
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
