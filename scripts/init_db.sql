-- ================================================
-- SaaS 微信商城 - 数据库初始化脚本 (PostgreSQL)
-- 版本：v1.0
-- ================================================

-- 数据库由 docker-compose 创建，这里只建表

SET NAMES 'UTF8';
SET search_path TO public;

-- ----------------------------
-- 1. plans（套餐定义表）
-- ----------------------------
DROP TABLE IF EXISTS "tenant_plan_logs" CASCADE;
DROP TABLE IF EXISTS "order_items" CASCADE;
DROP TABLE IF EXISTS "order_logs" CASCADE;
DROP TABLE IF EXISTS "orders" CASCADE;
DROP TABLE IF EXISTS "member_coupons" CASCADE;
DROP TABLE IF EXISTS "coupons" CASCADE;
DROP TABLE IF EXISTS "distribution_commissions" CASCADE;
DROP TABLE IF EXISTS "distribution_relations" CASCADE;
DROP TABLE IF EXISTS "seckill_products" CASCADE;
DROP TABLE IF EXISTS "seckill_activities" CASCADE;
DROP TABLE IF EXISTS "group_buy_orders" CASCADE;
DROP TABLE IF EXISTS "group_buys" CASCADE;
DROP TABLE IF EXISTS "points_logs" CASCADE;
DROP TABLE IF EXISTS "member_addresses" CASCADE;
DROP TABLE IF EXISTS "members" CASCADE;
DROP TABLE IF EXISTS "member_levels" CASCADE;
DROP TABLE IF EXISTS "product_skus" CASCADE;
DROP TABLE IF EXISTS "product_attribute_values" CASCADE;
DROP TABLE IF EXISTS "product_attributes" CASCADE;
DROP TABLE IF EXISTS "products" CASCADE;
DROP TABLE IF EXISTS "product_categories" CASCADE;
DROP TABLE IF EXISTS "payments" CASCADE;
DROP TABLE IF EXISTS "admin_action_logs" CASCADE;
DROP TABLE IF EXISTS "uploads" CASCADE;
DROP TABLE IF EXISTS "tenant_site_configs" CASCADE;
DROP TABLE IF EXISTS "admins" CASCADE;
DROP TABLE IF EXISTS "tenants" CASCADE;
DROP TABLE IF EXISTS "plans" CASCADE;

