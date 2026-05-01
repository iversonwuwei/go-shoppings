# SaaS 微信商城系统 - 完整设计文档

> 技术栈：Go + Gin + GORM + MySQL + Redis + Docker
> 多租户模式：行级隔离（tenant_id）
> 套餐订阅：基础版 / 专业版 / 旗舰版

---

## 一、项目概述

### 1.1 系统定位

面向企业提供微信小程序商城 SaaS 服务，支持多租户入驻、订阅计费、微信生态深度对接。

### 1.2 支付账户归属原则

系统存在两条支付业务线，产品和账务边界必须隔离：

- 平台会员 / 套餐订阅付款：付款方为租户，收款方为平台，资金进入平台账户，用于开通、续订、升级或恢复套餐。
- 顾客购买商品付款：前期采用平台统一收款模式，付款方为 C 端顾客，收款方为平台微信支付商户号，订单归属和待结算租户通过 `payments.tenant_id` / `payments.settlement_tenant_id` 记录。
- 平台按账期对顾客订单货款做结算汇总，扣除平台服务费、退款、售后、分销佣金等应扣项后，再通过线下转账或后续自动打款能力结算给租户。
- 微信支付服务商 + 子商户模式作为后续增强能力保留字段和迁移空间，前期顾客订单支付不得强依赖租户子商户配置。
- 非生产环境可保留模拟支付以验证订单状态流；生产环境缺少平台微信支付配置时必须拒绝支付下单，不能降级为模拟成功。

### 1.3 小程序会员登录归属原则

平台默认采用统一微信小程序承载多租户商城。C 端用户进入某个租户商城后，会员登录应使用当前小程序的统一 AppID / AppSecret 调用微信 `code2session` 获取用户 `openid`，而不是使用租户自己的小程序 AppID。

会员归属由当前租户上下文决定：小程序先通过启动参数、scene 或默认租户编码解析出租户，后续会员接口必须携带 `X-Tenant-ID`。后端使用 `(tenant_id, openid)` 查找会员；若不存在，则创建带当前 `tenant_id` 的会员记录。同一个微信用户进入不同租户时，可以在不同 `tenant_id` 下拥有相互隔离的会员记录。

租户表中的 `wechat_appid` / `wechat_secret` 仅作为后续“租户独立小程序 / 私有化小程序”能力的扩展配置，不应作为平台统一小程序会员登录的前置条件。生产环境缺少平台统一小程序配置时，正式微信登录必须拒绝；非生产环境可使用开发登录兜底。

### 1.4 技术选型

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| 语言 | Go 1.21+ | 高并发、低内存、编译型二进制 |
| Web框架 | Gin 1.9+ | 性能优秀、中间件生态成熟 |
| ORM | GORM | Go ORM 主流选择 |
| 数据库 | PostgreSQL 15 | 事务支持、多租户行级隔离、JSONB原生支持 |
| 缓存 | Redis 7.0 | 租户配置缓存、分布式锁、库存扣减 |
| 容器 | Docker + docker-compose | 本地开发与生产部署 |
| 微信SDK | 原生实现（复用微信官方API） | 无第三方依赖 |

### 1.4 项目结构

```
wechat-mall-saas/
├── cmd/
│   ├── api/main.go              # HTTP API 服务入口
│   └── worker/main.go           # 异步任务/Cron 入口
├── internal/
│   ├── handler/                 # 控制器层（HTTP Handler）
│   │   ├── middleware/           # 中间件
│   │   │   ├── tenant.go         # 租户上下文注入
│   │   │   ├── auth.go           # JWT 鉴权
│   │   │   ├── cors.go           # 跨域
│   │   │   ├── ratelimit.go      # 限流
│   │   │   ├── recovery.go       # Panic 恢复
│   │   │   └── logging.go        # 请求日志
│   │   ├── admin/               # 管理端接口
│   │   │   ├── auth.go           # 管理员登录
│   │   │   ├── product.go        # 商品管理
│   │   │   ├── order.go          # 订单管理
│   │   │   ├── member.go         # 会员管理
│   │   │   ├── coupon.go         # 优惠券
│   │   │   ├── seckill.go        # 秒杀
│   │   │   ├── groupbuy.go       # 拼团
│   │   │   └── platform.go       # 平台运营（套餐/租户/财务）
│   │   ├── member/              # 小程序端接口
│   │   │   ├── auth.go           # 微信登录/手机号
│   │   │   ├── product.go        # 商品/分类
│   │   │   ├── order.go          # 订单
│   │   │   ├── cart.go           # 购物车
│   │   │   ├── coupon.go         # 优惠券
│   │   │   ├── address.go        # 收货地址
│   │   │   └── points.go         # 积分
│   │   └── payment.go           # 支付回调
│   ├── service/                  # 业务逻辑层
│   │   ├── product_service.go
│   │   ├── order_service.go
│   │   ├── member_service.go
│   │   ├── payment_service.go
│   │   ├── coupon_service.go
│   │   ├── seckill_service.go
│   │   ├── groupbuy_service.go
│   │   ├── distribution_service.go
│   │   ├── subscription_service.go  # 套餐/计费
│   │   ├── tenant_service.go
│   │   └── usage_service.go         # 用量校验
│   ├── repository/               # 数据访问层
│   │   ├── base_repo.go          # 基类（自动 tenant_id 注入）
│   │   ├── product_repo.go
│   │   ├── product_sku_repo.go
│   │   ├── order_repo.go
│   │   ├── member_repo.go
│   │   ├── coupon_repo.go
│   │   ├── tenant_repo.go
│   │   ├── plan_repo.go
│   │   └── payment_repo.go
│   ├── model/                    # 数据模型（带 GORM tags）
│   │   ├── product.go
│   │   ├── order.go
│   │   ├── member.go
│   │   ├── coupon.go
│   │   ├── tenant.go
│   │   ├── plan.go
│   │   └── payment.go
│   └── pkg/                     # 公共工具包
│       ├── wxapp/                # 微信小程序 SDK
│       │   ├── client.go
│       │   ├── auth.go
│       │   └── decrypt.go
│       ├── wxpay/                # 微信支付 SDK
│       │   ├── client.go
│       │   ├── v3.go
│       │   ├── order.go
│       │   ├── callback.go
│       │   └── refund.go
│       ├── response/             # 统一响应
│       ├── errors/               # 业务错误定义
│       ├── logger/               # Zap 日志
│       ├── cache/                # Redis 封装
│       ├── jwt/                  # JWT 封装
│       └── utils/                # 工具函数（Snowflake/加密/字符串）
├── configs/
│   ├── config.yaml               # 主配置
│   ├── config_prod.yaml          # 生产配置
│   └── config_test.yaml          # 测试配置
├── scripts/
│   ├── init_db.sql              # 建表 SQL
│   └── seed.sql                 # 初始数据（套餐）
├── test/
│   ├── handler/
│   ├── service/
│   └── integration/
├── Makefile
├── go.mod
├── go.sum
├── .air.toml                    # Air 热重载
├── docker-compose.infra.yml      # PostgreSQL + Redis + MinIO
└── docker-compose.app.yml        # API + AI 图片服务
```

---

## 二、多租户架构设计

