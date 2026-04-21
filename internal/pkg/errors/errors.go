package errors

import "net/http"

type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }

func (e *BizError) HTTPStatus() int {
	switch {
	case e.Code >= 10000 && e.Code < 20000:
		return http.StatusUnauthorized
	case e.Code >= 20000 && e.Code < 30000:
		return http.StatusBadRequest
	case e.Code >= 30000 && e.Code < 40000:
		return http.StatusOK // 业务错误通过 code 区分
	case e.Code >= 40000 && e.Code < 50000:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

func New(code int, msg string) *BizError { return &BizError{Code: code, Msg: msg} }

// 常见错误
var (
	ErrUnauthorized    = New(10001, "未登录或登录已过期")
	ErrForbidden       = New(10002, "无权访问")
	ErrInvalidToken    = New(10003, "无效的Token")
	ErrTenantRequired  = New(10004, "缺少租户标识")
	ErrTenantInvalid   = New(10005, "租户无效或已封禁")
	ErrParamInvalid    = New(20001, "参数错误")
	ErrNotFound        = New(20404, "资源不存在")
	ErrStockShortage   = New(30001, "库存不足")
	ErrBalanceShortage = New(30002, "余额不足")
	ErrPlanExpired     = New(30003, "套餐已到期，请续费")
	ErrFeatureDisabled = New(30004, "当前套餐未开通该功能")
	ErrLimitExceeded   = New(30005, "已达到套餐用量上限")
	ErrDuplicated      = New(30006, "记录已存在")
	ErrWechatAPI       = New(40001, "微信接口调用失败")
	ErrWechatPay       = New(40002, "微信支付失败")
	ErrInternal        = New(50001, "服务器内部错误")
)
