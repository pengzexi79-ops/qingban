package common

// 统一响应与分页模型。
// 背景(PHASE1 §2):分页响应统一 { items:[], nextCursor:string|null };
// 游标 before=<lastId> + limit(1~100,默认 20);省略 before=最新一页。

// Page:通用游标分页响应体(items 泛型化,实体由各 server 文件传入)。
type Page[T any] struct {
	// Items:本页记录(≤ limit)。
	Items []T `json:"items"`
	// NextCursor:更早一页起点的记录 id(本页最旧一条);无更多旧数据时省略。
	NextCursor string `json:"nextCursor,omitempty"`
}

// CursorQuery:分页查询入参解析结果(service 层统一翻页条件)。
type CursorQuery struct {
	// Before:游标记录 id(空=最新一页)。
	Before string
	// Limit:本页条数(已夹取到 [1,100])。
	Limit int
}

// ParseCursor:解析游标并夹取 limit(1~100,默认 20)。
func ParseCursor(before string, limit int) CursorQuery {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return CursorQuery{Before: before, Limit: limit}
}

// HasMore:配合"取 limit+1 条"翻页惯用法:多取的那条说明还有更旧数据。
func HasMore(fetched int, limit int) bool {
	return fetched > limit
}
