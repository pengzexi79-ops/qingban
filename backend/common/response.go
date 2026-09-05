package common

import (
	"qingban/core"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Response 统一响应体:成功 {code:0, data, message:"ok"};错误 {code:-1, message}。
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
}

func Success(c *gin.Context, data interface{}) {
	sendSuccess(c, data)
}

// sendSuccess debug 模式下记录响应日志后返回 200 统一成功体
func sendSuccess(c *gin.Context, data interface{}) {
	if core.Mode == "debug" {
		logrus.Debug("成功的响应:", data)
	}
	c.JSON(200, Response{Code: 0, Data: data, Message: "ok"})
}

// Error 返回统一错误响应体(HTTP 状态码由调用方给出)
func Error(c *gin.Context, code int, message string) {
	if core.Mode == "debug" {
		logrus.Debugf("错误响应: %s %s → %d %s", c.Request.Method, c.Request.URL.Path, code, message)
	}
	c.JSON(code, Response{Code: -1, Message: message})
}

func Unauthorized(c *gin.Context, message string) {
	if core.Mode == "debug" {
		logrus.Debugf("未授权: %s %s → 401 %s", c.Request.Method, c.Request.URL.Path, message)
	}
	c.JSON(401, Response{Code: -1, Message: message})
}

func Forbidden(c *gin.Context, message string) {
	if core.Mode == "debug" {
		logrus.Debugf("禁止访问: %s %s → 403 %s", c.Request.Method, c.Request.URL.Path, message)
	}
	c.JSON(403, Response{Code: -1, Message: message})
}