### 2.1 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        负载均衡层                            │
│                    （Nginx / 云 LB）                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       API Gateway                           │
│            （租户识别 / 鉴权 / 限流 / 路由）                  │
└─────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
   ┌────────────┐      ┌────────────┐      ┌────────────┐
   │  Mall API  │      │  Admin API │      │   Worker   │
   │   (Gin)    │      │   (Gin)    │      │   (Cron)   │
   └────────────┘      └────────────┘      └────────────┘
          │                   │                   │
          └───────────────────┼───────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
   ┌────────────┐      ┌────────────┐      ┌────────────┐
   │  (多租户库) │      │  (租户缓存) │      │  (OSS/COS) │
   │  (多租户库) │      │  (租户缓存) │      │  (OSS/COS) │
   └────────────┘      └────────────┘      └────────────┘
```

### 2.2 多租户隔离策略

**行级隔离**：所有业务表包含 `tenant_id` 字段，Repository 层自动追加条件。

**TenantContext 注入流程：**

```
1. 请求进入 → TenantMiddleware
2. 从 Header X-Tenant-ID 取租户 ID
3. 验证租户状态（status != 1 → 403）
4. 加载租户套餐配置（Redis 缓存，5分钟 TTL）
5. 注入 TenantContext 到 context.Context
6. 后续 Handler/Service/Repository 自动获取
```

**Repository 基类实现：**

```go
// 所有查询自动追加 tenant_id 条件
// 所有 INSERT 自动填入 tenant_id
// 所有 UPDATE/DELETE 携带 tenant_id 条件
// 防止跨租户数据访问
```

### 2.3 套餐功能开关

每个套餐定义 `features` JSON 数组，所有功能型操作前必须校验：

```go
// 套餐功能枚举
const (
    FeatureMultiSKU          = "multi_sku"         // 多规格SKU
    FeatureVirtualProduct     = "virtual_product"    // 虚拟商品
    FeatureSeckill           = "seckill"            // 秒杀
    FeatureGroupBuy          = "group_buy"          // 拼团
    FeatureDistribution       = "distribution"       // 分销
    FeatureCoupon             = "coupon"             // 优惠券
    FeaturePoints             = "points"             // 积分体系
    FeatureMemberLevel        = "member_level"       // 会员等级
    FeatureExpressDelivery    = "express_delivery"  // 快递配送
    FeatureCityDelivery       = "city_delivery"      // 同城配送
    FeatureSelfPickup         = "self_pickup"       // 到店自提
    FeatureCustomDomain       = "custom_domain"      // 自定义域名
    FeaturePrivateDeployment   = "private_deployment" // 私有化部署
    FeatureWhiteLabel         = "white_label"        // 白标
    FeatureSmsNotification    = "sms_notification"  // 短信通知
)
```

---

## 三、数据库表结构设计

### 3.1 平台层表

#### plans（套餐表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键 |
| name | VARCHAR(50) | 套餐名 |
| code | VARCHAR(30) UNIQUE | 套餐代码 |
| monthly_fee | DECIMAL(10,2) | 月费 |
| yearly_fee | DECIMAL(10,2) | 年费 |
| product_limit | INT | 商品数量上限（0=无限制） |
| order_limit | INT | 月订单上限（0=无限制） |
| user_limit | INT | 会员数量上限（0=无限制） |
| features | JSON | 功能列表 |
| is_default | TINYINT(1) | 是否默认套餐 |
| status | TINYINT(1) | 状态 |

#### tenants（租户表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键 |
| code | VARCHAR(30) UNIQUE | 租户编码 |
| company_name | VARCHAR(100) | 公司名称 |
| contact_name | VARCHAR(50) | 联系人 |
| contact_phone | VARCHAR(20) | 联系电话 |
| contact_email | VARCHAR(100) | 联系邮箱 |
| wechat_appid | VARCHAR(50) | 租户独立小程序 AppID（预留扩展；平台统一小程序登录不依赖） |
| wechat_secret | VARCHAR(100) | 租户独立小程序 AppSecret（加密，预留扩展） |
| wechat_mchid | VARCHAR(30) | 历史直连商户号字段，仅兼容旧配置；前期顾客订单生产支付以平台 `wxpay_mch_id` 为准 |
| wechat_apiv3_key | VARCHAR(100) | 历史直连 APIv3 密钥（加密），平台统一收款模式不依赖该字段 |
| wechat_cert_serial | VARCHAR(100) | 历史直连支付证书序列号 |
| plan_id | BIGINT UNSIGNED | 套餐ID（外键） |
| plan_expire_at | DATETIME | 套餐到期时间 |
| brand_name | VARCHAR(50) | 店铺名称 |
| brand_logo | VARCHAR(255) | 店铺Logo |
| brand_theme | VARCHAR(20) | 主题色 |
| brand_domain | VARCHAR(100) | 独立域名 |
| billing_cycle | VARCHAR(10) | 计费周期：monthly/yearly |
| extra_features | JSONB | 平台额外授予功能列表 |
| status | TINYINT(1) | 状态：0待审核/1正常/2欠费/3封禁 |
| reject_reason | VARCHAR(255) | 审核拒绝原因 |

#### tenant_plan_logs（套餐变更记录）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED | 主键 |
| tenant_id | BIGINT UNSIGNED | 租户ID |
| old_plan_id | BIGINT UNSIGNED | 变更前套餐 |
| new_plan_id | BIGINT UNSIGNED | 变更后套餐 |
| change_type | VARCHAR(20) | 类型：create/renew/upgrade/downgrade |
| effective_at | DATETIME | 生效时间 |
| expire_at | DATETIME | 到期时间 |
| amount | DECIMAL(10,2) | 金额 |

### 3.2 商品层表

#### products（商品主表）

所有字段含 `tenant_id`，索引：idx_tenant, idx_category, idx_status, idx_sold DESC, FULLTEXT(name)

关键字段：name, subtitle, cover_image, images(JSON), video_url, description(HTML富文本), price, cost_price, stock, stock_warning, has_sku, delivery_type(JSON), delivery_fee, status, is_recommend, is_hot, seo_*, sort, sold_count, view_count, deleted_at(软删除)

#### product_categories（分类表）

含 tenant_id，parent_id（0=一级），支持树形结构。

#### product_skus（SKU表）

tenant_id + product_id，sku_code 在租户内唯一，attributes(JSON 规格组合), price, cost_price, stock, image

#### product_attributes / product_attribute_values（规格属性/值表）

独立表，支持多规格组合（颜色+尺码+内存等）。

### 3.3 会员层表

#### members（会员表）

tenant_id + openid 唯一，含 level_id（会员等级）、points、growth_value、parent_id（推荐人）、level1_count、level2_count

租户后台会员管理必须以当前登录商户的 `tenant_id` 为边界：商户只能分页查看、筛选、启用/禁用、查看详情和调整自己租户下的会员。会员详情可展示基础资料、收货地址和最近积分明细；禁用会员后，会员端再次登录或继续访问需要鉴权的会员接口时必须被拦截，避免禁用只停留在后台展示层。会员等级调整属于 `member_level` 功能范围；未开通该功能的租户仍可使用会员列表、详情和状态管理。

小程序端会员登录以当前租户下微信用户的 `openid` 为身份主键：微信用户进入商城后调用 `wx.login` 获取 code，后端向微信换取 openid；若当前租户已存在该 openid 的会员，则读取并刷新该会员的会话、微信资料和最近登录时间；若不存在，则自动创建会员后签发会员 Token，并继续执行用户原本要完成的浏览、领券、下单、查看订单等操作。手机号属于后续授权绑定资料，不是微信登录/自动注册的前置条件；数据库唯一约束只能限制已绑定手机号的会员，不能阻塞多个未绑定手机号的微信会员注册。

本地开发环境（`app.env != production`）允许平台统一小程序 AppID/Secret 暂未配置：`login-by-wechat` 会降级为开发微信会员登录，使用当前租户内固定的本地 openid 创建或读取会员，便于开发者工具调试登录后流程。生产环境必须配置平台统一小程序 AppID/Secret，不能启用该降级路径。

#### member_levels（会员等级表）

含 tenant_id，min_growth（成长值门槛）、discount_rate（折扣率）、points_mult（积分倍数）

#### points_logs（积分变动记录）

tenant_id + member_id，含 change_type(order/gift/sign/refund/manual)、change_value、balance_before/after

### 3.4 订单层表

#### orders（订单主表）

tenant_id + member_id + order_no(唯一)

状态流：pending_pay（待支付）→ paid / preparing（待发货）→ shipped / delivered（待收货/待确认）→ completed
取消/退款：cancelled / refunding / refunded

含完整的收货信息（receiver_*）、配送信息（express_company, express_no, self_pickup_code）、金额明细（total_amount, delivery_fee, discount_amount, coupon_id, points_discount, actual_amount）

索引：idx_tenant, idx_member, idx_status, idx_order_no, idx_paid_at, idx_created

#### order_items（订单商品明细）

下单时快照商品信息（product_name, sku_desc, cover_image, price, quantity），含退款状态 refund_status

#### order_logs（订单操作日志）

记录每次状态变更，含 operator_type(system/member/admin)、action、before_status、after_status、remark

#### after_sale_orders（售后单表）

顾客订单售后以独立售后单承载，首版只做整单售后闭环，后续再扩展到明细级部分退款：

- 基础字段：`tenant_id`、`after_sale_no`、`order_id`、`order_no`、`order_item_id`、`member_id`、`type`、`status`
- 申请字段：`reason`、`description`、`images`、`amount`、`order_status_before`、`applied_at`
- 审核字段：`audit_remark`、`audited_at`
- 退货字段：`return_express_company`、`return_express_no`、`returned_at`、`received_at`
- 退款字段：`refund_no`、`refund_remark`、`refunded_at`、`cancelled_at`、`created_at`、`updated_at`

`type` 取值：`refund`（仅退款）、`return_refund`（退货退款）。`status` 取值：`pending`、`approved`、`rejected`、`returning`、`received`、`refunded`、`cancelled`。

售后状态机：会员在已支付、处理中、已发货、已送达或已完成订单上发起售后，订单进入 `refunding`；商户审核通过后，仅退款可直接标记退款完成，退货退款需要会员提交退货物流，商户确认收货后再标记退款完成；驳回或会员取消时订单恢复到 `order_status_before`；退款完成时订单进入 `refunded`。生产环境接入微信退款前，管理端的“退款完成”只能表示人工或线下退款状态确认，不能替代真实支付退款调用。

#### after_sale_reasons（售后原因配置表）

售后原因由平台统一维护，租户和小程序只读取已启用原因，避免会员端自由输入导致统计和审核不可控：

- 基础字段：`id`、`code`、`label`、`type`、`sort_order`、`enabled`、`created_at`、`updated_at`
- `type` 取值：`all`、`refund`、`return_refund`；小程序按当前售后类型过滤展示 `all + 当前类型` 的启用原因。
- 平台端可新增、编辑、停用、删除原因；删除只允许用于未被售后单引用的配置，生产建议优先停用。
- 初始化数据至少包含：不想要了、拍错/多拍、商品破损、商品与描述不符、少件/漏发、质量问题、协商一致退款。

### 3.5 营销层表

#### coupons（优惠券表）

tenant_id，含 type(cash/discount/shipping)、threshold_amount、discount_value、max_discount、发放/使用时间控制、applicable_type(all/product/category)

#### member_coupons（会员优惠券记录）

唯一索引：(member_id, coupon_id, received_at)，含下单时快照字段

#### group_buys / group_buy_orders（拼团）

#### seckills / seckill_products（秒杀）

#### distribution_relations（分销关系）

唯一索引：member_id（每个会员只有一个上级）

#### distribution_commissions（分佣记录）

含 level（1/2级）、commission_rate、commission_amount、status(pending/settled/withdrawn)

#### groupon_activities / groupons / groupon_members（拼团运行表）

拼团管理端和小程序拼团流程使用运行态拼团模型：`groupon_activities` 记录活动、商品、团购价、成团人数和有效期；`groupons` 记录每个团实例和成团状态；`groupon_members` 记录参团会员与订单绑定。初始化和补丁脚本必须创建这三张表，避免拼团页面或接口因缺表失败。

#### distribution_settings / distributors / commission_logs（分销运行表）

分销中心、分销员审核、佣金结算流程使用运行态分销模型：`distribution_settings` 记录租户分销规则；`distributors` 记录会员分销员状态和上下级；`commission_logs` 记录订单产生的待结算/已结算佣金。历史 `distribution_relations`、`distribution_commissions` 可作为旧结构保留，但新流程以运行态模型表为准。

#### points_settings（积分规则表）

每个租户一行，记录积分开关、下单积分发放比例、最低发放金额和抵扣汇率。积分明细仍写入 `points_logs`。

#### sms_settings / sms_templates / sms_logs（短信通知表）

短信配置、模板和发送日志分表存储，覆盖验证码、订单通知、套餐到期提醒等流程。`tenant_id=0` 固定表示平台自身短信能力，用于平台入驻申请、平台管理员登录、找回密码等平台侧验证码或通知；`tenant_id>0` 表示租户自己的短信能力，应由租户管理/商户侧配置和使用。

平台端 `/platform/sms` 只维护平台自身短信能力和平台功能到短信模板的绑定，不承载租户短信审核或租户模板维护。租户不需要配置短信服务，平台入驻申请、平台用户短信登录、找回密码等验证码统一使用平台短信服务。

阿里云短信发送使用已有短信服务时，平台不在本系统内申请模板或签名，签名审核与模板审核都在阿里云控制台完成。平台页面维护 `access_key`/`access_secret`、已审核签名名称、以及一个阿里云验证码 `TemplateCode`（例如 `SMS_123456789`）。启用短信时必须同时填写 AK/SK、签名名称和 TemplateCode；保存时系统自动把该模板 Code 绑定到 `apply`（入驻申请验证码）、`login`（平台用户短信登录）、`reset_password`（平台用户找回密码）三个平台验证码用途。签名名称保存到平台短信配置，作为运行时 `ALIYUN_SMS_SIGN_NAME` 的页面配置来源；环境变量 `ALIYUN_SMS_SIGN_NAME` 仅作为后端兜底。页面不提供 Endpoint/Region、备注等手动配置项。发送时后端调用阿里云 `SendSms`，传入 `PhoneNumbers`、页面保存或服务端配置的已审核签名名称、`TemplateCode` 和 `TemplateParam={"code":"六位验证码"}`；签名名称不包含 `【】`；阿里云模板变量应使用 `${code}`。

#### api_tokens / api_request_logs（开放 API 表）

开放 API 功能需要 `api_tokens` 存储租户密钥、权限范围和状态，`api_request_logs` 记录调用路径、状态码、耗时和来源 IP。

### 3.6 支付层表

#### payments（支付记录表）

顾客订单支付记录，使用 `tenant_id + member_id` 归属到实际经营租户，`payment_no` 全局唯一。前期生产环境下该表记录平台统一收款的微信支付下单与回调结果：

- 基础字段：`tenant_id`、`member_id`、`payment_no`、`order_no`、`amount`、`status`、`expire_at`、`closed_at`、`close_reason`
- 微信交易字段：`wechat_trade_type`、`wechat_transaction_id`、`wechat_payer_openid`、`wechat_paid_at`
- 平台收款与结算字段：`settlement_tenant_id`、`pay_scene`，以及可选的 `sp_appid/sp_mchid/sub_appid/sub_mchid` 兼容字段；平台统一收款模式下子商户字段可为空，`sp_appid/sp_mchid` 可作为平台收款账户快照使用。
- 退款字段：`refund_amount`、`refund_status`

`pay_scene` 固定区分 `member_order` 与后续扩展场景；顾客订单支付只能写 `member_order`，不得复用租户订阅订单号。回调处理必须以 `payment_no/out_trade_no` 幂等更新支付状态，再推进订单、库存、积分与分佣。

#### tenant_payment_configs（租户结算资料配置）

每个租户每种结算方式一条配置。前期顾客订单支付统一进入平台账户，该表不再作为顾客下单前置条件，而用于租户提交结算资料、平台审核和后续账期结算：

- `provider`: 前期使用 `manual_settlement`，后续可扩展 `wechat` / `alipay` / `bank`
- `settlement_account_name`、`settlement_account_no`、`settlement_bank_name`、`settlement_remark`: 租户账期结算资料，审核通过后供平台财务放款使用
- `mch_id`、`app_id`、`sub_mchid`、`sub_appid`: 历史直连和服务商字段，仅兼容旧配置或后续服务商模式，不作为前期顾客订单支付必填项
- `notify_url`: 前期可作为结算备注或外部结算系统回调地址保留，不参与顾客订单支付下单
- `enabled`、`audit_status`、`audit_remark`、`submitted_at`、`audited_at`: 平台审核与启用状态

商户提交结算资料后必须经平台审核。审核通过且 `enabled=1` 后，该租户进入账期结算候选范围；审核中、驳回或未提交不会阻塞 C 端顾客下单支付，但平台结算时应提示该租户结算资料未完成。

#### tenant_subscription_orders（租户订阅订单表）

租户续费、升级、降级前先创建订阅订单，记录 `plan_id`、`billing_cycle`、`amount`、`order_no`、创建管理员、支付流水、支付前后到期时间。微信支付回调成功后再更新租户套餐和 `tenant_plan_logs`。

订阅订单必须记录创建人快照：`created_by_admin_id` 保存创建订单的商户后台管理员 ID，`created_by_admin_username` 保存创建时的用户名快照。列表页和审计排查默认展示创建人，避免多个商户管理员共用同一租户时无法追溯谁发起了订阅订单。

订阅付款的收款方是平台，使用平台自有商户号或服务商体系下的平台自营收款商户。订阅订单不得写入 `payments` 表，避免和顾客订单货款混账；回调入口使用 `/api/v1/public/subscription/callback` 或等价的订阅专用回调。

入驻后的默认付款闭环如下：

1. 访客在公开页提交入驻申请（可选择套餐与计费周期），系统创建租户与管理员账号，租户状态为 `pending`。
2. 平台在“租户审核”中审核通过，租户进入 `active` 并开启试用期。
3. 商户使用申请时创建的管理员账号登录后台，进入“订阅付费”（`/admin/billing`）创建订阅订单。
4. 订阅订单支付成功后，系统通过订阅回调更新 `tenant_subscription_orders`、租户 `plan_expire_at` 与 `tenant_plan_logs`。
5. 若商户在申请后看不到付款入口，前端必须明确提示“先审核通过，再登录后台在订阅付费中完成付款”，避免用户误以为无付款通道。

平台“租户审核/租户管理”列表必须展示租户订阅摘要：是否已付费、会员起始时间、会员到期时间。`is_paid` 由该租户是否存在已支付订阅订单判定；`membership_start_at` 表示租户整体会员/试用周期的起点，取租户 `created_at`；`membership_end_at` 取当前租户 `plan_expire_at`。最近一次订阅订单的 `expire_before` / `expire_after` 只用于订阅订单明细和审计排查，不用于覆盖租户列表的整体会员周期，避免续费场景下列表显示成未来某一笔订单的续费区间。

订阅订单的 `expire_before` 必须在支付成功落账时刷新为“本次续期实际基准到期时间”，不能只保留创建订单时的快照。若同一租户存在多笔待支付订单并依次支付，每笔订单只能从付款时的当前到期日继续延长一个计费周期，订单自身展示区间应始终保持月付约 1 个月、年付约 1 年，不能因为多笔订单叠加而表现为单笔订单订阅多年。

### 3.7 系统层表

#### admin_action_logs（操作日志）

含 tenant_id（平台管理员为0）、admin_id、action、target_*、request_*、request_ip、user_agent

#### uploads（文件记录）

含 tenant_id、storage_type(local/oss/cos)、storage_url

#### delivery_settings（配送设置表）

每个租户一行，覆盖快递、同城配送、自提三类配送能力的开关、费用、半径、门店地址和联系电话。

### 3.8 数据库脚本覆盖要求

- `scripts/init_db.sql` 必须覆盖基础表和当前运行态模型所需表；后续增量统一放入 `scripts/migrations/`，并保持 `CREATE TABLE IF NOT EXISTS`、`ALTER TABLE ... ADD COLUMN IF NOT EXISTS`、`CREATE INDEX IF NOT EXISTS` 的幂等写法。
- 本地演示库必须至少包含：`api_tokens`、`api_request_logs`、`sms_settings`、`sms_templates`、`sms_logs`、`after_sale_orders`、`tenant_payment_configs`、`distribution_settings`、`distributors`、`commission_logs`、`groupon_activities`、`groupons`、`groupon_members`、`points_settings`、`delivery_settings`、`tenant_subscription_orders`。
- 租户表必须包含 `billing_cycle` 和 `extra_features`，用于入驻计费周期和平台额外授权功能。缺失字段或缺失表应通过可重复执行的补丁脚本修复，并在执行后用 information_schema 校验。
- 平台统一收款迁移必须保留 `payments.settlement_tenant_id/pay_scene` 用于账期结算归属；`sp_mchid/sp_appid/sub_mchid/sub_appid` 字段作为服务商模式兼容字段保留，不作为前期支付下单门禁。

---

## 四、API 接口设计

### 4.1 API 规范

- 路径格式：`/api/v1/{module}/{resource}`
- 认证：`Header Authorization: Bearer {token}`
- 租户标识：`Header X-Tenant-ID: {tenant_id}`
- 请求/响应：`JSON`
- 响应格式：`{ code: 0, message: "success", data: {} }`
- 时间入参：管理端所有时间字段优先使用 RFC3339 / RFC3339Nano（如 `2026-04-27T00:00:00+08:00`）；为兼容 Element Plus 日期选择器和历史前端，后端同时接受 `YYYY-MM-DDTHH:mm:ss`、`YYYY-MM-DD HH:mm:ss`、`YYYY-MM-DD`，无时区格式按服务本地时区解析。涉及优惠券、秒杀、拼团、租户套餐到期时间等字段时，前端默认提交带时区格式。

### 4.2 接口列表

#### 认证与租户

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/admin/auth/login | 管理员登录 |
| POST | /api/v1/admin/auth/refresh | 刷新Token |
| POST | /api/v1/tenant/register | 提交入驻申请 |
| GET | /api/v1/tenant/plans | 获取套餐列表（公开） |
| GET | /api/v1/tenant/info | 获取当前租户信息 |
| PUT | /api/v1/tenant/info | 修改租户信息 |
| PUT | /api/v1/tenant/wechat-config | 配置微信信息 |
| GET | /api/v1/platform/tenants | 租户列表（分页+筛选） |
| POST | /api/v1/platform/tenants/{id}/audit | 审核租户 |
| PUT | /api/v1/platform/tenants/{id}/status | 修改租户状态 |

本地演示账号：平台端使用 `admin / admin123`；商户端使用租户编号 `TEST001`、账号 `smokeadmin22 / admin123`。租户编号解析需兼容大小写输入，商户密码登录必须校验管理员 `tenant_id` 与请求头 `X-Tenant-ID` 一致，避免账号跨租户登录。

#### 小程序端 - 会员

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/member/auth/login-by-wechat | 微信登录 |
| POST | /api/v1/member/auth/wechat-phone | 微信获取手机号 |
| GET | /api/v1/member/profile | 获取个人资料 |
| PUT | /api/v1/member/profile | 修改个人资料 |
| GET | /api/v1/member/points | 获取积分余额 |
| GET | /api/v1/member/points/logs | 积分明细 |
| GET | /api/v1/member/coupons | 我的优惠券 |
| GET | /api/v1/member/distribution | 分销中心 |
| POST | /api/v1/member/distribution/apply | 申请成为分销员 |
| POST | /api/v1/member/distribution/bind | 绑定邀请人 |
| GET | /api/v1/member/distribution/commissions | 我的佣金记录 |

`POST /api/v1/member/auth/login-by-wechat` 请求体至少包含 `code`，可同时携带 `nickname`、`avatar`、`gender`。响应返回 `{ token, member }`：`member` 为当前租户下匹配 openid 的既有会员，或本次自动创建的新会员。若会员已被商户禁用，登录和会员端鉴权接口均返回认证错误，前端应清理本租户会员会话并引导重新登录或联系商户。该接口使用平台统一小程序 AppID/Secret 调用微信 `code2session`；开发环境下如果平台统一小程序未配置，则使用固定本地 openid 走同一套会员读取/自动注册流程；生产环境缺少平台统一小程序配置时返回“平台微信小程序未配置”。

#### 租户后台 - 小程序码

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/tenant/site/mini-qrcode | 当前租户管理员生成本租户入口小程序码 |

`GET /api/v1/tenant/site/mini-qrcode` 必须使用平台统一小程序 `wechat.app_id` / `wechat.app_secret` 获取 access token，并调用微信 `wxa/getwxacodeunlimit` 生成真实小程序码；不得使用普通二维码或自定义 scheme 模拟。入口页固定为 `pages/home/index`，`scene` 使用 `t={tenant_code}`，小程序启动解析 `scene` 后确定当前租户。响应返回 `image_data_url`、`tenant_code`、`page`、`scene`、`path`、`env_version`、`check_path` 等字段；缺少平台统一小程序配置、AppID 不匹配或微信接口失败时，应返回明确错误，前端不能展示可误扫码的模拟图。小程序码生成环境由 `wechat.mini_qrcode_env_version` / `WECHAT_MINI_QRCODE_ENV_VERSION` 控制，可取 `release`、`trial`、`develop`；路径校验由 `wechat.mini_qrcode_check_path` / `WECHAT_MINI_QRCODE_CHECK_PATH` 控制。本地和测试环境默认 `trial + check_path=false`，用于小程序尚未发布或页面只在体验版时避免微信返回 `41030 invalid page`；生产环境默认 `release + check_path=true`，发布前必须确认 `pages/home/index` 已存在于线上正式版。

会员端分销接口必须位于会员鉴权之后，并受 `distribution` 套餐功能开关保护。`GET /api/v1/member/distribution` 返回 `{ settings, distributor, can_apply, invite_code, bound_parent_id }`，供个人中心展示佣金、申请状态和邀请路径；`POST /api/v1/member/distribution/apply` 幂等创建当前会员的分销员申请；`POST /api/v1/member/distribution/bind` 使用 `inviter_member_id` 绑定上级分销员会员 ID，已绑定后不可重复改绑；`GET /api/v1/member/distribution/commissions` 只返回当前会员作为分销员产生的佣金记录。未开通分销功能时返回套餐功能错误，前端展示“暂未开通”，不能表现为 404；如果店铺配置没有展示分销模块，会员中心不应主动请求分销接口，避免可选功能在控制台反复输出业务 warning。

#### 小程序端 - 商品

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/member/products | 商品列表（分页+筛选） |
| GET | /api/v1/member/products/{id} | 商品详情（含 SKU 列表） |
| GET | /api/v1/member/products/hot | 热门商品（分页） |
| GET | /api/v1/member/products/recommend | 推荐商品（分页） |
| GET | /api/v1/member/categories | 平台统一商品分类列表 |

商品列表查询参数：`page` 默认 1，`size` 默认 20、最大 100，`category_id` 按分类筛选，`keyword` 按商品名称模糊搜索。接口仅返回当前租户已上架商品，响应 `data` 为 `{ list, total, page, size }`。

小程序分类页 `/pages/catalog/index`：默认“全部商品”不传 `category_id`、`keyword` 和推荐/热门模式，首屏加载第 1 页；用户触底时按相同筛选条件递增 `page` 批量追加，直到已加载数量达到 `total`。切换分类、搜索词、热门/推荐筛选时必须重置列表和页码，重新从第 1 页加载；空态仅在第 1 页无数据时展示。

#### 小程序端 - 店铺配置与主题

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/public/tenant/resolve?code={tenant_code} | 解析租户摘要，返回 `id`、`code`、`brand_name`、`brand_theme` 等启动所需字段 |
| GET | /api/v1/member/storefront/config | 获取当前租户小程序店铺配置，返回 `primary_color`、首页模块、会员中心入口等前台配置 |

小程序主题色优先级：`member/storefront/config.primary_color` > `public/tenant/resolve.brand_theme` > 默认色 `#FF6B4A`。主题色兼容 `#RRGGBB`、`#RGB` 和 `rgb(...)` 格式；小程序启动时先解析租户并缓存摘要；进入页面或拉取店铺配置后，将主题色转换为 CSS 变量，应用到页面背景、主按钮、价格、标签、优惠券、会员中心头图、分类筛选、购物车结算按钮和自定义 TabBar 激活态。

