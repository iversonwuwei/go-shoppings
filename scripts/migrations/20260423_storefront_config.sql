ALTER TABLE "tenant_site_configs"
    ADD COLUMN IF NOT EXISTS "storefront_notice" VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "storefront_hero_title" VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "storefront_hero_subtitle" VARCHAR(500) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "storefront_search_placeholder" VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "storefront_category_title" VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "storefront_coupon_title" VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "storefront_hot_title" VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "storefront_recommend_title" VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS "storefront_quick_entries" TEXT NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS "storefront_service_cards" TEXT NOT NULL DEFAULT '[]';
