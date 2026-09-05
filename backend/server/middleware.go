package server

// HTTP 中间件:令牌鉴权、traceId、幂等键。
// 鉴权背景(PHASE1 §2):桌面本地进程,监听 127.0.0.1;启动令牌 X-Local-Token
// 仅作"本机进程间护栏"(防浏览器其他网页误连),非账号体系;
// 浏览器调试模式(core.Cfg.AllowEmptyToken)可留空。日后若引入云账号,
// 令牌头之上再叠加登录态,接口面不变。
// 伪代码草稿:中间件闭包内为逻辑占位;实现时恢复 import(qingban/core、qingban/utils 等)。

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// headerLocalToken:X-Local-Token 请求头名。
const headerLocalToken = "X-Local-Token"

// headerIdempotency:Idempotency-Key 请求头名(发送类接口专用)。
const headerIdempotency = "Idempotency-Key"

// hdrTrace:透传给响应与日志的 traceId 头名。
const hdrTrace = "X-Trace-Id"

// traceCtxKey:gin.Context 存取 traceId 的键。
const traceCtxKey = "traceId"

// MiddlewareTrace:每请求注入 traceId(响应头 X-Trace-Id + 日志上下文)。
func MiddlewareTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		// tid := utils.PrefixedID("trace")
		// c.Set(traceCtxKey, tid); c.Header(hdrTrace, tid)
		// start := time.Now()
		// c.Next()
		// log(c, "access", {method: c.Request.Method, path: c.Request.URL.Path, status: c.Writer.Status(), ms: time.Since(start).Milliseconds(), trace: tid})
	}
}

// MiddlewareLocalToken:X-Local-Token 校验(挂令牌保护端点组)。
func MiddlewareLocalToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// got := c.GetHeader(headerLocalToken)
		// if got == core.LocalToken { c.Next(); return }                    // 令牌匹配
		// if core.Cfg.AllowEmptyToken && got == "" { c.Next(); return }    // 浏览器调试(仅本地)
		// abortErr(c, 401, CodeUnauthorized, "本地令牌无效或缺失")
	}
}

// MiddlewareIdempotency:幂等中间件(挂发送/触发轮次/导入等 POST 端点)。
// 语义:同一 Idempotency-Key 只执行一次,命中直接重放首次响应体(不重复产生副作用)。
func MiddlewareIdempotency() gin.HandlerFunc {
	return func(c *gin.Context) {
		// key := c.GetHeader(headerIdempotency)
		// if key == "" { c.Next(); return }                                // 无幂等要求
		// if len(key) > 128 { abortErr(c, 422, CodeValidationError, "Idempotency-Key 过长"); return }
		// if rec, ok := core.Idem.Get(key); ok {                           // 已执行 → 按首次状态码+响应体重放
		//     c.Data(rec.StatusCode, rec.Response); c.Abort(); return
		// }
		// buf := newCaptureWriter(c.Writer)                                // 覆写 Write 缓存响应体
		// c.Writer = buf
		// c.Next()
		// if buf.status == 200 || buf.status == 201 {                      // 成功才登记;错误不登记允许修参重试
		//     core.Idem.Register(key, buf.status, buf.body)
		// }
	}
}

// isStreamRequest:请求是否要求 SSE 流式(Accept: text/event-stream)。
// 调用点:hSendMessage 分流(流式/同步)。
func isStreamRequest(c *gin.Context) bool {
	// accept := c.GetHeader("Accept")
	// return strings.Contains(accept, "text/event-stream")
	return strings.Contains(c.GetHeader("Accept"), "text/event-stream")
}