产品约束：同一租户内前台所有页面必须保持一致主题；切换租户后必须清理旧租户会员会话并刷新主题；接口失败或主题色格式非法时必须回退默认主题，不能阻断商品浏览、登录、下单等核心流程。

发布约束：小程序正式版不得暴露测试用户登录入口，不得调用 `/member/auth/dev-login`，不得使用 `http://`、公网 IP、localhost 或局域网地址作为 API 基础地址；发布配置必须使用平台统一小程序 AppID，开启微信开发者工具 URL 合法域名检查，并关闭上传 source map。域名备案或 HTTPS 反代未就绪时，可以临时将小程序 API 基础地址切到公网 IP `http://39.96.201.126:18080/api/v1` 并关闭 URL 检查做联调，但提交审核前必须恢复 HTTPS 合法域名。生产 API 域名必须是 HTTPS 合法域名并已配置到微信小程序 request 合法域名，否则发布前视为阻塞项。

会员中心快捷入口使用短路径契约：订单 `/orders`、优惠券 `/coupons`、收货地址 `/addresses`、购物车 `/cart`。其中“收货地址”必须跳转到地址列表页，不允许配置为 `/profile`；小程序端需要对后台配置做路径归一化，避免历史配置或误配置导致点击后停留在会员中心。

会员中心信息结构：会员头像、昵称、会员 ID、积分和成长值摘要放在会员头图内展示；头图下方不再额外展示“会员积分 / 积分记录 / 成长值”三张独立统计卡，避免信息重复和页面纵向占用过多。更完整的资产信息统一放在“会员资产”模块内承载。

