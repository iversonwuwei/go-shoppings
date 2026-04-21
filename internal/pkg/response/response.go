package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperr "wechat-mall-saas/internal/pkg/errors"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "success", Data: data})
}

func Fail(c *gin.Context, err error) {
	if be, ok := err.(*apperr.BizError); ok {
		c.JSON(be.HTTPStatus(), Body{Code: be.Code, Message: be.Msg})
		return
	}
	c.JSON(http.StatusInternalServerError, Body{Code: 50000, Message: err.Error()})
}

func FailCode(c *gin.Context, code int, msg string) {
	status := http.StatusBadRequest
	if code >= 10000 && code < 20000 {
		status = http.StatusUnauthorized
	} else if code >= 50000 {
		status = http.StatusInternalServerError
	}
	c.JSON(status, Body{Code: code, Message: msg})
}
