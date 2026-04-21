-- 2026-04-21 新增：套餐功能目录（平台统一管理）
CREATE TABLE IF NOT EXISTS "plan_features" (
    "id"            BIGSERIAL PRIMARY KEY,
    "code"          VARCHAR(40) NOT NULL,
    "name"          VARCHAR(50) NOT NULL,
    "description"   VARCHAR(255) NOT NULL DEFAULT '',
    "group_name"    VARCHAR(30) NOT NULL DEFAULT '',
    "sort"          INT NOT NULL DEFAULT 0,
    "status"        SMALLINT NOT NULL DEFAULT 1,
    "created_at"    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at"    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS "uniq_plan_features_code" ON "plan_features" ("code");

INSERT INTO "plan_features" ("code","name","description","group_name","sort","status") VALUES
  ('multi_sku','多规格 SKU','支持按颜色/尺寸等多维属性管理库存与价格','商品',10,1),
  ('virtual_product','虚拟商品','卡密、课程等无需物流的商品类型','商品',20,1),
  ('seckill','限时秒杀','倒计时秒杀活动，提升瞬时转化','营销',30,1),
  ('group_buy','拼团','多人成团享折扣，低成本拉新','营销',40,1),
  ('distribution','分销','多级分销体系与佣金结算','营销',50,1),
  ('coupon','优惠券','现金券/折扣券/免邮券统一发放与核销','营销',60,1),
  ('points','积分','下单赚积分，积分抵现/兑换','会员',70,1),
  ('member_level','会员等级','按消费/成长值划分等级与差异化权益','会员',80,1),
  ('express_delivery','快递配送','接入主流快递，订单物流全链路跟踪','履约',90,1),
  ('city_delivery','同城配送','同城即时达/众包配送接入','履约',100,1),
  ('self_pickup','到店自提','支持到店/自提点自提模式','履约',110,1),
  ('custom_domain','品牌域名','小程序/H5 绑定品牌自定义域名','品牌',120,1),
  ('api_access','API 开放','开放商品/订单/会员等 REST API','开放',130,1),
  ('white_label','白标定制','全站品牌白标，无 SaaS 标识','品牌',140,1),
  ('sms_notification','短信通知','订单发货/退款短信通知','通知',150,1),
  ('private_deployment','私有部署','独立服务器/数据库部署交付','企业',160,1)
ON CONFLICT ("code") DO NOTHING;