设计约束：主题色只作为品牌主色和关键行动色，不覆盖文本层级、卡片白底、危险提示色和成功/失败状态色；由主色派生浅色背景、深色强调、阴影色和渐变色，保证按钮文字始终使用白色并保持可读性。

字体与字号：小程序前台统一使用系统无衬线字体栈，中文优先 `PingFang SC`，英文/数字优先系统 San Francisco，避免页面之间出现不同字体观感。字号采用克制层级：辅助信息 24rpx、正文/按钮 26-28rpx、卡片标题 30-32rpx、页面/运营视觉标题 38-40rpx；常规文本字重 400-500，重点信息 600，少量页面主标题最高 700，避免大面积使用 800/900 造成突兀。价格、按钮、状态标签可以略强调，但必须保持行高充足、不断行挤压、不与图片或卡片边缘冲突。

首页商品卡片：商品卡片必须保持足够信息密度，除图片、商品名和价格外，优先展示商品副标题、已售数量、库存/售罄状态等轻量辅助信息；没有副标题时使用“近期热卖 / 店铺推荐 / 品质好物”等前端兜底文案填充，不让卡体出现大面积空白。卡体布局采用从上到下的信息流，不使用垂直居中撑高；价格和库存状态放在底部同一行，保证用户扫视时能同时看到购买价值和可售状态。