CREATE TABLE "plans" (
    "id"              BIGSERIAL PRIMARY KEY,
    "name"            VARCHAR(50) NOT NULL,
    "code"            VARCHAR(30) NOT NULL UNIQUE,
    "monthly_fee"     NUMERIC(10,2) NOT NULL DEFAULT 0,
    "yearly_fee"      NUMERIC(10,2) NOT NULL DEFAULT 0,
    "product_limit"   INT NOT NULL DEFAULT 0,
    "order_limit"     INT NOT NULL DEFAULT 0,
    "user_limit"      INT NOT NULL DEFAULT 0,
    "features"        JSONB NOT NULL DEFAULT '[]',
    "is_default"      SMALLINT NOT NULL DEFAULT 0,
    "status"          SMALLINT NOT NULL DEFAULT 1,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_plans_code" ON "plans" ("code");
CREATE INDEX "idx_plans_status" ON "plans" ("status");

-- ----------------------------
-- 2. tenants（租户表）
-- ----------------------------
CREATE TABLE "tenants" (
    "id"                  BIGSERIAL PRIMARY KEY,
    "code"                VARCHAR(30) NOT NULL UNIQUE,
    "company_name"        VARCHAR(100) NOT NULL,
    "contact_name"        VARCHAR(50) NOT NULL,
    "contact_phone"       VARCHAR(20) NOT NULL,
    "contact_email"       VARCHAR(100) NOT NULL,
    "wechat_appid"        VARCHAR(50),
    "wechat_secret"       VARCHAR(255),
    "wechat_mchid"       VARCHAR(30),
    "wechat_apiv3_key"   VARCHAR(255),
    "wechat_cert_serial"  VARCHAR(100),
    "plan_id"             BIGINT NOT NULL REFERENCES "plans"("id"),
    "plan_expire_at"      TIMESTAMP NOT NULL,
    "billing_cycle"       VARCHAR(10) NOT NULL DEFAULT 'yearly',
    "brand_name"          VARCHAR(50),
    "brand_logo"          VARCHAR(255),
    "brand_theme"         VARCHAR(20) DEFAULT '#1989fa',
    "brand_domain"       VARCHAR(100),
    "status"              SMALLINT NOT NULL DEFAULT 0,
    "reject_reason"       VARCHAR(255),
    "extra_features"      JSONB NOT NULL DEFAULT '[]'::jsonb,
    "created_at"          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_tenants_plan_expire" ON "tenants" ("plan_expire_at");
CREATE INDEX "idx_tenants_status" ON "tenants" ("status");

-- ----------------------------
-- 3. admins（管理员表）
-- ----------------------------
CREATE TABLE "admins" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL DEFAULT 0,
    "username"        VARCHAR(50) NOT NULL UNIQUE,
    "password"        VARCHAR(255) NOT NULL,
    "real_name"       VARCHAR(50),
    "phone"           VARCHAR(20),
    "email"           VARCHAR(100),
    "role"            VARCHAR(20) NOT NULL DEFAULT 'admin',
    "status"          SMALLINT NOT NULL DEFAULT 1,
    "last_login_at"   TIMESTAMP,
    "last_login_ip"   VARCHAR(50),
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_admins_username" ON "admins" ("username");
CREATE INDEX "idx_admins_tenant_phone" ON "admins" ("tenant_id", "phone");
CREATE INDEX "idx_admins_phone" ON "admins" ("phone");

-- 超级管理员: admin / admin123
INSERT INTO "admins" ("tenant_id", "username", "password", "real_name", "role", "status")
VALUES (0, 'admin', '$2a$10$.laeckOx4u7cZPzuSihBReyLGYuM65e7qhw9I3gshd6GWMs6EXD/C', '超级管理员', 'super', 1);

-- ----------------------------
-- 4. tenant_plan_logs（套餐变更记录）
-- ----------------------------
CREATE TABLE "tenant_plan_logs" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "old_plan_id"     BIGINT REFERENCES "plans"("id"),
    "new_plan_id"     BIGINT NOT NULL REFERENCES "plans"("id"),
    "change_type"     VARCHAR(20) NOT NULL,
    "effective_at"    TIMESTAMP NOT NULL,
    "expire_at"       TIMESTAMP NOT NULL,
    "amount"          NUMERIC(10,2) NOT NULL,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_tenant_plan_logs_tenant" ON "tenant_plan_logs" ("tenant_id");

-- ----------------------------
-- 5. product_categories（商品分类表）
-- ----------------------------
CREATE TABLE "product_categories" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL DEFAULT 0,
    "parent_id"       BIGINT NOT NULL DEFAULT 0,
    "name"            VARCHAR(50) NOT NULL,
    "icon"            VARCHAR(255),
    "cover_image"     VARCHAR(255),
    "sort"            INT NOT NULL DEFAULT 0,
    "status"          SMALLINT NOT NULL DEFAULT 1,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_product_categories_tenant" ON "product_categories" ("tenant_id");
CREATE INDEX "idx_product_categories_parent" ON "product_categories" ("parent_id");

CREATE TABLE "tenant_category_assets" (
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "category_id"     BIGINT NOT NULL REFERENCES "product_categories"("id"),
    "icon"            VARCHAR(255),
    "cover_image"     VARCHAR(255),
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("tenant_id", "category_id")
);
CREATE INDEX "idx_tenant_category_assets_category" ON "tenant_category_assets" ("category_id");

-- ----------------------------
-- 6. member_levels（会员等级表）
-- ----------------------------
CREATE TABLE "member_levels" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "name"            VARCHAR(30) NOT NULL,
    "icon"            VARCHAR(255),
    "color"           VARCHAR(20),
    "min_growth"      INT NOT NULL DEFAULT 0,
    "discount_rate"   NUMERIC(4,2) NOT NULL DEFAULT 100,
    "points_mult"     NUMERIC(3,2) NOT NULL DEFAULT 1,
    "sort"            INT NOT NULL DEFAULT 0,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_member_levels_tenant" ON "member_levels" ("tenant_id");
CREATE INDEX "idx_member_levels_growth" ON "member_levels" ("min_growth");

-- ----------------------------
-- 7. members（会员表）
-- ----------------------------
CREATE TABLE "members" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "openid"          VARCHAR(50),
    "unionid"         VARCHAR(50),
    "session_key"     VARCHAR(255),
    "nickname"        VARCHAR(50),
    "avatar"          VARCHAR(255),
    "gender"          SMALLINT,
    "birthday"        DATE,
    "phone"           VARCHAR(20),
    "level_id"        BIGINT REFERENCES "member_levels"("id"),
    "level_expire_at" TIMESTAMP,
    "points"          INT NOT NULL DEFAULT 0,
    "growth_value"    INT NOT NULL DEFAULT 0,
    "parent_id"       BIGINT,
    "level1_count"    INT NOT NULL DEFAULT 0,
    "level2_count"    INT NOT NULL DEFAULT 0,
    "status"          SMALLINT NOT NULL DEFAULT 1,
    "last_login_at"   TIMESTAMP,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"      TIMESTAMP,
    UNIQUE ("tenant_id", "openid")
);
CREATE INDEX "idx_members_tenant" ON "members" ("tenant_id");
CREATE INDEX "idx_members_parent" ON "members" ("parent_id");
CREATE INDEX "idx_members_level" ON "members" ("level_id");
CREATE UNIQUE INDEX "uk_members_tenant_phone_present" ON "members" ("tenant_id", "phone") WHERE "phone" IS NOT NULL AND "phone" <> '';

-- ----------------------------
-- 8. member_addresses（收货地址表）
-- ----------------------------
CREATE TABLE "member_addresses" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "member_id"       BIGINT NOT NULL REFERENCES "members"("id"),
    "receiver_name"   VARCHAR(50) NOT NULL,
    "receiver_phone"  VARCHAR(20) NOT NULL,
    "province"        VARCHAR(20) NOT NULL,
    "city"            VARCHAR(20) NOT NULL,
    "district"        VARCHAR(20) NOT NULL,
    "address"         VARCHAR(255) NOT NULL,
    "postcode"        VARCHAR(10),
    "is_default"      SMALLINT NOT NULL DEFAULT 0,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_member_addresses_tenant" ON "member_addresses" ("tenant_id");
CREATE INDEX "idx_member_addresses_member" ON "member_addresses" ("member_id");

-- ----------------------------
-- 9. points_logs（积分变动记录）
-- ----------------------------
CREATE TABLE "points_logs" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "member_id"       BIGINT NOT NULL REFERENCES "members"("id"),
    "change_type"     VARCHAR(20) NOT NULL,
    "change_value"    INT NOT NULL,
    "balance_before"  INT NOT NULL,
    "balance_after"   INT NOT NULL,
    "source_id"       BIGINT,
    "source_desc"     VARCHAR(200),
    "remark"          VARCHAR(500),
    "operator_id"     BIGINT,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_points_logs_tenant" ON "points_logs" ("tenant_id");
CREATE INDEX "idx_points_logs_member" ON "points_logs" ("member_id");

-- ----------------------------
-- 10. products（商品主表）
-- ----------------------------
CREATE TABLE "products" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "category_id"     BIGINT REFERENCES "product_categories"("id"),
    "name"            VARCHAR(200) NOT NULL,
    "subtitle"        VARCHAR(500),
    "cover_image"     VARCHAR(255) NOT NULL,
    "images"          JSONB NOT NULL DEFAULT '[]',
    "video_url"       VARCHAR(500),
    "description"     TEXT,
    "price"           NUMERIC(10,2) NOT NULL DEFAULT 0,
    "cost_price"      NUMERIC(10,2),
    "stock"           INT NOT NULL DEFAULT 0,
    "stock_warning"   INT NOT NULL DEFAULT 10,
    "has_sku"         SMALLINT NOT NULL DEFAULT 0,
    "delivery_type"   JSONB NOT NULL DEFAULT '[]',
    "delivery_fee"    NUMERIC(10,2) NOT NULL DEFAULT 0,
    "status"          SMALLINT NOT NULL DEFAULT 1,
    "is_recommend"    SMALLINT NOT NULL DEFAULT 0,
    "is_hot"          SMALLINT NOT NULL DEFAULT 0,
    "seo_title"       VARCHAR(200),
    "seo_keywords"    VARCHAR(500),
    "seo_description" VARCHAR(500),
    "sort"            INT NOT NULL DEFAULT 0,
    "sold_count"      INT NOT NULL DEFAULT 0,
    "view_count"      INT NOT NULL DEFAULT 0,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"      TIMESTAMP
);
CREATE INDEX "idx_products_tenant" ON "products" ("tenant_id");
CREATE INDEX "idx_products_category" ON "products" ("category_id");
CREATE INDEX "idx_products_status" ON "products" ("status");
CREATE INDEX "idx_products_sold" ON "products" ("sold_count" DESC);

