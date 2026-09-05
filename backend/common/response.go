package common

// gin 响应层:统一 JSON 成功/失败出口。
// 依赖:github.com/gin-gonic/gin(已在 go.mod)。所有 handler 通过本层返回,保证错误体结构一致。

import (
	"github.com/gin-gonic/gin"
)

// ginTraceKey:gin.Context 中存放 traceId 的键名(供响应与日志读取)。
const ginTraceKey = "traceId"

// TraceFrom:从请求上下文取 traceId(middleware 注入;无则空串)。
func TraceFrom(c *gin.Context) string {
	if v, ok := c.Get(ginTraceKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// OK:成功响应(c.JSON)。
func OK(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

// Fail:统一失败出口:error 归一为 ErrorResponse 输出(不再写第二次响应)。
func Fail(c *gin.Context, err error) {
	// appErr := AsAppError(err)                       // ① 归一(未知错误 → INTERNAL)
	// status := appErr.ToHTTPStatus()                 // ② 错误码 → HTTP 状态映射
	// core.Log.Error(...)                             // ③ 错误日志(带 traceId/code/底层 err)
	// c.JSON(status, ErrorResponse{Code, Message,     // ④ 统一错误体
	//     TraceId: TraceFrom(c), Details})
	appErr := AsAppError(err)
	resp := ErrorResponse{
		Code:    appErr.Code,
		Message: appErr.Message,
		TraceId: TraceFrom(c),
		Details: appErr.Details,
	}
	c.JSON(appErr.ToHTTPStatus(), resp)
	// 注意:handler 返回后必须 return,避免继续写 200 响应(单次请求只允许一个响应)。
}

// AbortFail:鉴权/中间件层失败(Abort 掉后续 handler 后统一响应)。
func AbortFail(c *gin.Context, err error) {
	c.Abort()
	Fail(c, err)
}

// ServerBusy/Page 辅助(无:分页已由各 handler 用 common.Page[T] 直接组装)。