底部导航安全区：首页、分类、购物车、我的等自定义底部导航页面必须使用统一 `tab-page` 根类，页面底部留白需要覆盖导航条高度、渐变背景和 `env(safe-area-inset-bottom)`，确保最后一张卡片、空态按钮和分页加载文案不会被底部导航遮挡。

请求可靠性：小程序请求必须设置明确超时时间，并在失败日志中输出业务路径、耗时和错误原因；启动期租户解析、店铺配置、首页数据存在并发触发场景，店铺配置需要按租户做短周期缓存和 in-flight 去重，避免 App `onLaunch`、`onShow` 与首页加载重复请求导致开发工具出现 `timeout` 噪音。对于未开通套餐功能等可预期业务拒绝，调用方可显式声明静默业务码，避免可选模块反复输出 warning，但仍需保留真实网络失败和非预期业务错误日志。

#### 小程序端 - 订单

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/member/orders | 创建订单 |
| GET | /api/v1/member/orders | 订单列表（含商品明细，支持 `status` 单值或逗号分隔分组） |
| GET | /api/v1/member/orders/{id} | 订单详情（含商品明细） |
| GET | /api/v1/member/orders/{id}/express | 订单物流轨迹 |
| POST | /api/v1/member/orders/{id}/cancel | 取消订单 |
| POST | /api/v1/member/orders/{id}/confirm | 确认收货 |
| GET | /api/v1/member/after-sale-reasons | 获取启用售后原因，支持 `type=refund/return_refund` |
| POST | /api/v1/member/orders/{id}/after-sales | 发起整单售后申请 |
| GET | /api/v1/member/after-sales | 我的售后单列表，支持 `status`、`order_id` |
| GET | /api/v1/member/after-sales/{id} | 我的售后单详情 |
| POST | /api/v1/member/after-sales/{id}/return | 提交退货物流 |
| POST | /api/v1/member/after-sales/{id}/cancel | 取消待处理售后申请 |

