package tests

// common 包测试:分页游标、错误码 → HTTP 状态映射、AsAppError 归一。

import (
	"errors"
	"testing"

	"qingban/common"
)

func TestParseCursor(t *testing.T) {
	// 默认 20
	c := common.ParseCursor("", 0)
	if c.Limit != 20 {
		t.Errorf("默认 limit=20,得到 %d", c.Limit)
	}
	// 下限保护
	if c = common.ParseCursor("x", -5); c.Limit != 20 {
		t.Errorf("负数应回落默认 20,得到 %d", c.Limit)
	}
	// 上限 100
	if c = common.ParseCursor("x", 999); c.Limit != 100 {
		t.Errorf("超上限应夹取 100,得到 %d", c.Limit)
	}
	// before 原样透传
	if c = common.ParseCursor("msg-123", 5); c.Before != "msg-123" || c.Limit != 5 {
		t.Errorf("透传异常: %+v", c)
	}
}

func TestHasMore(t *testing.T) {
	if !common.HasMore(21, 20) {
		t.Error("取到 limit+1 条应有更多")
	}
	if common.HasMore(20, 20) {
		t.Error("恰好 limit 条不应有更多")
	}
}

func TestErrorCodeToHTTPStatus(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		{common.CodeUnauthorized, 401},
		{common.CodeNotFound, 404},
		{common.CodeValidationError, 422},
		{common.CodeRateLimited, 429},
		{common.CodeProactiveDisabled, 409},
		{common.CodeCooldownActive, 409},
		{common.CodeFileReferenced, 409},
		{common.CodeConflict, 409},
		{common.CodeProviderError, 502},
		{common.CodeNotImplemented, 501},
		{common.CodeInternal, 500},
		{"UNKNOWN_CODE", 500}, // 未知码兜底 500
	}
	for _, c := range cases {
		err := common.NewError(c.code, "x", 0, nil)
		if got := err.ToHTTPStatus(); got != c.want {
			t.Errorf("code=%s 映射 %d,期望 %d", c.code, got, c.want)
		}
	}
}

func TestAppError_ExplicitStatusWins(t *testing.T) {
	// 显式 HTTPStatus 优先于 code 映射
	err := common.NewError(common.CodeNotFound, "x", 422, nil)
	if got := err.ToHTTPStatus(); got != 422 {
		t.Errorf("显式 422 被覆盖为 %d", got)
	}
}

func TestAsAppError(t *testing.T) {
	// 已包装的错误原样返回
	orig := common.NewError(common.CodeNotFound, "角色不存在", 0, nil)
	if got := common.AsAppError(orig); got != orig {
		t.Error("AsAppError 应透传 *AppError")
	}
	// 任意 error → INTERNAL 包装
	plain := errors.New("boom")
	wrapped := common.AsAppError(plain)
	if wrapped.Code != common.CodeInternal {
		t.Errorf("未知错误应归一 INTERNAL,得到 %s", wrapped.Code)
	}
	// Error() 信息包含 code 前缀(日志友好)
	if err := orig.Error(); err != "NOT_FOUND: 角色不存在" {
		t.Errorf("Error()=%q", err)
	}
}

func TestErrorResponse_JSONShape(t *testing.T) {
	// 契约错误体字段:code/message/traceId/details
	_ = common.ErrorResponse{Code: "X", Message: "m"}
	_ = common.Page[int]{Items: []int{1}, NextCursor: "c"}
}
