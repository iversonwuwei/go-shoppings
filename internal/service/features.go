package service

// 套餐功能常量（与设计文档一致）
const (
	FeatureMultiSKU          = "multi_sku"
	FeatureVirtualProduct    = "virtual_product"
	FeatureSeckill           = "seckill"
	FeatureGroupBuy          = "group_buy"
	FeatureDistribution      = "distribution"
	FeatureCoupon            = "coupon"
	FeaturePoints            = "points"
	FeatureMemberLevel       = "member_level"
	FeatureExpressDelivery   = "express_delivery"
	FeatureCityDelivery      = "city_delivery"
	FeatureSelfPickup        = "self_pickup"
	FeatureCustomDomain      = "custom_domain"
	FeaturePrivateDeployment = "private_deployment"
	FeatureWhiteLabel        = "white_label"
	FeatureSmsNotification   = "sms_notification"
	FeatureAPIAccess         = "api_access"
)

// 租户状态
const (
	TenantStatusPending int8 = 0
	TenantStatusActive  int8 = 1
	TenantStatusOverdue int8 = 2
	TenantStatusBanned  int8 = 3
)
