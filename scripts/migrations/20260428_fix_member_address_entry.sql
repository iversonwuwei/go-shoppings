-- 修复历史店铺配置：会员中心“收货地址”快捷入口不应指回会员中心自身。
-- tenant_site_configs.storefront_member_entries 以 JSON 文本保存，保持原数组顺序仅修正目标 path。

UPDATE "tenant_site_configs" AS cfg
SET "storefront_member_entries" = fixed.entries::text
FROM (
    SELECT
        source."id",
        COALESCE(
            jsonb_agg(
                CASE
                    WHEN item.entry ->> 'title' = '收货地址'
                         AND item.entry ->> 'path' = '/profile'
                    THEN jsonb_set(item.entry, '{path}', to_jsonb('/addresses'::text))
                    ELSE item.entry
                END
                ORDER BY item.ordinality
            ),
            '[]'::jsonb
        ) AS entries
    FROM "tenant_site_configs" AS source
    CROSS JOIN LATERAL jsonb_array_elements(
        COALESCE(NULLIF(source."storefront_member_entries", ''), '[]')::jsonb
    ) WITH ORDINALITY AS item(entry, ordinality)
    WHERE EXISTS (
        SELECT 1
        FROM jsonb_array_elements(
            COALESCE(NULLIF(source."storefront_member_entries", ''), '[]')::jsonb
        ) AS existing(entry)
        WHERE existing.entry ->> 'title' = '收货地址'
          AND existing.entry ->> 'path' = '/profile'
    )
    GROUP BY source."id"
) AS fixed
WHERE cfg."id" = fixed."id";