-- ----------------------------
-- 11. product_attributes（规格属性表）
-- ----------------------------
CREATE TABLE "product_attributes" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "product_id"      BIGINT NOT NULL REFERENCES "products"("id"),
    "name"            VARCHAR(30) NOT NULL,
    "sort"            INT NOT NULL DEFAULT 0,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_product_attributes_tenant" ON "product_attributes" ("tenant_id");

-- ----------------------------
-- 12. product_attribute_values（规格值表）
-- ----------------------------
CREATE TABLE "product_attribute_values" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "attribute_id"    BIGINT NOT NULL REFERENCES "product_attributes"("id"),
    "value"           VARCHAR(50) NOT NULL,
    "sort"            INT NOT NULL DEFAULT 0,
    "image"           VARCHAR(255),
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_product_attribute_values_tenant" ON "product_attribute_values" ("tenant_id");

-- ----------------------------
-- 13. product_skus（SKU表）
-- ----------------------------
CREATE TABLE "product_skus" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "product_id"      BIGINT NOT NULL REFERENCES "products"("id"),
    "sku_code"        VARCHAR(50) NOT NULL,
    "attributes"       JSONB NOT NULL DEFAULT '{}',
    "price"           NUMERIC(10,2) NOT NULL,
    "cost_price"      NUMERIC(10,2),
    "stock"           INT NOT NULL DEFAULT 0,
    "image"           VARCHAR(255),
    "weight"          NUMERIC(10,2),
    "volume"          NUMERIC(10,2),
    "status"          SMALLINT NOT NULL DEFAULT 1,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE ("tenant_id", "sku_code")
);
CREATE INDEX "idx_product_skus_product" ON "product_skus" ("product_id");

-- ----------------------------
-- 14. orders（订单主表）
-- ----------------------------
CREATE TABLE "orders" (
    "id"                  BIGSERIAL PRIMARY KEY,
    "tenant_id"           BIGINT NOT NULL REFERENCES "tenants"("id"),
    "order_no"            VARCHAR(32) NOT NULL UNIQUE,
    "member_id"           BIGINT NOT NULL REFERENCES "members"("id"),
    "total_amount"        NUMERIC(10,2) NOT NULL,
    "delivery_fee"       NUMERIC(10,2) NOT NULL DEFAULT 0,
    "discount_amount"     NUMERIC(10,2) NOT NULL DEFAULT 0,
    "coupon_id"           BIGINT,
    "points_discount"     NUMERIC(10,2) NOT NULL DEFAULT 0,
    "actual_amount"       NUMERIC(10,2) NOT NULL,
    "cost_amount"         NUMERIC(10,2),
    "status"              VARCHAR(20) NOT NULL DEFAULT 'pending_pay',
    "receiver_name"       VARCHAR(50),
    "receiver_phone"      VARCHAR(20),
    "receiver_province"   VARCHAR(20),
    "receiver_city"       VARCHAR(20),
    "receiver_district"   VARCHAR(20),
    "receiver_address"   VARCHAR(255),
    "receiver_postcode"  VARCHAR(10),
    "delivery_type"       VARCHAR(20) NOT NULL,
    "express_company"     VARCHAR(30),
    "express_no"         VARCHAR(50),
    "self_pickup_code"   VARCHAR(20),
    "self_pickup_address" VARCHAR(255),
    "buyer_remark"         VARCHAR(500),
    "distribution_status" VARCHAR(20) DEFAULT 'pending',
    "paid_at"             TIMESTAMP,
    "shipped_at"          TIMESTAMP,
    "delivered_at"       TIMESTAMP,
    "completed_at"       TIMESTAMP,
    "cancelled_at"       TIMESTAMP,
    "expired_at"         TIMESTAMP,
    "created_at"          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"          TIMESTAMP
);
CREATE INDEX "idx_orders_tenant" ON "orders" ("tenant_id");
CREATE INDEX "idx_orders_member" ON "orders" ("member_id");
CREATE INDEX "idx_orders_status" ON "orders" ("status");
CREATE INDEX "idx_orders_order_no" ON "orders" ("order_no");
CREATE INDEX "idx_orders_paid_at" ON "orders" ("paid_at");
CREATE INDEX "idx_orders_created" ON "orders" ("created_at");