小程序“我的订单”必须以会员自己的订单为边界，列表接口需要返回 `items` 商品明细，列表页至少展示订单号、状态、首件商品、总件数、创建时间和实付金额。状态筛选采用分组语义：待支付 `pending_pay`，待发货 `paid,preparing`，待收货 `shipped,delivered`，已完成 `completed`，已取消 `cancelled`；已发货或已送达状态都允许会员在详情页确认收货。

订单主流程为：会员确认订单后创建 `pending_pay` 订单并保留入库；结账页跳转到订单详情，会员可支付或取消。待支付订单从订单列表再次点击也必须进入详情页并展示“支付订单 / 取消订单”动作。支付成功后订单进入 `paid`，商户在管理端执行“开始处理”进入 `preparing`，完成拣货后发货并写入物流公司和运单号进入 `shipped`；会员可在详情页查看物流跟踪，收到货后确认收货，订单进入 `completed` 并结束。开发环境没有真实微信支付配置时，支付接口允许返回并执行开发模拟支付，但生产环境必须通过微信 JSAPI 支付参数和回调推进支付成功状态。

售后首版只支持整单售后。会员提交 `{ type, reason, description, amount, images }`，其中 `reason` 必须来自平台启用的售后原因列表；后端必须校验订单属于当前会员、订单处于可售后状态、当前订单没有未完结售后单，并记录 `order_status_before`。退货退款审核通过后，会员通过 `{ return_express_company, return_express_no }` 提交退货物流；售后完成、驳回或取消都必须写入订单日志和商户订单消息。

小程序结账页如果订单包含实物商品，只展示当前已选择的一条收货地址信息，不在确认订单页展开全部地址列表；如需更换地址，会员点击“选择或管理收货地址”进入地址选择页。由结账页进入地址管理时使用选择模式：地址页必须展示当前已选地址状态，会员点选目标地址后再点击“确认使用该地址”完成更换并返回结账页；新增地址成功后可直接作为当前订单地址并自动返回。如果小程序页面栈导致 `navigateTo` 失败，需要降级为 `redirectTo` 并在地址页选择完成后回到结账页。从个人中心进入地址管理时保持普通管理模式，只新增/查看地址，不强制返回结账。未登录进入地址页时，登录完成后需要回到原地址页模式，不能丢失选择流程。

#### 小程序端 - 营销

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/coupons/available | 可领取优惠券 |
| POST | /api/v1/coupons/{id}/receive | 领取优惠券 |
| GET | /api/v1/coupons/can-use | 可用优惠券（下单时） |
| GET | /api/v1/seckills/active | 当前秒杀场次 |
| GET | /api/v1/seckills/{id}/products | 秒杀商品 |
| POST | /api/v1/seckills/{product_id}/seckill-order | 秒杀下单 |
| GET | /api/v1/group-buys/active | 进行中拼团 |
| POST | /api/v1/group-buys/{id}/join | 参与拼团 |

