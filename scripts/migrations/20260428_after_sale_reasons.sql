-- 2026-04-28 售后原因配置
-- 目标：平台统一维护售后原因，小程序申请售后时以下拉方式选择。

CREATE TABLE IF NOT EXISTS "after_sale_reasons" (
    "id" BIGSERIAL PRIMARY KEY,
    "code" VARCHAR(40) NOT NULL UNIQUE,
    "label" VARCHAR(80) NOT NULL,
    "type" VARCHAR(20) NOT NULL DEFAULT 'all',
    "sort_order" INT NOT NULL DEFAULT 0,
    "enabled" SMALLINT NOT NULL DEFAULT 1,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS "idx_after_sale_reasons_type" ON "after_sale_reasons" ("type");
CREATE INDEX IF NOT EXISTS "idx_after_sale_reasons_enabled" ON "after_sale_reasons" ("enabled");
CREATE INDEX IF NOT EXISTS "idx_after_sale_reasons_sort" ON "after_sale_reasons" ("sort_order");

INSERT INTO "after_sale_reasons" ("code", "label", "type", "sort_order", "enabled") VALUES
    ('no_longer_needed', '不想要了', 'refund', 10, 1),
    ('wrong_or_duplicate', '拍错/多拍', 'refund', 20, 1),
    ('damaged_goods', '商品破损', 'return_refund', 30, 1),
    ('not_as_described', '商品与描述不符', 'return_refund', 40, 1),
    ('missing_items', '少件/漏发', 'return_refund', 50, 1),
    ('quality_issue', '质量问题', 'return_refund', 60, 1),
    ('negotiated_refund', '协商一致退款', 'all', 70, 1)
ON CONFLICT ("code") DO NOTHING;

-- Verification SQL:
-- SELECT code, label, type, enabled
-- FROM after_sale_reasons
-- ORDER BY sort_order ASC, id ASC;