-- ----------------------------
-- 15. order_items（订单商品明细）
-- ----------------------------
CREATE TABLE "order_items" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "order_id"        BIGINT NOT NULL REFERENCES "orders"("id"),
    "product_id"      BIGINT NOT NULL,
    "sku_id"          BIGINT,
    "product_name"    VARCHAR(200) NOT NULL,
    "sku_desc"        VARCHAR(200),
    "cover_image"     VARCHAR(255) NOT NULL,
    "price"           NUMERIC(10,2) NOT NULL,
    "quantity"        INT NOT NULL DEFAULT 1,
    "item_total"      NUMERIC(10,2) NOT NULL,
    "refund_status"   VARCHAR(20) DEFAULT 'none',
    "refund_amount"   NUMERIC(10,2),
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_order_items_tenant" ON "order_items" ("tenant_id");
CREATE INDEX "idx_order_items_order" ON "order_items" ("order_id");

-- ----------------------------
-- 16. order_logs（订单操作日志）
-- ----------------------------
CREATE TABLE "order_logs" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "order_id"        BIGINT NOT NULL REFERENCES "orders"("id"),
    "operator_type"   VARCHAR(20) NOT NULL,
    "operator_id"     BIGINT,
    "action"          VARCHAR(50) NOT NULL,
    "before_status"   VARCHAR(20),
    "after_status"    VARCHAR(20),
    "remark"          VARCHAR(500),
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_order_logs_tenant" ON "order_logs" ("tenant_id");
CREATE INDEX "idx_order_logs_order" ON "order_logs" ("order_id");

CREATE TABLE "order_messages" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "order_id"        BIGINT NOT NULL REFERENCES "orders"("id"),
    "order_no"        VARCHAR(32) NOT NULL,
    "event_type"      VARCHAR(40) NOT NULL,
    "title"           VARCHAR(120) NOT NULL,
    "content"         VARCHAR(500) NOT NULL,
    "status"          VARCHAR(20) NOT NULL DEFAULT 'unread',
    "read_at"         TIMESTAMP,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_order_messages_tenant" ON "order_messages" ("tenant_id");
CREATE INDEX "idx_order_messages_order" ON "order_messages" ("order_id");
CREATE INDEX "idx_order_messages_status" ON "order_messages" ("status");

-- ----------------------------
-- 17. coupons（优惠券表）
-- ----------------------------
CREATE TABLE "coupons" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "name"            VARCHAR(50) NOT NULL,
    "type"            VARCHAR(20) NOT NULL,
    "threshold_amount" NUMERIC(10,2),
    "discount_value"   NUMERIC(10,2) NOT NULL,
    "max_discount"     NUMERIC(10,2),
    "receive_limit_type" VARCHAR(20) NOT NULL DEFAULT 'limited',
    "total_count"      INT NOT NULL DEFAULT 0,
    "remain_count"     INT NOT NULL DEFAULT 0,
    "per_limit"        INT NOT NULL DEFAULT 1,
    "use_limit"        INT NOT NULL DEFAULT 1,
    "receive_start_at" TIMESTAMP,
    "receive_end_at"   TIMESTAMP,
    "valid_start_at"   TIMESTAMP,
    "valid_end_at"     TIMESTAMP,
    "valid_days"       INT,
    "applicable_type"  VARCHAR(20) NOT NULL DEFAULT 'all',
    "applicable_ids"   JSONB,
    "member_levels"     JSONB,
    "status"          SMALLINT NOT NULL DEFAULT 1,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_coupons_tenant" ON "coupons" ("tenant_id");
CREATE INDEX "idx_coupons_status" ON "coupons" ("status");

-- ----------------------------
-- 18. member_coupons（会员优惠券记录）
-- ----------------------------
CREATE TABLE "member_coupons" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "member_id"       BIGINT NOT NULL REFERENCES "members"("id"),
    "coupon_id"       BIGINT NOT NULL REFERENCES "coupons"("id"),
    "coupon_name"     VARCHAR(50) NOT NULL,
    "coupon_type"     VARCHAR(20) NOT NULL,
    "threshold_amount" NUMERIC(10,2),
    "discount_value"   NUMERIC(10,2) NOT NULL,
    "max_discount"     NUMERIC(10,2),
    "use_limit"        INT NOT NULL DEFAULT 1,
    "received_at"     TIMESTAMP NOT NULL,
    "valid_start_at"  TIMESTAMP NOT NULL,
    "valid_end_at"    TIMESTAMP NOT NULL,
    "used_at"         TIMESTAMP,
    "used_order_id"   BIGINT,
    "status"          VARCHAR(20) NOT NULL DEFAULT 'unused',
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE ("member_id", "coupon_id", "received_at")
);
CREATE INDEX "idx_member_coupons_tenant" ON "member_coupons" ("tenant_id");
CREATE INDEX "idx_member_coupons_member" ON "member_coupons" ("member_id");
CREATE INDEX "idx_member_coupons_status" ON "member_coupons" ("status");
CREATE INDEX "idx_member_coupons_valid_end" ON "member_coupons" ("valid_end_at");

