# SaaS 微信商城系统 - 完整设计文档

> 技术栈：Go + Gin + GORM + MySQL + Redis + Docker
> 多租户模式：行级隔离（tenant_id）
> 套餐订阅：基础版 / 专业版 / 旗舰版

---

## 一、项目概述

### 1.1 系统定位

面向企业提供微信小程序商城 SaaS 服务，支持多租户入驻、订阅计费、微信生态深度对接。

### 1.2 技术选型

| 层级 | 技术选型 | 说明 |
|------|---------|------|
| 语言 | Go 1.21+ | 高并发、低内存、编译型二进制 |
| Web框架 | Gin 1.9+ | 性能优秀、中间件生态成熟 |
| ORM | GORM | Go ORM 主流选择 |
| 数据库 | PostgreSQL 15 | 事务支持、多租户行级隔离、JSONB原生支持 |
| 缓存 | Redis 7.0 | 租户配置缓存、分布式锁、库存扣减 |
| 容器 | Docker + docker-compose | 本地开发与生产部署 |
| 微信SDK | 原生实现（复用微信官方API） | 无第三方依赖 |

### 1.3 项目结构

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
| wechat_appid | VARCHAR(50) | 微信小程序 AppID |
| wechat_secret | VARCHAR(100) | 微信小程序 AppSecret（加密） |
| wechat_mchid | VARCHAR(30) | 微信商户号 |
| wechat_apiv3_key | VARCHAR(100) | APIv3 密钥（加密） |
| wechat_cert_serial | VARCHAR(100) | 支付证书序列号 |
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

本地开发环境（`app.env != production`）允许租户暂未配置微信小程序 AppID/Secret：`login-by-wechat` 会降级为开发微信会员登录，使用当前租户内固定的本地 openid 创建或读取会员，便于开发者工具调试登录后流程。生产环境必须配置租户微信小程序，不能启用该降级路径。

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

短信配置、模板和发送日志分表存储，覆盖验证码、订单通知、套餐到期提醒等流程。

#### api_tokens / api_request_logs（开放 API 表）

开放 API 功能需要 `api_tokens` 存储租户密钥、权限范围和状态，`api_request_logs` 记录调用路径、状态码、耗时和来源 IP。

### 3.6 支付层表

#### payments（支付记录表）

tenant_id + member_id，payment_no(唯一)，含 wechat_transaction_id、refund_amount/refund_status

#### tenant_subscription_orders（租户订阅订单表）

租户续费、升级、降级前先创建订阅订单，记录 plan_id、billing_cycle、amount、order_no、支付流水、支付前后到期时间。微信支付回调成功后再更新租户套餐和 `tenant_plan_logs`。

### 3.7 系统层表

#### admin_action_logs（操作日志）

含 tenant_id（平台管理员为0）、admin_id、action、target_*、request_*、request_ip、user_agent

#### uploads（文件记录）

含 tenant_id、storage_type(local/oss/cos)、storage_url

#### delivery_settings（配送设置表）

每个租户一行，覆盖快递、同城配送、自提三类配送能力的开关、费用、半径、门店地址和联系电话。

### 3.8 数据库脚本覆盖要求

- `scripts/init_db.sql` 必须覆盖基础表和当前运行态模型所需表；后续增量统一放入 `scripts/migrations/`，并保持 `CREATE TABLE IF NOT EXISTS`、`ALTER TABLE ... ADD COLUMN IF NOT EXISTS`、`CREATE INDEX IF NOT EXISTS` 的幂等写法。
- 本地演示库必须至少包含：`api_tokens`、`api_request_logs`、`sms_settings`、`sms_templates`、`sms_logs`、`distribution_settings`、`distributors`、`commission_logs`、`groupon_activities`、`groupons`、`groupon_members`、`points_settings`、`delivery_settings`、`tenant_subscription_orders`。
- 租户表必须包含 `billing_cycle` 和 `extra_features`，用于入驻计费周期和平台额外授权功能。缺失字段或缺失表应通过可重复执行的补丁脚本修复，并在执行后用 information_schema 校验。

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

`POST /api/v1/member/auth/login-by-wechat` 请求体至少包含 `code`，可同时携带 `nickname`、`avatar`、`gender`。响应返回 `{ token, member }`：`member` 为当前租户下匹配 openid 的既有会员，或本次自动创建的新会员。若会员已被商户禁用，登录和会员端鉴权接口均返回认证错误，前端应清理本租户会员会话并引导重新登录或联系商户。开发环境下，如果当前租户未配置微信小程序 AppID/Secret，该接口不调用微信 `code2session`，而是使用固定本地 openid 走同一套会员读取/自动注册流程；生产环境仍返回“租户未配置微信小程序”。

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

会员中心快捷入口使用短路径契约：订单 `/orders`、优惠券 `/coupons`、收货地址 `/addresses`、购物车 `/cart`。其中“收货地址”必须跳转到地址列表页，不允许配置为 `/profile`；小程序端需要对后台配置做路径归一化，避免历史配置或误配置导致点击后停留在会员中心。