#### 小程序端 - 支付

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/member/payments | 创建顾客订单支付；前期生产使用平台统一微信商户号 JSAPI 下单，返回 `payment_no` 与小程序支付参数；开发可模拟支付成功 |
| POST | /api/v1/payments/callback/wechat | 顾客订单微信支付回调；按 `payment_no/out_trade_no` 幂等更新 `payments` 和订单状态 |
| POST | /api/v1/public/subscription/callback | 租户订阅微信支付回调；仅处理 `tenant_subscription_orders` |

`POST /api/v1/member/payments` 请求体：`{ order_no }`。响应体：`{ payment_no, pay_params, mock_paid, status }`。生产环境必须校验平台微信支付配置完整，并以当前订单租户作为 `settlement_tenant_id` 写入支付记录，后续账期结算按该字段归集租户应结金额。`pay_params` 只包含小程序端 `wx.requestPayment` 所需字段，不返回平台商户密钥或证书材料。

微信回调契约：顾客订单回调只处理 `payments.pay_scene = member_order`；订阅回调只处理 `tenant_subscription_orders`。两类回调都必须验签、解密、校验金额、校验商户号、校验交易状态，并保证重复通知幂等返回成功。

#### 管理端 - 表格关联信息展示约定

管理端表格中的关联字段不能只展示原始 ID。订单、分销、营销活动、平台审核、域名、部署、短信和 API 凭据等页面遇到 `member_id`、`tenant_id`、`product_id`、`token_id`、`parent_id`、`buyer_id`、`distributor_id` 等关联字段时，前端需要通过已有列表或详情接口加载对应资源，并在表格主内容中展示业务可读信息：会员展示昵称/手机号/等级，租户展示公司名称/编号/联系人，商品展示商品名/价格，API Token 展示凭据名称/AppKey。ID 可以作为次要辅助信息保留，用于排查和精确定位；如果关联资源已被删除或当前接口无法获取，页面展示“未找到 #ID”，不能只裸露数字让用户猜含义。

#### 管理端 - 商品

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/admin/products | 创建商品 |
| PUT | /api/v1/admin/products/{id} | 修改商品 |
| PUT | /api/v1/admin/products/{id}/status | 上架/下架 |
| DELETE | /api/v1/admin/products/{id} | 删除（软删除） |
| POST | /api/v1/admin/products/{id}/skus | 创建SKU |
| PUT | /api/v1/admin/skus/{id} | 修改SKU |
| POST | /api/v1/admin/categories | 创建分类 |
| PUT | /api/v1/admin/categories/{id} | 修改分类 |

#### 管理端 - 订单

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/orders | 订单列表（多条件） |
| POST | /api/v1/admin/orders/{id}/prepare | 开始处理订单（paid → preparing） |
| POST | /api/v1/admin/orders/{id}/ship | 发货 |
| GET | /api/v1/admin/after-sales | 售后单列表，支持 `status`、`order_id`、`member_id` |
| GET | /api/v1/admin/after-sales/{id} | 售后单详情 |
| POST | /api/v1/admin/after-sales/{id}/approve | 审核通过售后申请 |
| POST | /api/v1/admin/after-sales/{id}/reject | 驳回售后申请 |
| POST | /api/v1/admin/after-sales/{id}/receive | 确认收到退货 |
| POST | /api/v1/admin/after-sales/{id}/refund | 标记退款完成 |
| GET | /api/v1/admin/orders/export | 导出Excel |

商户售后审核必须以当前租户为边界。驳回和取消需恢复订单原状态；退款完成前不能把订单从 `refunding` 推进到其他主流程状态。真实微信退款接入前，`refund` 接口只更新业务状态并保留操作日志，后续必须接入微信支付退款 API、退款回调、金额校验、幂等退款单号和积分/分佣回滚。

#### 管理端 - 会员

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/members | 当前租户会员列表，支持 `page`、`size`、`keyword`、`status`、`level_id` 筛选 |
| GET | /api/v1/admin/members/{id} | 当前租户会员详情，返回基础资料、收货地址和最近积分明细 |
| PATCH | /api/v1/admin/members/{id}/status | 启用/禁用当前租户会员，`status` 取值 `1`/`0` |
| PATCH | /api/v1/admin/members/{id}/level | 调整或清空当前租户会员等级，需开通 `member_level` 功能 |

会员管理接口必须经过租户中间件和管理员鉴权，Repository 查询必须自动携带或显式携带当前 `tenant_id` 条件。列表响应统一为 `{ list, total, page, size }`；会员等级仅作为可选增强能力，不影响基础会员管理流程。

#### 管理端 - 结算资料

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/settings/payment | 获取当前租户结算资料配置 |
| PUT | /api/v1/admin/settings/payment | 提交当前租户结算资料，提交后进入平台审核 |

商户侧 `PUT /api/v1/admin/settings/payment` 请求体前期支持 `provider=manual_settlement`，并提交 `settlement_account_name`、`settlement_account_no`、`settlement_bank_name`、`settlement_remark`；现有 `mch_id`、`app_id`、`sub_mchid`、`sub_appid`、`api_v3_key`、`cert_serial_no`、`private_key_pem`、`cert_pem` 作为历史直连或服务商兼容字段保留。提交成功后 `audit_status` 重置为待审核、`enabled` 重置为 0。该配置不再作为 C 端支付下单前置条件，只影响平台账期结算能否放款。

#### 管理端 - 优惠券/拼团/秒杀/分销（CRUD + 统计）

#### 平台运营

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/platform/dashboard | 运营仪表盘 |
| GET | /api/v1/platform/plans | 套餐列表 |
| POST | /api/v1/platform/plans | 创建套餐 |
| PUT | /api/v1/platform/plans/{id} | 修改套餐 |
| GET | /api/v1/platform/settings | 获取平台全局设置与平台微信支付配置 |
| PUT | /api/v1/platform/settings | 更新平台基础信息和平台统一收款配置 |
| GET | /api/v1/platform/payment-configs | 查询租户结算资料审核列表 |
| POST | /api/v1/platform/payment-configs/{id}/audit | 审核租户结算资料 |
| GET | /api/v1/platform/after-sale-reasons | 售后原因列表 |
| POST | /api/v1/platform/after-sale-reasons | 新增售后原因 |
| PUT | /api/v1/platform/after-sale-reasons/{id} | 修改售后原因 |
| PATCH | /api/v1/platform/after-sale-reasons/{id}/enabled | 启用或停用售后原因 |
| DELETE | /api/v1/platform/after-sale-reasons/{id} | 删除未使用的售后原因 |
| GET | /api/v1/platform/bills | 账单列表 |
| GET | /api/v1/platform/bills/revenue | 收入报表 |
| GET | /api/v1/platform/sms/settings | 获取平台自身短信网关配置（tenant_id=0） |
| PUT | /api/v1/platform/sms/settings | 更新平台自身短信网关配置（tenant_id=0） |
| GET | /api/v1/platform/sms/templates | 查询平台功能绑定的阿里云 TemplateCode（tenant_id=0） |
| POST | /api/v1/platform/sms/templates | 为固定平台功能首次保存阿里云 TemplateCode |
| PUT | /api/v1/platform/sms/templates/{id} | 更换固定平台功能绑定的阿里云 TemplateCode |
| DELETE | /api/v1/platform/sms/templates/{id} | 删除平台自身短信模板 |
| GET | /api/v1/platform/sms/logs | 查询平台自身短信发送日志（tenant_id=0） |

