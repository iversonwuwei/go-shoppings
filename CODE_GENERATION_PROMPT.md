# SaaS 微信商城 - AI代码生成 Prompt

```
# SaaS 微信商城系统 - 完整代码生成 Prompt

## 项目概述

基于 Go + Gin + GORM 构建一套 SaaS 版微信多商户商城系统。
- 多租户隔离：行级隔离（tenant_id），中间件注入
- 套餐订阅计费：基础版/专业版/旗舰版三档
- 微信生态：微信小程序（独立/模板）、微信支付（服务商模式）、微信登录
- 支付账户归属：租户订阅付款进入平台账户；顾客商品订单付款必须通过服务商模式结算到当前租户子商户账户
- 技术栈：Go 1.21+ / Gin 1.9+ / GORM / **PostgreSQL 15** / Redis 7.0 / Docker

## 一、项目结构

```
wechat-mall-saas/
├── cmd/
│   ├── api/main.go          # HTTP API 服务入口
│   └── worker/main.go       # 异步任务/Cron 入口
├── internal/
│   ├── handler/             # 控制器层（HTTP Handler）
│   │   ├── middleware/      # 中间件（租户/JWT/CORS/限流/日志）
│   │   ├── admin/          # 管理端接口
│   │   ├── member/          # 小程序端接口
│   │   └── payment.go      # 支付回调
│   ├── service/             # 业务逻辑层
│   ├── repository/         # 数据访问层
│   ├── model/              # 数据模型（带 GORM tags）
│   └── pkg/                # 公共工具包（wxapp/wxpay/jwt/cache）
├── configs/
│   └── config.yaml          # 配置文件
├── scripts/
│   ├── init_db.sql         # 建表 SQL
│   └── seed.sql            # 初始数据
├── Makefile
├── go.mod
├── docker-compose.infra.yml # PostgreSQL 15 + Redis 7 + MinIO
├── docker-compose.app.yml  # API + AI 图片服务
```

### 数据库设计（必须全部实现，Go模型使用 GORM + PostgreSQL tags）

- 使用 `gorm:"column:xxx"` 显式指定列名
- ID统一用 `BIGINT` → Go模型用 `uint64`，GORM tag: `gorm:"primaryKey;autoIncrement"`
- 时间统一用 `TIMESTAMP` → Go模型用 `*time.Time`
- 金额统一用 `NUMERIC(10,2)` → Go模型用 `decimal.Decimal`（github.com/shopspring/decimal）
- JSON字段用 `JSONB` → Go模型用 `pq.StringArray` 或 `[]byte`，GORM tag: `gorm:"type:jsonb"`
- 软删除用 `deleted_at TIMESTAMP` → GORM tag: `gorm:"index;default:null"`
- 唯一索引: `UNIQUE ("tenant_id", "sku_code")` → GORM tag: `gorm:"uniqueIndex:uk_tenant_sku"`

**plans（套餐表）**
- id, name, code(varchar30,unique), monthly_fee, yearly_fee
- product_limit, order_limit, user_limit（0=无限制）
- features（JSON数组：coupon/seckill/group_buy/distribution等）
- is_default, status, created_at, updated_at

**tenants（租户表）**
- id, code(varchar30,unique), company_name, contact_name, contact_phone, contact_email
- wechat_appid, wechat_secret（加密）
- wechat_mchid, wechat_apiv3_key（加密）, wechat_cert_serial 为历史直连商户号兼容字段；顾客订单生产支付不得依赖这些字段
- plan_id（外键plans）, plan_expire_at, status（0待审核/1正常/2欠费/3封禁）
- brand_name, brand_logo, brand_theme, brand_domain（品牌定制）
- billing_cycle（monthly/yearly）, extra_features（JSONB，平台额外授予功能）
- created_at, updated_at

**platform_settings（平台全局设置）**
- 平台订阅收款：wxpay_app_id, wxpay_mch_id, wxpay_apiv3_key, wxpay_cert_serial, wxpay_notify_url
- 服务商配置：sp_appid, sp_mchid, sp_apiv3_key, sp_cert_serial, partner_notify_url
- 平台订阅付款只更新 tenant_subscription_orders；顾客订单付款只更新 payments/orders

**tenant_payment_configs（租户子商户收款配置）**
- id, tenant_id, provider(wechat), enabled, audit_status, audit_remark, submitted_at, audited_at
- sub_mchid, sub_appid 为顾客订单生产支付使用的子商户配置
- sp_mchid, sp_appid 可从平台服务商配置冗余快照，便于回调校验和审计
- mch_id, app_id, api_v3_key, cert_serial_no, private_key_pem, cert_pem 为历史直连兼容字段，新增生产链路优先不用

**tenant_plan_logs（套餐变更记录）**
- id, tenant_id, old_plan_id, new_plan_id, change_type（create/renew/upgrade/downgrade）
- effective_at, expire_at, amount, created_at

### 数据库流程表覆盖要求

- 初始化脚本和迁移脚本必须覆盖运行态 Go 模型中声明的表，避免接口流程进入后才暴露 `relation does not exist`。
- 增量建表必须幂等：`CREATE TABLE IF NOT EXISTS`、`ALTER TABLE ... ADD COLUMN IF NOT EXISTS`、`CREATE INDEX IF NOT EXISTS`。
- 当前必须覆盖的增量流程表：`api_tokens`、`api_request_logs`、`sms_settings`、`sms_templates`、`sms_logs`、`tenant_payment_configs`、`distribution_settings`、`distributors`、`commission_logs`、`groupon_activities`、`groupons`、`groupon_members`、`points_settings`、`delivery_settings`、`tenant_subscription_orders`。
- 当前必须覆盖的租户扩展字段：`tenants.billing_cycle`、`tenants.extra_features`。
- 服务商支付迁移必须补齐 `tenant_payment_configs.sp_mchid/sp_appid/sub_mchid/sub_appid`，以及 `payments.sp_mchid/sp_appid/sub_mchid/sub_appid/settlement_tenant_id/pay_scene`，并提供 information_schema 校验 SQL。

### 商城业务层表

**products（商品主表）** - 所有字段加 tenant_id
- id, tenant_id, category_id（外键product_categories）, name, subtitle, cover_image, images（JSON数组）
- video_url, description（富文本HTML）
- price, cost_price, stock, stock_warning, has_sku（TINYINT）
- delivery_type（JSON数组：express/city/self_pickup）, delivery_fee
- status（0下架/1上架/2售罄）, is_recommend, is_hot
- seo_title, seo_keywords, seo_description, sort, sold_count, view_count
- deleted_at（软删除）, created_at, updated_at
- 索引：idx_tenant, idx_category, idx_status, idx_sold_count DESC, FULLTEXT(name)
- 外键：tenant_id→tenants(id), category_id→product_categories(id)

**product_categories（分类表）** - tenant_id
- id, tenant_id, parent_id（0=一级）, name, icon, cover_image, sort, status, created_at, updated_at

**product_skus（SKU表）** - tenant_id + product_id
- id, tenant_id, product_id, sku_code(varchar50,unique:tenant+code)
- attributes（JSON：[{"attr_id":1,"attr_name":"颜色","value_id":1,"value_name":"红色"}]）
- price, cost_price, stock, image, weight, volume, status, created_at, updated_at

**product_attributes / product_attribute_values（规格属性/值表）** - tenant_id

**members（会员表）** - tenant_id
- id, tenant_id, openid, unionid, session_key（加密存储）
- nickname, avatar, gender, birthday, phone（unique:tenant+phone）
- level_id（外键member_levels）, level_expire_at, points, growth_value
- parent_id（推荐人）, level1_count, level2_count, status
- last_login_at, created_at, updated_at, deleted_at（软删除）
- 唯一索引：(tenant_id, openid), (tenant_id, phone)

**member_levels（会员等级表）** - tenant_id
- id, tenant_id, name, icon, color, min_growth, discount_rate, points_mult, sort, created_at

**points_logs（积分变动记录）** - tenant_id + member_id
- id, tenant_id, member_id, change_type（order/gift/sign/refund/manual）, change_value, balance_before, balance_after
- source_id, source_desc, remark, operator_id, created_at

**orders（订单主表）** - tenant_id + member_id
- id, tenant_id, order_no(varchar32,unique), member_id
- total_amount, delivery_fee, discount_amount, coupon_id, points_discount, actual_amount, cost_amount
- status（pending_pay待支付/paid已支付/preparing备货中/shipped已发货/delivered已收货/completed已完成/cancelled已取消/refunding退款中/refunded已退款）
- receiver_name/phone/province/city/district/address/postcode, delivery_type
- express_company, express_no, self_pickup_code, self_pickup_address
- buyer_remark, distribution_status
- paid_at, shipped_at, delivered_at, completed_at, cancelled_at, expired_at
- created_at, updated_at, deleted_at（软删除）
- 索引：idx_tenant, idx_member, idx_status, idx_order_no, idx_paid_at, idx_created

**order_items（订单商品明细）** - tenant_id + order_id
- id, tenant_id, order_id, product_id, sku_id
- product_name, sku_desc, cover_image（下单时快照）
- price, quantity, item_total（下单时快照）
- refund_status（none/partial/full）, refund_amount
- created_at

**order_logs（订单操作日志）** - tenant_id + order_id
- id, tenant_id, order_id, operator_type（system/member/admin）, operator_id
- action（create/paid/shipped/delivered/cancelled/refund）, before_status, after_status, remark, created_at

**coupons（优惠券表）** - tenant_id
- id, tenant_id, name, type（cash满减/discount折扣/shipping包邮）
- threshold_amount, discount_value, max_discount（折扣上限）
- total_count, remain_count, per_limit
- receive_start_at, receive_end_at, valid_start_at, valid_end_at, valid_days（互斥）
- applicable_type（all全场/product指定商品/category指定分类）, applicable_ids（JSON）, member_levels（JSON）
- status, created_at, updated_at

**member_coupons（会员优惠券记录）** - tenant_id + member_id
- id, tenant_id, member_id, coupon_id
- coupon_name, coupon_type, threshold_amount, discount_value, max_discount（下单时快照）
- received_at, valid_start_at, valid_end_at, used_at, used_order_id
- status（unused/used/expired）, created_at
- 唯一索引：(member_id, coupon_id, received_at)

**group_buys（拼团活动表）** - tenant_id
- id, tenant_id, product_id, sku_id, group_price
- needed_people, group_valid_hours, total_stock, per_person_limit
- start_at, end_at, status, created_at

**group_buy_orders（拼团记录表）** - tenant_id
- id, tenant_id, group_buy_id, order_id, leader_id, needed_people, joined_people
- status（ongoing/success/failed/cancelled）, expire_at, success_at, created_at

**seckills / seckill_products（秒杀）** - tenant_id

**distribution_relations（分销关系）** - tenant_id
- id, tenant_id, member_id, parent_id, level（1/2）, created_at
- 唯一索引：member_id

**distribution_commissions（分佣记录）** - tenant_id
- id, tenant_id, order_id, order_item_id, buyer_id, agent_id, level
- commission_rate, commission_amount, status（pending/settled/withdrawn）, settled_at, created_at

**payments（支付记录表）** - tenant_id
- id, tenant_id, payment_no(varchar32,unique), order_no, member_id
- amount, status（pending/success/failed/closed）, pay_scene(member_order)
- wechat_trade_type, wechat_transaction_id, wechat_payer_openid, wechat_paid_at
- 服务商字段：sp_appid, sp_mchid, sub_appid, sub_mchid, settlement_tenant_id
- refund_amount, refund_status（none/partial/full）
- closed_at, close_reason, expire_at, created_at, updated_at
- 索引：idx_tenant, idx_order, idx_wechat_transaction_id, idx_status

**after_sale_orders（售后单表）** - tenant_id
- id, tenant_id, after_sale_no(unique), order_id, order_no, order_item_id, member_id
- type(refund/return_refund), status(pending/approved/rejected/returning/received/refunded/cancelled)
- amount, reason, description, images(JSONB), order_status_before
- audit_remark, return_express_company, return_express_no, refund_no
- applied_at, audited_at, returned_at, received_at, refunded_at, cancelled_at, created_at, updated_at
- 首版仅支持整单售后；真实微信退款接入前，商户“退款完成”只代表业务状态确认。

**after_sale_reasons（售后原因配置表）**
- id, code(unique), label, type(all/refund/return_refund), sort_order, enabled, created_at, updated_at
- 平台统一维护；小程序只能从启用原因下拉选择，提交 reason 文本快照。
- 初始化默认原因：不想要了、拍错/多拍、商品破损、商品与描述不符、少件/漏发、质量问题、协商一致退款。

**admin_action_logs（操作日志）**
- id, tenant_id, admin_id, admin_username, action, target_type, target_id, target_desc
- request_method, request_path, request_body, request_ip, user_agent, created_at

**uploads（文件记录）** - tenant_id
- id, tenant_id, file_key, original_name, file_size, file_type, file_ext
- storage_type（local/oss/cos）, storage_url, uploaded_by, created_at

## 三、多租户中间件实现

### TenantContext 定义
```go
type TenantContext struct {
    TenantID     uint64
    TenantCode   string
    PlanID       uint64
    PlanFeatures []string  // 从套餐读取的功能列表
}
```

### 中间件流程
1. 从 Header X-Tenant-ID 取租户ID
2. 验证租户状态（非正常状态返回403）
3. 加载租户套餐配置，写入缓存（5分钟TTL）
4. 注入 tenant_id 到 context

### Repository 层
- 所有查询自动追加 tenant_id 条件
- 所有 INSERT 自动填入 tenant_id
- 基类 BaseRepository 实现公共逻辑

### 用量校验中间件
- CheckProductLimit：商品数量是否超套餐上限
- CheckOrderLimit：月订单是否超套餐上限
- CheckFeatureEnabled：操作是否在套餐功能范围内

## 四、套餐与计费实现

### 套餐功能枚举
```go
const (
    FeatureMultiSKU          = "multi_sku"
    FeatureCoupon            = "coupon"
    FeatureSeckill          = "seckill"
    FeatureGroupBuy         = "group_buy"
    FeatureDistribution      = "distribution"
    FeaturePoints           = "points"
    FeatureMemberLevel      = "member_level"
    FeatureCustomDomain     = "custom_domain"
    FeaturePrivateDeployment = "private_deployment"
)
```

### 三档套餐数据
```go
var Plans = []Plan{
    {Name:"基础版", Code:"basic", MonthlyFee:299, YearlyFee:2990, ProductLimit:100, OrderLimit:500, UserLimit:1000, Features:[]string{FeatureMultiSKU, FeatureCoupon, FeaturePoints}},
    {Name:"专业版", Code:"professional", MonthlyFee:799, YearlyFee:7990, ProductLimit:2000, OrderLimit:10000, UserLimit:50000, Features:[]string{FeatureMultiSKU, FeatureCoupon, FeatureSeckill, FeatureGroupBuy, FeatureDistribution, FeaturePoints, FeatureMemberLevel}},
    {Name:"旗舰版", Code:"enterprise", MonthlyFee:1999, YearlyFee:19990, ProductLimit:0, OrderLimit:0, UserLimit:0, Features:[]string{ALL_FEATURES}},
}
```

### 套餐到期检测（每天定时任务）
- 到期前7天/3天/1天发送预警通知
- 到期7天后自动封禁租户（status=3）

### 续费/升级/降级
- 续费：延长 plan_expire_at，创建 plan_log
- 升级：立即生效，按剩余时间折算差价
- 降级：下个账期生效

## 五、API 设计（必须全部实现）

### 认证与租户
- POST /api/v1/admin/auth/login {username, password} → {token, admin_info}
- 商户登录必须携带 X-Tenant-ID，并校验管理员 tenant_id 与请求租户一致；租户编号解析需兼容大小写输入；本地演示商户账号为租户编号 TEST001、用户名 smokeadmin22、密码 admin123。
- POST /api/v1/admin/auth/refresh {refresh_token} → {token}
- POST /api/v1/tenant/register {company_name, contact_info, plan_id} → {tenant_id, status}
- GET /api/v1/tenant/plans → [plans]
- PUT /api/v1/tenant/wechat-config {appid, secret, mchid, apiv3_key, cert_serial}
- GET /api/v1/platform/tenants（平台：分页+状态筛选）
- POST /api/v1/platform/tenants/{id}/audit {status, reject_reason}
- PUT /api/v1/platform/tenants/{id}/status

### 小程序端 - 会员
- POST /api/v1/member/auth/login-by-wechat {code} → {token, openid, is_new, member_info}
- POST /api/v1/member/auth/wechat-phone {code, encrypted_data, iv} → {phone}
- GET /api/v1/member/profile
- PUT /api/v1/member/profile {nickname, avatar, gender, birthday}
- GET /api/v1/member/points
- GET /api/v1/member/points/logs {page, page_size}
- GET /api/v1/member/coupons {status}
- GET /api/v1/member/distribution（我的下线+佣金）

### 小程序端 - 商品

- GET /api/v1/member/products {page, size, category_id, keyword} → {list, total, page, size}
- GET /api/v1/member/products/{id} → {product, skus}
- GET /api/v1/member/products/hot {page, size} → {list, total}
- GET /api/v1/member/products/recommend {page, size} → {list, total}
- GET /api/v1/member/categories → [categories]
- 小程序分类页“全部商品”默认调用 /api/v1/member/products，首屏 page=1，触底递增 page 并追加 list，切换分类/搜索/热门/推荐时重置页码和列表。

### 小程序端 - UI 交互

- 按钮布局必须按场景判断：登录、保存、空态引导等主行动居中或满宽；结算/订单详情使用跟随内容的行内操作组；搜索、领取、数量加减等局部工具按钮保持紧凑。
- 全局按钮样式只负责外观，不应通过通用选择器强制所有卡片按钮居右。
- 会员中心头图下不要生成独立三宫格统计卡；积分、成长值摘要放在头图或“会员资产”模块中，避免重复占位。
- 使用自定义底部导航的页面必须在根节点加 `tab-page`，底部 padding 要覆盖导航条高度、渐变壳和 `env(safe-area-inset-bottom)`，避免页面底部内容被导航遮挡。

### 小程序端 - 订单
- POST /api/v1/orders {items:[{product_id, sku_id, quantity}], address_id, coupon_id, buyer_remark, delivery_type}
  → {order_no, actual_amount}
- GET /api/v1/orders {page, page_size, status}
- GET /api/v1/orders/{id}/detail
- POST /api/v1/orders/{id}/cancel（待支付状态）
- POST /api/v1/orders/{id}/confirm（确认收货）
- GET /api/v1/member/after-sale-reasons {type} → 启用售后原因列表，用于小程序下拉
- POST /api/v1/member/orders/{id}/after-sales {type, reason, description, amount, images} → 发起整单售后，reason 必须来自启用原因，订单进入 refunding
- GET /api/v1/member/after-sales {status, order_id, page, size} → 我的售后单
- GET /api/v1/member/after-sales/{id} → 售后详情
- POST /api/v1/member/after-sales/{id}/return {return_express_company, return_express_no} → 提交退货物流
- POST /api/v1/member/after-sales/{id}/cancel → 取消待处理售后并恢复订单原状态
- GET /api/v1/orders/{id}/express（查询快递）

### 小程序端 - 营销
- GET /api/v1/coupons/available → 可领取优惠券列表
- POST /api/v1/coupons/{id}/receive
- GET /api/v1/coupons/can-use {order_amount} → 下单时可用的优惠券
- GET /api/v1/seckills/active → 当前秒杀场次
- GET /api/v1/seckills/{id}/products
- GET /api/v1/group-buys/active
- POST /api/v1/group-buys/{id}/join

### 小程序端 - 支付

- POST /api/v1/member/payments {order_no} → {payment_no, pay_params, mock_paid, status}
- POST /api/v1/payments/callback/wechat → 顾客订单微信支付回调（V3 AES-256-GCM 解密，平台商户号与金额校验）
- POST /api/v1/public/subscription/callback → 租户订阅微信支付回调，仅处理 tenant_subscription_orders
- 顾客订单支付前期生产环境使用平台统一微信商户号 `/v3/pay/transactions/jsapi`，资金先进入平台账户，并通过 `payments.settlement_tenant_id` 归集到待结算租户；平台订阅付款同样使用平台自有收款配置。

### 管理端 - 商品

- POST /api/v1/admin/products
- PUT /api/v1/admin/products/{id}
- PUT /api/v1/admin/products/{id}/status {status:0下架/1上架}
- DELETE /api/v1/admin/products/{id}（软删除）
- POST /api/v1/admin/products/{id}/skus
- PUT /api/v1/admin/skus/{id}
- POST /api/v1/admin/categories
- PUT /api/v1/admin/categories/{id}

### 管理端 - 订单

- GET /api/v1/admin/orders {page, page_size, status, order_no, date_range}
- PUT /api/v1/admin/orders/{id}/status
- POST /api/v1/admin/orders/{id}/ship {express_company, express_no}
- GET /api/v1/admin/after-sales {page,size,status,order_id,member_id} → 售后审核列表
- GET /api/v1/admin/after-sales/{id} → 售后详情
- POST /api/v1/admin/after-sales/{id}/approve {remark} → 审核通过
- POST /api/v1/admin/after-sales/{id}/reject {remark} → 驳回并恢复订单原状态
- POST /api/v1/admin/after-sales/{id}/receive {remark} → 确认收到退货
- POST /api/v1/admin/after-sales/{id}/refund {remark} → 标记退款完成，订单进入 refunded
- GET /api/v1/admin/orders/export（导出Excel）

### 平台端 - 售后原因

- GET /api/v1/platform/after-sale-reasons → 平台售后原因列表
- POST /api/v1/platform/after-sale-reasons {code,label,type,sort_order,enabled} → 新增原因
- PUT /api/v1/platform/after-sale-reasons/{id} {label,type,sort_order,enabled} → 编辑原因
- PATCH /api/v1/platform/after-sale-reasons/{id}/enabled {enabled} → 启用/停用
- DELETE /api/v1/platform/after-sale-reasons/{id} → 删除未使用原因；已上线原因优先停用。

### 管理端 - 收款配置

- GET /api/v1/admin/settings/payment → 当前租户收款配置列表
- PUT /api/v1/admin/settings/payment {provider:'wechat', sub_mchid, sub_appid, notify_url, mch_id?, app_id?, api_v3_key?, cert_serial_no?, private_key_pem?, cert_pem?} → 提交后 audit_status=pending, enabled=0
- GET /api/v1/platform/settings → 平台全局设置，包含订阅收款 wxpay_*和服务商 sp_* 字段
- PUT /api/v1/platform/settings → 更新平台基础信息、wxpay_*、sp_appid、sp_mchid、sp_apiv3_key、sp_cert_serial、partner_notify_url
- GET /api/v1/platform/payment-configs {page,size,status} → 租户子商户配置审核列表
- POST /api/v1/platform/payment-configs/{id}/audit {approve, remark} → 审核通过后 enabled=1
- 服务商字段用于顾客订单，wxpay_* 字段用于平台订阅收款，两条链路不得复用。

### 管理端 - 优惠券/拼团/秒杀/分销（CRUD + 统计）

### 管理端 - 仪表盘

- GET /api/v1/platform/dashboard → {tenant_count, active_tenant_count, today_gmv, today_orders, today_new_members, expire_soon_tenants, top_tenants}

### 响应格式

```json
// 成功
{"code":0,"message":"success","data":{}}
// 列表
{"code":0,"message":"success","data":{"list":[],"pagination":{"page":1,"page_size":20,"total":100,"total_pages":5}}}
// 错误
{"code":30001,"message":"商品库存不足","data":null}
```

### 错误码规范

- 1xxxx：认证/权限
- 2xxxx：参数校验
- 3xxxx：业务逻辑（库存不足=30001，余额不足=30002，套餐到期=30003，功能未开通=30004）
- 4xxxx：微信API错误
- 5xxxx：系统内部错误

## 六、微信对接实现

### 微信登录

```go
// 1. 小程序调用 wx.login() 获取 code
// 2. 我们的 API 调用微信 code2session
// URL: https://api.weixin.qq.com/sns/jscode2session
// 参数: appid + secret + js_code + grant_type=authorization_code
// 返回: openid + session_key
// 3. 查询或创建会员（openid 匹配）
// 4. JWT 签发：{member_id, tenant_id, openid, exp}
// 5. 手机号解密：session_key 作为 AES-256-CBC Key 解密 encryptedData
```

### 微信支付（JSAPI / 服务商模式）

```go
// 1. 平台订阅付款：POST https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi
// Header: Authorization: WECHATPAY2-SHA256-RSA2048 + 证书签名
// Body: {appid, mchid, description, out_trade_no, notify_url, amount:{total}, payer:{openid}}
// - appid/mchid 使用平台收款配置，out_trade_no = tenant_subscription_orders.order_no
// - 回调：POST /api/v1/public/subscription/callback
//
// 2. 顾客订单付款：POST https://api.mch.weixin.qq.com/v3/pay/partner/transactions/jsapi
// Header: Authorization: WECHATPAY2-SHA256-RSA2048 + 服务商证书签名
// Body: {sp_appid, sp_mchid, sub_appid, sub_mchid, description, out_trade_no, notify_url, amount:{total}, payer:{sp_openid/sub_openid}}
// - sp_appid/sp_mchid 使用平台服务商配置，sub_appid/sub_mchid 使用当前租户审核通过的子商户配置
// - out_trade_no = payments.payment_no，payments.pay_scene = member_order
// - 回调：POST /api/v1/payments/callback/wechat
//
// 3. 拿 prepay_id，生成调起小程序支付参数
// 4. 回调处理
// - 验证签名（微信发来的是 AES-256-GCM 加密的 ciphertext）
// - 校验 out_trade_no、金额、平台 mchid、trade_state
// - 解密后处理：更新 payment 状态 → 更新 order 状态 → 扣库存 → 发积分 → 计算分佣
// - 返回 HTTP 200 + {"code":"SUCCESS","message":"SUCCESS"}
// - 生产环境缺少平台微信支付配置时必须拒绝支付下单，不允许模拟支付成功
```

## 七、实现要求

### 必做
1. 所有 Repository 实现 BaseRepository 基类的自动 tenant_id 注入
2. 中间件实现：租户注入、JWT鉴权、限流、日志、panic恢复、CORS
3. 用量校验中间件：ProductLimit、OrderLimit、FeatureEnabled
4. 微信code2session、支付JSAPI统一下单、退款API的完整实现
5. 所有 API Handler 配单元测试（至少覆盖 happy path）
6. 库存扣减使用 Redis 分布式锁（SETNX + TTL）
7. 订单号使用 Snowflake ID
8. 配置文件全用 YAML（不写死任何值）
9. Docker Compose 按基础服务与应用服务拆分维护

### 选做
1. 拼团/秒杀的防超卖（Redis原子扣减 + 乐观锁双重保障）
2. 分佣的定时结算任务
3. 操作日志的 AOP 切面自动记录
4. 支付宝支付对接（结构预留）
```