设计约束：主题色只作为品牌主色和关键行动色，不覆盖文本层级、卡片白底、危险提示色和成功/失败状态色；由主色派生浅色背景、深色强调、阴影色和渐变色，保证按钮文字始终使用白色并保持可读性。

字体与字号：小程序前台统一使用系统无衬线字体栈，中文优先 `PingFang SC`，英文/数字优先系统 San Francisco，避免页面之间出现不同字体观感。字号采用克制层级：辅助信息 24rpx、正文/按钮 26-28rpx、卡片标题 30-32rpx、页面/运营视觉标题 38-40rpx；常规文本字重 400-500，重点信息 600，少量页面主标题最高 700，避免大面积使用 800/900 造成突兀。价格、按钮、状态标签可以略强调，但必须保持行高充足、不断行挤压、不与图片或卡片边缘冲突。

首页商品卡片：商品卡片必须保持足够信息密度，除图片、商品名和价格外，优先展示商品副标题、已售数量、库存/售罄状态等轻量辅助信息；没有副标题时使用“近期热卖 / 店铺推荐 / 品质好物”等前端兜底文案填充，不让卡体出现大面积空白。卡体布局采用从上到下的信息流，不使用垂直居中撑高；价格和库存状态放在底部同一行，保证用户扫视时能同时看到购买价值和可售状态。

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

小程序“我的订单”必须以会员自己的订单为边界，列表接口需要返回 `items` 商品明细，列表页至少展示订单号、状态、首件商品、总件数、创建时间和实付金额。状态筛选采用分组语义：待支付 `pending_pay`，待发货 `paid,preparing`，待收货 `shipped,delivered`，已完成 `completed`，已取消 `cancelled`；已发货或已送达状态都允许会员在详情页确认收货。

订单主流程为：会员确认订单后创建 `pending_pay` 订单并保留入库；结账页跳转到订单详情，会员可支付或取消。待支付订单从订单列表再次点击也必须进入详情页并展示“支付订单 / 取消订单”动作。支付成功后订单进入 `paid`，商户在管理端执行“开始处理”进入 `preparing`，完成拣货后发货并写入物流公司和运单号进入 `shipped`；会员可在详情页查看物流跟踪，收到货后确认收货，订单进入 `completed` 并结束。开发环境没有真实微信支付配置时，支付接口允许返回并执行开发模拟支付，但生产环境必须通过微信 JSAPI 支付参数和回调推进支付成功状态。

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
| POST | /api/v1/member/payments | 创建会员订单支付（生产返回 JSAPI 支付参数；开发可模拟支付成功） |
| POST | /api/v1/payments/callback/wechat | 微信支付回调 |

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
| POST | /api/v1/admin/orders/{id}/refund | 同意/拒绝退款 |
| GET | /api/v1/admin/orders/export | 导出Excel |

#### 管理端 - 会员

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/admin/members | 当前租户会员列表，支持 `page`、`size`、`keyword`、`status`、`level_id` 筛选 |
| GET | /api/v1/admin/members/{id} | 当前租户会员详情，返回基础资料、收货地址和最近积分明细 |
| PATCH | /api/v1/admin/members/{id}/status | 启用/禁用当前租户会员，`status` 取值 `1`/`0` |
| PATCH | /api/v1/admin/members/{id}/level | 调整或清空当前租户会员等级，需开通 `member_level` 功能 |

会员管理接口必须经过租户中间件和管理员鉴权，Repository 查询必须自动携带或显式携带当前 `tenant_id` 条件。列表响应统一为 `{ list, total, page, size }`；会员等级仅作为可选增强能力，不影响基础会员管理流程。

#### 管理端 - 优惠券/拼团/秒杀/分销（CRUD + 统计）

#### 平台运营

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/platform/dashboard | 运营仪表盘 |
| GET | /api/v1/platform/plans | 套餐列表 |
| POST | /api/v1/platform/plans | 创建套餐 |
| PUT | /api/v1/platform/plans/{id} | 修改套餐 |
| GET | /api/v1/platform/bills | 账单列表 |
| GET | /api/v1/platform/bills/revenue | 收入报表 |

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

### 5.2 微信支付（JSAPI）

```
1. 统一下单 POST https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi
   Header: Authorization: WECHATPAY2-SHA256-RSA2048（证书签名）
   Body: {appid, mchid, description, out_trade_no, notify_url, amount:{total}, payer:{openid}}
2. 拿 prepay_id，生成调起小程序支付参数
3. 回调 POST /api/v1/payments/callback/wechat
   - 验证签名（微信使用 AES-256-GCM 加密）
   - 解密后处理：
     更新payment状态 → 更新order状态 → 扣库存 → 发积分 → 计算分佣
   - 返回 HTTP 200 + {"code":"SUCCESS","message":"SUCCESS"}
4. 本地开发环境无真实微信支付配置时，`POST /api/v1/member/payments` 可创建支付记录并模拟支付成功，用于验证“待支付 → 待发货 → 处理中 → 已发货 → 已完成”的页面和接口链路；该能力不得作为生产支付替代。
```

### 5.3 小程序多租户方案

**推荐：模板小程序模式（服务商模式）**

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
