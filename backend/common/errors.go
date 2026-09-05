package common

// 统一错误模型(与 openapi.phase1.yaml components ErrorResponse 对齐)。
// 用途:全后端唯一错误出口,保证前端只拿到结构化错误,绝不外泄供应商密钥/原始响应。
// 本文件为纯定义+辅助,不依赖 gin(HTTP 映射在 server/response 或 common 的响应层完成)。

// 错误码常量(第一阶段常用;扩展码后续阶段追加)。
// 约定:错误码决定可读性语义,HTTP 状态码决定传输语义,两者由 ToHTTPStatus 映射。
const (
	// CodeUnauthorized:本地令牌缺失/不匹配(401)
	CodeUnauthorized = "UNAUTHORIZED"
	// CodeNotFound:资源不存在或不属于当前本地空间(404)
	CodeNotFound = "NOT_FOUND"
	// CodeValidationError:请求参数校验失败(422)
	CodeValidationError = "VALIDATION_ERROR"
	// CodeRateLimited:触发频率超限(429)
	CodeRateLimited = "RATE_LIMITED"
	// CodeProactiveDisabled:用户/角色关闭了主动消息总开关(409)
	CodeProactiveDisabled = "PROACTIVE_DISABLED"
	// CodeCooldownActive:群聊轮次/主动消息仍在冷却期(409)
	CodeCooldownActive = "COOLDOWN_ACTIVE"
	// CodeProviderError:模型服务商不可达/调用失败(502)
	CodeProviderError = "PROVIDER_ERROR"
	// CodeFileReferenced:文件仍被消息等引用,拒绝物理删除(409)
	CodeFileReferenced = "FILE_REFERENCED"
	// CodeNotImplemented:路由已注册但本阶段未实现(501)
	CodeNotImplemented = "NOT_IMPLEMENTED"
	// CodeConflict:数据冲突(如本地空间已初始化、唯一配置删除保护)(409)
	CodeConflict = "CONFLICT"
	// CodeInternal:未分类内部错误(500)
	CodeInternal = "INTERNAL"
)

// ErrorResponse:统一错误响应体(JSON 结构,字段与 yaml 一致)。
type ErrorResponse struct {
	// Code:错误码(上方常量之一)。
	Code string `json:"code"`
	// Message:给用户看的简短说明(中文,可直接展示;不携带密钥/供应商原文)。
	Message string `json:"message"`
	// TraceId:本次请求追踪号(为空时前端可忽略;有值时便于日志定位)。
	TraceId string `json:"traceId,omitempty"`
	// Details:可选的补充结构(如 FILE_REFERENCED 的 refCount)。
	Details any `json:"details,omitempty"`
}

// AppError:后端内部流转的错误对象。作用:让任意深层调用(service/AI 层)都能
// 抛出一个"可映射为统一响应"的错误,而不必关心 HTTP 细节。
type AppError struct {
	// Code:错误码。
	Code string
	// Message:用户可读信息。
	Message string
	// HTTPStatus:建议的 HTTP 状态码(可为 0,由 ToHTTPStatus 按 Code 兜底)。
	HTTPStatus int
	// Err:底层原始错误(仅日志使用,绝不透传给前端)。
	Err error
	// Details:随响应返回的补充结构。
	Details any
}

// Error:实现 error 接口(日志打印时携带 code)。
func (e *AppError) Error() string {
	// 实现:组装 "code: message"(err != nil 时附底层原因)
	return e.Code + ": " + e.Message
}

// NewError:构造 AppError(业务层通用错误出口)。
// 调用示例:return nil, common.NewError(common.CodeNotFound, "角色不存在", http.StatusNotFound, err)
func NewError(code, message string, httpStatus int, cause error) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus, Err: cause}
}

// ToHTTPStatus:按错误码兜底映射 HTTP 状态码(显式 HTTPStatus 优先)。
// 用途:统一 401/404/409/422/429/501/502 语义,避免各 handler 散落状态码。
func (e *AppError) ToHTTPStatus() int {
	if e.HTTPStatus != 0 {
		return e.HTTPStatus
	}
	switch e.Code {
	case CodeUnauthorized:
		return 401
	case CodeNotFound:
		return 404
	case CodeValidationError:
		return 422
	case CodeRateLimited:
		return 429
	case CodeProactiveDisabled, CodeCooldownActive, CodeFileReferenced, CodeConflict:
		return 409
	case CodeProviderError:
		return 502
	case CodeNotImplemented:
		return 501
	default:
		return 500
	}
}

// AsAppError:将任意 error 归一为 *AppError(未知错误映射为 INTERNAL,并保留 trace 需求)。
func AsAppError(err error) *AppError {
	// 逻辑:类型断言 *AppError 成功直接返回;失败则包一层 INTERNAL(原始 err 只进日志)
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewError(CodeInternal, "服务内部错误,请稍后重试", 0, err)
}
