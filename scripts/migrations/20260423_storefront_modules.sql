ALTER TABLE "tenant_site_configs"
    ADD COLUMN IF NOT EXISTS "storefront_banners" TEXT NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS "storefront_promo_cards" TEXT NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS "storefront_member_entries" TEXT NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS "storefront_home_sections" TEXT NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS "storefront_profile_sections" TEXT NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS "storefront_search_keywords" TEXT NOT NULL DEFAULT '[]';
