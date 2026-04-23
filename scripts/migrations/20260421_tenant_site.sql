-- 租户站点/品牌配置（合并 custom_domain / white_label / private_deployment）
CREATE TABLE IF NOT EXISTS "tenant_site_configs" (
    "tenant_id"            BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    -- 自定义域名
    "custom_domain"        VARCHAR(128) NOT NULL DEFAULT '',
    "domain_verified"      SMALLINT NOT NULL DEFAULT 0,
    "ssl_status"           VARCHAR(20) NOT NULL DEFAULT 'none', -- none/pending/issued/failed
    -- 白标/品牌
    "brand_name"           VARCHAR(100) NOT NULL DEFAULT '',
    "brand_logo"           VARCHAR(500) NOT NULL DEFAULT '',
    "primary_color"        VARCHAR(32) NOT NULL DEFAULT '#409EFF',
    "hide_platform_brand"  SMALLINT NOT NULL DEFAULT 0,
    "footer_text"          VARCHAR(255) NOT NULL DEFAULT '',
    -- 私有部署
    "deployment_mode"      VARCHAR(16) NOT NULL DEFAULT 'shared', -- shared / private
    "private_endpoint"     VARCHAR(255) NOT NULL DEFAULT '',
    "private_notes"        VARCHAR(500) NOT NULL DEFAULT '',
    -- 商城装修
    "storefront_notice"               VARCHAR(255) NOT NULL DEFAULT '',
    "storefront_hero_title"           VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_hero_subtitle"        VARCHAR(500) NOT NULL DEFAULT '',
    "storefront_search_placeholder"   VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_category_title"       VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_coupon_title"         VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_hot_title"            VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_recommend_title"      VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_quick_entries"        TEXT NOT NULL DEFAULT '[]',
    "storefront_service_cards"        TEXT NOT NULL DEFAULT '[]',
    "storefront_banners"              TEXT NOT NULL DEFAULT '[]',
    "storefront_promo_cards"          TEXT NOT NULL DEFAULT '[]',
    "storefront_member_entries"       TEXT NOT NULL DEFAULT '[]',
    "storefront_home_sections"        TEXT NOT NULL DEFAULT '[]',
    "storefront_profile_sections"     TEXT NOT NULL DEFAULT '[]',
    "storefront_search_keywords"      TEXT NOT NULL DEFAULT '[]',
    "updated_at"           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