验证码发送成功后，前端只提示“验证码已发送”等固定文案；即使本地联调接口返回 `dev_code`，也不得把验证码展示在 toast / message / notification 等临时提示中。

平台侧 `PUT /api/v1/platform/settings` 请求体除平台名称、Logo、客服信息外，`wxpay_*` 字段作为平台统一微信收款配置，前期同时用于租户套餐订阅付款和顾客订单付款。`sp_appid`、`sp_mchid`、`sp_apiv3_key`、`sp_cert_serial`、`partner_notify_url` 作为后续服务商模式预留字段，前期不参与顾客订单支付门禁。

平台审核租户结算资料时，审核通过才允许 `enabled=1`。审核列表必须展示租户、结算方式、历史商户号或子商户号兼容信息、备注、审核状态和提交时间，便于平台排查结算资料来源。

平台售后原因管理用于统一小程序售后申请下拉项。原因配置变更即时影响新申请，不回写历史售后单；历史售后单保留申请时的 `reason` 文本快照。

### 4.3 错误码规范

| 区间 | 类别 |
|------|------|
| 1xxxx | 认证/权限错误 |
| 2xxxx | 参数校验错误 |
| 3xxxx | 业务逻辑错误（30001库存不足, 30002余额不足, 30003套餐到期, 30004功能未开通） |
| 4xxxx | 微信API错误 |
| 5xxxx | 系统内部错误 |

---

## 五、微信生态对接方案

### 5.1 微信登录流程

```
1. 小程序调用 wx.login() → 获取临时 code
2. 我们的 API 调用微信 code2session
   URL: https://api.weixin.qq.com/sns/jscode2session
   参数: appid + secret + js_code + grant_type=authorization_code
   返回: openid + session_key（解密手机号用）
3. 查询或创建会员（openid 匹配）
4. JWT 签发：{member_id, tenant_id, openid, exp}
5. 手机号解密：session_key（AES-256-CBC）解密 encryptedData
```

### 5.2 微信支付（平台统一收款 / JSAPI）

```
1. 平台订阅付款：统一下单 POST https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi
   Header: Authorization: WECHATPAY2-SHA256-RSA2048（证书签名）
   Body: {appid, mchid, description, out_trade_no, notify_url, amount:{total}, payer:{openid}}
   - appid/mchid 使用平台自有收款配置
   - out_trade_no 对应 tenant_subscription_orders.order_no
   - 回调 POST /api/v1/public/subscription/callback

2. 顾客订单付款：统一下单 POST https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi
   Header: Authorization: WECHATPAY2-SHA256-RSA2048（平台商户证书签名）
   Body: {appid, mchid, description, out_trade_no, notify_url, amount:{total}, payer:{openid}}
   - appid/mchid 使用平台统一收款配置
   - 订单所属租户写入 payments.tenant_id / payments.settlement_tenant_id，作为后续账期结算归属
   - out_trade_no 对应 payments.payment_no
   - 回调 POST /api/v1/payments/callback/wechat

3. 拿 prepay_id，生成调起小程序支付参数
4. 回调处理
   - 验证签名（微信使用 AES-256-GCM 加密）
   - 校验 out_trade_no、金额、平台 mchid、trade_state
   - 解密后处理：
     更新payment状态 → 更新order状态 → 扣库存 → 发积分 → 计算分佣
   - 返回 HTTP 200 + {"code":"SUCCESS","message":"SUCCESS"}
5. 本地开发环境无真实微信支付配置时，`POST /api/v1/member/payments` 可创建支付记录并模拟支付成功，用于验证“待支付 → 待发货 → 处理中 → 已发货 → 已完成”的页面和接口链路；该能力不得作为生产支付替代。
```

上线门禁：平台统一收款模式未完成真实签名、验签、回调解密、金额校验和平台商户号校验前，不得开启生产顾客订单支付。租户结算资料未审核通过时不阻塞顾客支付，但账期结算必须把该租户标记为“结算资料未完成”。

### 5.3 小程序多租户方案

**推荐：模板小程序模式（平台统一收款）**

- 平台统一注册"半自动小程序"
- 每个租户通过 ext.json 隔离配置
- 租户上传 logo/店名后自动替换
- 云开发或静态托管切换数据源

### 5.4 小程序按钮布局约定

- 空态引导、登录、资料保存、地址保存等单一主行动应居中或满宽显示，强化当前页面的下一步，不统一靠右。
- 结算栏、订单详情等包含金额或多动作的区域，按钮跟随内容上下文做行内布局；主按钮保留清晰视觉权重，次要按钮与主按钮组成等宽或紧凑操作组。
- 搜索、领取优惠券、数量加减等短动作属于局部工具操作，可保留紧凑按钮并贴近对应输入或数据项。
- 全局样式只定义按钮外观，不强制所有卡片按钮右对齐；具体页面按业务场景决定宽度、居中、行内或操作组布局。

---

## 六、订阅计费设计

### 6.1 三档套餐

| 套餐 | 月费 | 年费 | 商品上限 | 月订单上限 | 会员上限 | 核心功能 |
|------|------|------|---------|-----------|---------|---------|
| 基础版 | 299元 | 2990元 | 100 | 500 | 1000 | 多规格SKU、优惠券、积分 |
| 专业版 | 799元 | 7990元 | 2000 | 10000 | 50000 | 秒杀、拼团、分销、会员等级、自定义域名 |
| 旗舰版 | 1999元 | 19990元 | 无限制 | 无限制 | 无限制 | 全部功能、开放API、白标、私有化部署 |

### 6.2 用量校验

- 每次商品创建前：检查 product_count < plan.product_limit
- 每月订单创建前：检查 month_order_count < plan.order_limit
- 所有功能操作前：检查 feature 在 plan.features 范围内

### 6.3 套餐生命周期

- **到期前**：7天/3天/1天 发送预警通知
- **宽限期**：到期后7天（status=2欠费，仍可登录管理后台查看）
- **封禁**：到期7天后自动封禁（status=3，API返回套餐到期错误）
- **续费**：延长 plan_expire_at，创建 plan_log
- **升级**：立即生效，按剩余时间折算差价
- **降级**：下个账期生效

---

## 七、工程化规范

### 分层原则

```
Handler → Service → Repository
- 禁止跨层调用（Handler不能直接访问Repository）
- 禁止下层调用上层
- 依赖注入使用 Google Wire
```

### 日志规范

```
- 使用 Zap 结构化日志
- 格式：{"level":"info","ts":"...","caller":"...","msg":"...","tenant_id":1,...}
- 禁止使用 fmt.Print / log.*
```

### 数据库规范

```
- 所有表必须有 tenant_id（平台数据除外）
- 软删除优先（deleted_at 字段）
- 时间字段统一 DATETIME
- 金额字段统一 DECIMAL(10,2)
- 金额计算在应用层用 decimal.Decimal（避免浮点精度问题）
```

### 并发安全

```
- 库存操作使用 Redis 分布式锁（SETNX + TTL）
- 订单号使用 Snowflake ID
- 关键操作（支付、退款）必须防重复（幂等）
```

### 配置文件

```
- 全用 YAML（config.yaml）
- 不在代码中硬编码任何值
- 环境变量覆盖：DEV/TEST/PROD 三环境
```