-- ----------------------------
-- 19. group_buys（拼团活动表）
-- ----------------------------
CREATE TABLE "group_buys" (
    "id"                  BIGSERIAL PRIMARY KEY,
    "tenant_id"           BIGINT NOT NULL REFERENCES "tenants"("id"),
    "product_id"          BIGINT NOT NULL,
    "sku_id"              BIGINT,
    "group_price"         NUMERIC(10,2) NOT NULL,
    "needed_people"        INT NOT NULL DEFAULT 2,
    "group_valid_hours"    INT NOT NULL DEFAULT 24,
    "total_stock"          INT NOT NULL,
    "per_person_limit"     INT NOT NULL DEFAULT 1,
    "start_at"             TIMESTAMP NOT NULL,
    "end_at"               TIMESTAMP NOT NULL,
    "status"               SMALLINT NOT NULL DEFAULT 1,
    "created_at"           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_group_buys_tenant" ON "group_buys" ("tenant_id");

-- ----------------------------
-- 20. group_buy_orders（拼团记录表）
-- ----------------------------
CREATE TABLE "group_buy_orders" (
    "id"                  BIGSERIAL PRIMARY KEY,
    "tenant_id"           BIGINT NOT NULL REFERENCES "tenants"("id"),
    "group_buy_id"        BIGINT NOT NULL,
    "order_id"            BIGINT NOT NULL,
    "leader_id"           BIGINT NOT NULL,
    "needed_people"        INT NOT NULL,
    "joined_people"        INT NOT NULL DEFAULT 1,
    "status"              VARCHAR(20) NOT NULL DEFAULT 'ongoing',
    "expire_at"           TIMESTAMP NOT NULL,
    "success_at"          TIMESTAMP,
    "created_at"          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_group_buy_orders_tenant" ON "group_buy_orders" ("tenant_id");
CREATE INDEX "idx_group_buy_orders_leader" ON "group_buy_orders" ("leader_id");
CREATE INDEX "idx_group_buy_orders_status" ON "group_buy_orders" ("status");

-- ----------------------------
-- 21. seckill_activities（秒杀活动表）
-- ----------------------------
CREATE TABLE "seckill_activities" (
    "id"                  BIGSERIAL PRIMARY KEY,
    "tenant_id"           BIGINT NOT NULL REFERENCES "tenants"("id"),
    "name"                VARCHAR(100) NOT NULL,
    "start_at"            TIMESTAMP NOT NULL,
    "end_at"              TIMESTAMP NOT NULL,
    "per_limit"           INT NOT NULL DEFAULT 1,
    "total_stock"         INT NOT NULL,
    "status"              SMALLINT NOT NULL DEFAULT 1,
    "created_at"          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_seckill_activities_tenant" ON "seckill_activities" ("tenant_id");
CREATE INDEX "idx_seckill_activities_time" ON "seckill_activities" ("start_at", "end_at");

-- ----------------------------
-- 22. seckill_products（秒杀商品表）
-- ----------------------------
CREATE TABLE "seckill_products" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "seckill_id"      BIGINT NOT NULL,
    "product_id"      BIGINT NOT NULL,
    "sku_id"          BIGINT,
    "seckill_price"   NUMERIC(10,2) NOT NULL,
    "stock"           INT NOT NULL,
    "sold_count"      INT NOT NULL DEFAULT 0
);
CREATE INDEX "idx_seckill_products_tenant" ON "seckill_products" ("tenant_id");

-- ----------------------------
-- 23. distribution_relations（分销关系表）
-- ----------------------------
CREATE TABLE "distribution_relations" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "member_id"       BIGINT NOT NULL UNIQUE,
    "parent_id"       BIGINT NOT NULL,
    "level"           SMALLINT NOT NULL,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_distribution_relations_tenant" ON "distribution_relations" ("tenant_id");

-- ----------------------------
-- 24. distribution_commissions（分佣记录表）
-- ----------------------------
CREATE TABLE "distribution_commissions" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "order_id"        BIGINT NOT NULL,
    "order_item_id"   BIGINT,
    "buyer_id"        BIGINT NOT NULL,
    "agent_id"        BIGINT NOT NULL,
    "level"           SMALLINT NOT NULL,
    "commission_rate" NUMERIC(5,2) NOT NULL,
    "commission_amount" NUMERIC(10,2) NOT NULL,
    "status"          VARCHAR(20) NOT NULL DEFAULT 'pending',
    "settled_at"      TIMESTAMP,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_distribution_commissions_tenant" ON "distribution_commissions" ("tenant_id");
CREATE INDEX "idx_distribution_commissions_agent" ON "distribution_commissions" ("agent_id");

-- ----------------------------
-- 25. payments（支付记录表）
-- ----------------------------
CREATE TABLE "payments" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL REFERENCES "tenants"("id"),
    "payment_no"      VARCHAR(32) NOT NULL UNIQUE,
    "order_no"        VARCHAR(32),
    "member_id"       BIGINT NOT NULL,
    "amount"          NUMERIC(10,2) NOT NULL,
    "status"          VARCHAR(20) NOT NULL DEFAULT 'pending',
    "wechat_trade_type" VARCHAR(20),
    "wechat_transaction_id" VARCHAR(64),
    "wechat_payer_openid" VARCHAR(64),
    "wechat_paid_at"  TIMESTAMP,
    "refund_amount"   NUMERIC(10,2) DEFAULT 0,
    "refund_status"   VARCHAR(20) DEFAULT 'none',
    "closed_at"       TIMESTAMP,
    "close_reason"    VARCHAR(200),
    "expire_at"       TIMESTAMP,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_payments_tenant" ON "payments" ("tenant_id");
CREATE INDEX "idx_payments_order" ON "payments" ("order_no");
CREATE INDEX "idx_payments_wechat" ON "payments" ("wechat_transaction_id");
CREATE INDEX "idx_payments_status" ON "payments" ("status");

-- ----------------------------
-- 26. admin_action_logs（操作日志）
-- ----------------------------
CREATE TABLE "admin_action_logs" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL DEFAULT 0,
    "admin_id"        BIGINT NOT NULL,
    "admin_username"  VARCHAR(50) NOT NULL,
    "action"          VARCHAR(50) NOT NULL,
    "target_type"     VARCHAR(50),
    "target_id"       BIGINT,
    "target_desc"     VARCHAR(200),
    "request_method"  VARCHAR(10),
    "request_path"    VARCHAR(200),
    "request_body"   TEXT,
    "request_ip"      VARCHAR(50),
    "user_agent"      VARCHAR(500),
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_admin_action_logs_tenant" ON "admin_action_logs" ("tenant_id");
CREATE INDEX "idx_admin_action_logs_admin" ON "admin_action_logs" ("admin_id");
CREATE INDEX "idx_admin_action_logs_target" ON "admin_action_logs" ("target_type", "target_id");
CREATE INDEX "idx_admin_action_logs_time" ON "admin_action_logs" ("created_at");

-- ----------------------------
-- 27. uploads（文件上传记录）
-- ----------------------------
CREATE TABLE "uploads" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL DEFAULT 0,
    "file_key"        VARCHAR(255) NOT NULL,
    "original_name"   VARCHAR(255) NOT NULL,
    "file_size"       BIGINT NOT NULL,
    "file_type"       VARCHAR(50) NOT NULL,
    "file_ext"        VARCHAR(10) NOT NULL,
    "storage_type"    VARCHAR(20) NOT NULL DEFAULT 'local',
    "storage_url"    VARCHAR(500) NOT NULL,
    "uploaded_by"     BIGINT,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX "idx_uploads_tenant" ON "uploads" ("tenant_id");
CREATE INDEX "idx_uploads_uploader" ON "uploads" ("uploaded_by");

-- 28. tenant_payment_configs (商户收款配置)
-- 2026-04-21 新增：商户收款配置 与 物流承运商
CREATE TABLE IF NOT EXISTS "tenant_payment_configs" (
    "id"                BIGSERIAL PRIMARY KEY,
    "tenant_id"         BIGINT NOT NULL,
    "provider"          VARCHAR(20) NOT NULL DEFAULT 'wechat',
    "mch_id"            VARCHAR(64),
    "app_id"            VARCHAR(64),
    "api_v3_key"        VARCHAR(128),
    "cert_serial_no"    VARCHAR(64),
    "private_key_pem"   TEXT,
    "cert_pem"          TEXT,
    "notify_url"        VARCHAR(255),
    "enabled"           SMALLINT NOT NULL DEFAULT 0,
    "audit_status"      SMALLINT NOT NULL DEFAULT 0,
    "audit_remark"      VARCHAR(500) DEFAULT '',
    "submitted_at"      TIMESTAMPTZ,
    "audited_at"        TIMESTAMPTZ,
    "created_at"        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS "uniq_tpc_tenant_provider"
    ON "tenant_payment_configs" ("tenant_id", "provider");

CREATE TABLE IF NOT EXISTS "shipping_carriers" (
    "id"               BIGSERIAL PRIMARY KEY,
    "tenant_id"        BIGINT NOT NULL,
    "code"             VARCHAR(30) NOT NULL,
    "name"             VARCHAR(50) NOT NULL,
    "api_provider"     VARCHAR(30) DEFAULT '',
    "api_customer"     VARCHAR(128) DEFAULT '',
    "api_key"          VARCHAR(256) DEFAULT '',
    "api_secret"       VARCHAR(256) DEFAULT '',
    "priority"         INT NOT NULL DEFAULT 0,
    "enabled"          SMALLINT NOT NULL DEFAULT 0,
    "audit_status"     SMALLINT NOT NULL DEFAULT 0,
    "audit_remark"     VARCHAR(500) DEFAULT '',
    "submitted_at"     TIMESTAMPTZ,
    "audited_at"       TIMESTAMPTZ,
    "created_at"       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS "idx_shipping_carriers_tenant"
    ON "shipping_carriers" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_shipping_carriers_tenant_code"
    ON "shipping_carriers" ("tenant_id", "code");

-- 29. platform_settings（平台全局设置，单行 id=1）
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

-- 30. regions（平台统一维护省 / 市 / 区数据）
CREATE TABLE IF NOT EXISTS "regions" (
    "id"          BIGSERIAL PRIMARY KEY,
    "parent_id"   BIGINT NOT NULL DEFAULT 0,
    "code"        VARCHAR(32) NOT NULL DEFAULT '',
    "name"        VARCHAR(50) NOT NULL,
    "level"       SMALLINT NOT NULL DEFAULT 1,
    "sort"        INT NOT NULL DEFAULT 0,
    "enabled"     SMALLINT NOT NULL DEFAULT 1,
    "created_at"  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS "idx_regions_parent"
    ON "regions" ("parent_id");
CREATE INDEX IF NOT EXISTS "idx_regions_level_enabled"
    ON "regions" ("level", "enabled", "sort", "id");
CREATE INDEX IF NOT EXISTS "idx_regions_code"
    ON "regions" ("code");

-- 31. tenant_site_configs（租户站点 / 商城装修配置）
CREATE TABLE IF NOT EXISTS "tenant_site_configs" (
    "tenant_id"                        BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "custom_domain"                    VARCHAR(128) NOT NULL DEFAULT '',
    "domain_verified"                  SMALLINT NOT NULL DEFAULT 0,
    "ssl_status"                       VARCHAR(20) NOT NULL DEFAULT 'none',
    "brand_name"                       VARCHAR(100) NOT NULL DEFAULT '',
    "brand_logo"                       VARCHAR(500) NOT NULL DEFAULT '',
    "primary_color"                    VARCHAR(32) NOT NULL DEFAULT '#409EFF',
    "hide_platform_brand"              SMALLINT NOT NULL DEFAULT 0,
    "footer_text"                      VARCHAR(255) NOT NULL DEFAULT '',
    "deployment_mode"                  VARCHAR(16) NOT NULL DEFAULT 'shared',
    "private_endpoint"                 VARCHAR(255) NOT NULL DEFAULT '',
    "private_notes"                    VARCHAR(500) NOT NULL DEFAULT '',
    "storefront_notice"                VARCHAR(255) NOT NULL DEFAULT '',
    "storefront_hero_title"            VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_hero_subtitle"         VARCHAR(500) NOT NULL DEFAULT '',
    "storefront_search_placeholder"    VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_category_title"        VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_coupon_title"          VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_hot_title"             VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_recommend_title"       VARCHAR(120) NOT NULL DEFAULT '',
    "storefront_quick_entries"         TEXT NOT NULL DEFAULT '[]',
    "storefront_service_cards"         TEXT NOT NULL DEFAULT '[]',
    "storefront_banners"               TEXT NOT NULL DEFAULT '[]',
    "storefront_promo_cards"           TEXT NOT NULL DEFAULT '[]',
    "storefront_member_entries"        TEXT NOT NULL DEFAULT '[]',
    "storefront_home_sections"         TEXT NOT NULL DEFAULT '[]',
    "storefront_profile_sections"      TEXT NOT NULL DEFAULT '[]',
    "storefront_search_keywords"       TEXT NOT NULL DEFAULT '[]',
    "updated_at"                       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 32. runtime flow tables（运行态功能表，保持和 Go model / migrations 一致）
CREATE TABLE IF NOT EXISTS "tenant_subscription_orders" (
    "id"                 BIGSERIAL PRIMARY KEY,
    "tenant_id"          BIGINT NOT NULL,
    "plan_id"            BIGINT NOT NULL,
    "billing_cycle"      VARCHAR(10) NOT NULL,
    "amount"             NUMERIC(10,2) NOT NULL,
    "status"             SMALLINT NOT NULL DEFAULT 0,
    "order_no"           VARCHAR(64) NOT NULL UNIQUE,
    "pay_transaction_id" VARCHAR(64) NOT NULL DEFAULT '',
    "pay_at"             TIMESTAMP NULL,
    "expire_before"      TIMESTAMP NULL,
    "expire_after"       TIMESTAMP NULL,
    "created_at"         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_tenant_sub_orders_tenant" ON "tenant_subscription_orders" ("tenant_id", "status");
CREATE INDEX IF NOT EXISTS "idx_tenant_sub_orders_created" ON "tenant_subscription_orders" ("created_at" DESC);

CREATE TABLE IF NOT EXISTS "points_settings" (
    "tenant_id"   BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "enabled"     SMALLINT NOT NULL DEFAULT 1,
    "earn_rate"   NUMERIC(10,4) NOT NULL DEFAULT 1,
    "min_amount"  NUMERIC(10,2) NOT NULL DEFAULT 0,
    "redeem_rate" INT NOT NULL DEFAULT 100,
    "remark"      VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "delivery_settings" (
    "tenant_id"            BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "express_enabled"      SMALLINT NOT NULL DEFAULT 1,
    "express_free_amount"  NUMERIC(10,2) NOT NULL DEFAULT 0,
    "express_default_fee"  NUMERIC(10,2) NOT NULL DEFAULT 0,
    "city_enabled"         SMALLINT NOT NULL DEFAULT 0,
    "city_radius_km"       NUMERIC(6,2) NOT NULL DEFAULT 5,
    "city_base_fee"        NUMERIC(10,2) NOT NULL DEFAULT 5,
    "city_per_km_fee"      NUMERIC(10,2) NOT NULL DEFAULT 1,
    "city_min_order"       NUMERIC(10,2) NOT NULL DEFAULT 0,
    "pickup_enabled"       SMALLINT NOT NULL DEFAULT 0,
    "pickup_address"       VARCHAR(255) NOT NULL DEFAULT '',
    "pickup_hours"         VARCHAR(100) NOT NULL DEFAULT '',
    "pickup_phone"         VARCHAR(30) NOT NULL DEFAULT '',
    "remark"               VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "sms_settings" (
    "tenant_id"     BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "enabled"       SMALLINT NOT NULL DEFAULT 0,
    "provider"      VARCHAR(32) NOT NULL DEFAULT 'aliyun',
    "access_key"    VARCHAR(128) NOT NULL DEFAULT '',
    "access_secret" VARCHAR(256) NOT NULL DEFAULT '',
    "sign_name"     VARCHAR(64) NOT NULL DEFAULT '',
    "region"        VARCHAR(32) NOT NULL DEFAULT '',
    "remark"        VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "sms_templates" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "code"        VARCHAR(64) NOT NULL,
    "name"        VARCHAR(100) NOT NULL,
    "template_id" VARCHAR(64) NOT NULL DEFAULT '',
    "content"     VARCHAR(500) NOT NULL DEFAULT '',
    "enabled"     SMALLINT NOT NULL DEFAULT 1,
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS "uk_sms_templates_tenant_code" ON "sms_templates" ("tenant_id", "code");

CREATE TABLE IF NOT EXISTS "sms_logs" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "phone"       VARCHAR(20) NOT NULL,
    "code"        VARCHAR(64) NOT NULL,
    "content"     VARCHAR(500) NOT NULL DEFAULT '',
    "status"      SMALLINT NOT NULL DEFAULT 1,
    "error"       VARCHAR(500) NOT NULL DEFAULT '',
    "biz_id"      VARCHAR(64) NOT NULL DEFAULT '',
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_sms_logs_tenant" ON "sms_logs" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_sms_logs_phone" ON "sms_logs" ("tenant_id", "phone");

CREATE TABLE IF NOT EXISTS "distribution_settings" (
    "tenant_id"     BIGINT PRIMARY KEY REFERENCES "tenants"("id"),
    "enabled"       SMALLINT NOT NULL DEFAULT 1,
    "level1_rate"   NUMERIC(5,4) NOT NULL DEFAULT 0.10,
    "level2_rate"   NUMERIC(5,4) NOT NULL DEFAULT 0.05,
    "min_withdraw"  NUMERIC(10,2) NOT NULL DEFAULT 10,
    "auto_become"   SMALLINT NOT NULL DEFAULT 0,
    "remark"        VARCHAR(500) NOT NULL DEFAULT '',
    "updated_at"    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "distributors" (
    "id"                   BIGSERIAL PRIMARY KEY,
    "tenant_id"            BIGINT NOT NULL,
    "member_id"            BIGINT NOT NULL,
    "parent_id"            BIGINT NOT NULL DEFAULT 0,
    "grandparent_id"       BIGINT NOT NULL DEFAULT 0,
    "status"               SMALLINT NOT NULL DEFAULT 0,
    "total_commission"     NUMERIC(12,2) NOT NULL DEFAULT 0,
    "pending_commission"   NUMERIC(12,2) NOT NULL DEFAULT 0,
    "withdrawn"            NUMERIC(12,2) NOT NULL DEFAULT 0,
    "invite_count"         INT NOT NULL DEFAULT 0,
    "created_at"           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "approved_at"          TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS "uk_distributors_tenant_member" ON "distributors" ("tenant_id", "member_id");
CREATE INDEX IF NOT EXISTS "idx_distributors_parent" ON "distributors" ("tenant_id", "parent_id");
CREATE INDEX IF NOT EXISTS "idx_distributors_status" ON "distributors" ("tenant_id", "status");

CREATE TABLE IF NOT EXISTS "commission_logs" (
    "id"             BIGSERIAL PRIMARY KEY,
    "tenant_id"      BIGINT NOT NULL,
    "distributor_id" BIGINT NOT NULL,
    "member_id"      BIGINT NOT NULL,
    "order_id"       BIGINT NOT NULL,
    "order_no"       VARCHAR(64) NOT NULL,
    "buyer_id"       BIGINT NOT NULL,
    "level"          SMALLINT NOT NULL,
    "amount"         NUMERIC(12,2) NOT NULL,
    "rate"           NUMERIC(5,4) NOT NULL,
    "status"         SMALLINT NOT NULL DEFAULT 1,
    "settled_at"     TIMESTAMP,
    "created_at"     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_commission_logs_tenant" ON "commission_logs" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_commission_logs_distributor" ON "commission_logs" ("tenant_id", "distributor_id");
CREATE INDEX IF NOT EXISTS "idx_commission_logs_order" ON "commission_logs" ("tenant_id", "order_id");

CREATE TABLE IF NOT EXISTS "groupon_activities" (
    "id"             BIGSERIAL PRIMARY KEY,
    "tenant_id"      BIGINT NOT NULL,
    "name"           VARCHAR(100) NOT NULL,
    "product_id"     BIGINT NOT NULL,
    "sku_id"         BIGINT NOT NULL DEFAULT 0,
    "group_price"    NUMERIC(10,2) NOT NULL,
    "original_price" NUMERIC(10,2) NOT NULL,
    "require_num"    INT NOT NULL DEFAULT 2,
    "expire_hours"   INT NOT NULL DEFAULT 24,
    "total_stock"    INT NOT NULL DEFAULT 0,
    "sold_count"     INT NOT NULL DEFAULT 0,
    "start_at"       TIMESTAMP NOT NULL,
    "end_at"         TIMESTAMP NOT NULL,
    "status"         SMALLINT NOT NULL DEFAULT 1,
    "created_at"     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_groupon_activities_tenant" ON "groupon_activities" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_groupon_activities_status" ON "groupon_activities" ("tenant_id", "status");

CREATE TABLE IF NOT EXISTS "groupons" (
    "id"              BIGSERIAL PRIMARY KEY,
    "tenant_id"       BIGINT NOT NULL,
    "activity_id"     BIGINT NOT NULL,
    "leader_id"       BIGINT NOT NULL,
    "require_num"     INT NOT NULL,
    "current_num"     INT NOT NULL DEFAULT 1,
    "status"          SMALLINT NOT NULL DEFAULT 1,
    "expires_at"      TIMESTAMP NOT NULL,
    "succeed_at"      TIMESTAMP,
    "created_at"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_groupons_tenant_activity" ON "groupons" ("tenant_id", "activity_id");
CREATE INDEX IF NOT EXISTS "idx_groupons_status" ON "groupons" ("tenant_id", "status");

CREATE TABLE IF NOT EXISTS "groupon_members" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "groupon_id"  BIGINT NOT NULL,
    "member_id"   BIGINT NOT NULL,
    "order_id"    BIGINT NOT NULL DEFAULT 0,
    "is_leader"   SMALLINT NOT NULL DEFAULT 0,
    "joined_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_groupon_members_groupon" ON "groupon_members" ("tenant_id", "groupon_id");
CREATE INDEX IF NOT EXISTS "idx_groupon_members_member" ON "groupon_members" ("tenant_id", "member_id");

CREATE TABLE IF NOT EXISTS "api_tokens" (
    "id"           BIGSERIAL PRIMARY KEY,
    "tenant_id"    BIGINT NOT NULL,
    "name"         VARCHAR(100) NOT NULL,
    "app_key"      VARCHAR(64) NOT NULL,
    "app_secret"   VARCHAR(128) NOT NULL,
    "scopes"       VARCHAR(500) NOT NULL DEFAULT '',
    "ip_whitelist" VARCHAR(500) NOT NULL DEFAULT '',
    "status"       SMALLINT NOT NULL DEFAULT 1,
    "last_used_at" TIMESTAMP NULL,
    "expires_at"   TIMESTAMP NULL,
    "created_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS "uk_api_tokens_app_key" ON "api_tokens" ("app_key");
CREATE INDEX IF NOT EXISTS "idx_api_tokens_tenant" ON "api_tokens" ("tenant_id");

CREATE TABLE IF NOT EXISTS "api_request_logs" (
    "id"          BIGSERIAL PRIMARY KEY,
    "tenant_id"   BIGINT NOT NULL,
    "token_id"    BIGINT NOT NULL,
    "app_key"     VARCHAR(64) NOT NULL,
    "method"      VARCHAR(10) NOT NULL,
    "path"        VARCHAR(255) NOT NULL,
    "status_code" INT NOT NULL DEFAULT 0,
    "ip"          VARCHAR(64) NOT NULL DEFAULT '',
    "cost_ms"     INT NOT NULL DEFAULT 0,
    "created_at"  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_api_logs_tenant" ON "api_request_logs" ("tenant_id");
CREATE INDEX IF NOT EXISTS "idx_api_logs_token" ON "api_request_logs" ("token_id");
