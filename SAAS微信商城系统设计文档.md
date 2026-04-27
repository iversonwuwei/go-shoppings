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
└── docker-compose.yml            # MySQL + Redis
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

#### member_levels（会员等级表）

含 tenant_id，min_growth（成长值门槛）、discount_rate（折扣率）、points_mult（积分倍数）

#### points_logs（积分变动记录）

tenant_id + member_id，含 change_type(order/gift/sign/refund/manual)、change_value、balance_before/after

### 3.4 订单层表

#### orders（订单主表）

tenant_id + member_id + order_no(唯一)

状态流：pending_pay → paid → preparing → shipped → delivered → completed
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

#### 小程序端 - 订单

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/v1/orders | 创建订单 |
| GET | /api/v1/orders | 订单列表 |
| GET | /api/v1/orders/{id}/detail | 订单详情（含商品明细） |
| POST | /api/v1/orders/{id}/cancel | 取消订单 |
| POST | /api/v1/orders/{id}/confirm | 确认收货 |
| POST | /api/v1/orders/{id}/apply-refund | 申请退款 |
| GET | /api/v1/orders/{id}/express | 物流查询 |

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
| POST | /api/v1/payments/create | 创建支付（返回支付参数） |
| POST | /api/v1/payments/query/{payment_no} | 查询支付状态 |
| POST | /api/v1/payments/callback/wechat | 微信支付回调 |

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
| PUT | /api/v1/admin/orders/{id}/status | 修改状态 |
| POST | /api/v1/admin/orders/{id}/ship | 发货 |
| POST | /api/v1/admin/orders/{id}/refund | 同意/拒绝退款 |
| GET | /api/v1/admin/orders/export | 导出Excel |

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
