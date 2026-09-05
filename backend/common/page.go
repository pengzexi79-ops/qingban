package common

import (
	"net/url"
	"qingban/core"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PageInfo 通用分页请求参数
type PageInfo struct {
	// 游标定位
	Cursor int64 `json:"cursor" form:"cursor"` // 起始游标(从这条记录开始)
	// 页码偏移
	Page     int    `json:"page" form:"page"`           // 从 cursor 开始的第几页(从 1 开始)
	PageSize int    `json:"page_size" form:"page_size"` // 每页条数
	Key      string `json:"key" form:"key"`             // 模糊匹配关键词(业务侧指定 KeyList 列)
	Order    string `json:"order" form:"order"`         // 排序(业务侧白名单枚举,如 name_asc/time_desc;空 = 默认)
}

// TimeInfo 查询时间范围(匹配 gorm.Model 的 CreatedAt,即 created_at 字段),独立字段
type TimeInfo struct {
	StartTime time.Time
	EndTime   time.Time
}

type Option struct {
	PageInfo
	DefaultOrder string                    // 默认排序方式
	Orderby      string                    // 排序字段
	Filters      []func(*gorm.DB) *gorm.DB // 额外查询条件
	KeyList      []string                  // key 所匹配的字段(OR 条件)
}

// PaginatedResult 通用分页响应
type PaginatedResult[T any] struct {
	List       []T   `json:"list"`        // 当前页数据列表
	Total      int64 `json:"total"`       // 总条数(与 List 同一过滤口径)
	Page       int   `json:"page"`        // 当前页码
	PageSize   int   `json:"page_size"`   // 每页条数
	TotalPages int   `json:"total_pages"` // 总页数
	HasMore    bool  `json:"has_more"`    // 是否还有更多
	NextCursor int64 `json:"next_cursor"` // 下一页起始游标
}

// NewOption 创建分页参数
//   - page: 页码（从 1 开始，小于 1 时修正为 1）
//   - pageSize: 每页条数（0 或负数时修正为 10，最大 100）
func NewOption(page, pageSize int) *Option {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &Option{
		PageInfo: PageInfo{
			Page:     page,
			PageSize: pageSize,
		},
	}
}

// ParseOption 从通用查询参数构造 Option(纯函数,不依赖 gin;gin 调用方传
// c.Request.URL.Query() 即可)。支持的参数:
//   - cursor:     起始游标
//   - page:       当前页码(默认 1)
//   - page_size:  每页条数(默认 10)
//   - key:        模糊搜索关键词
//   - order:      排序方向(ASC / DESC)
//
// 时间范围不在此解析:Option 不承载时间字段,需要时间过滤时由调用方自行
// 解析参数并构造 TimeInfo 传给 common.PageQuery 的 timeInfo 参数。
//
// 推荐用法:调用方构造好 Option 后传入 common.PageQuery,不要在业务代码里
// 逐个从请求拆字段。
func ParseOption(values url.Values) *Option {
	opt := NewOption(intVal(values.Get("page"), 1), intVal(values.Get("page_size"), 10))
	opt.Cursor = int64(intVal(values.Get("cursor"), 0))
	opt.Key = values.Get("key")
	opt.Order = values.Get("order")
	return opt
}

// PageQuery 通用分页查询(全项目唯一分页入口):model 同时确定"表"与"行类型",
// opt 携带分页/排序/过滤参数(NewOption / ParseOption 构造),timeInfo 携带
// 时间范围(created_at 闭区间;nil = 不限)。内部走全局 core.DB,不依赖 gin;
// 返回统一 PaginatedResult(含 Total,调用方无需再单独 Count)。
//
// 支持两种分页模式:
//   - 游标优先(Cursor > 0):WHERE id > Cursor [AND 其他条件] [OFFSET (Page-1)*PageSize] LIMIT PageSize
//     利用索引跳过已翻页区间,避免深 Offset 性能问题;
//   - 传统 Offset(Cursor == 0):[OFFSET (Page-1)*PageSize] LIMIT PageSize。
//
// 内置:时间范围过滤(created_at)、关键词模糊搜索(KeyList 限定列,OR 语义)、
// 额外条件(Filters,Scope 模式)、安全排序(字段名白名单 + 方向归一)、
// 自动提取末条记录 ID 作为 NextCursor。
//
// 用法:
//
//	opt := common.ParseOption(c.Request.URL.Query()) // 或 common.NewOption(page, pageSize)
//	opt.DefaultOrder = "created_at DESC"
//	opt.Filters = append(opt.Filters, func(db *gorm.DB) *gorm.DB {
//		return db.Where("creator_id = ?", userID)
//	})
//	res, err := common.PageQuery(model.ShareLink{}, *opt, nil)
func PageQuery[T any](model T, opt Option, timeInfo *TimeInfo) (*PaginatedResult[T], error) {
	// 分页参数归一(调用方直接构造 Option 时的防御:page<1 → 1,pageSize 0/负 → 10,超 100 → 100)
	if opt.Page < 1 {
		opt.Page = 1
	}
	if opt.PageSize <= 0 {
		opt.PageSize = 10
	}
	if opt.PageSize > 100 {
		opt.PageSize = 100
	}

	// 时间范围(created_at 闭区间;timeInfo 为 nil = 不限)
	startTime, endTime := time.Time{}, time.Time{}
	if timeInfo != nil {
		startTime, endTime = timeInfo.StartTime, timeInfo.EndTime
	}

	// 以 model 绑定表名(gorm 反射取 schema);结果容器由泛型推导
	query := core.DB.Model(model)
	result := new([]T)

	// 1. 游标条件 — 利用索引范围扫描
	if opt.Cursor > 0 {
		query = query.Where("id > ?", opt.Cursor)
	}

	// 2. 时间范围(created_at 闭区间)
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	// 3. 模糊搜索(OR 语义,限定白名单列)
	if opt.Key != "" && len(opt.KeyList) > 0 {
		var conds []string
		var vals []any
		for _, field := range opt.KeyList {
			conds = append(conds, field+" LIKE ?")
			vals = append(vals, opt.Key+"%")
		}
		query = query.Where("("+strings.Join(conds, " OR ")+")", vals...)
	}

	// 4. 额外条件(Scope 模式,无副作用)
	query = query.Scopes(opt.Filters...)

	// 5. 统计总数(必须早于 Limit/Offset,只带过滤条件)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 6. 排序(字段名白名单 + 方向归一,非法回退,防注入)
	if order := buildOrder(&opt); order != "" {
		query = query.Order(order)
	}

	// 7. 分页偏移 & 限制
	if opt.Page > 1 {
		query = query.Offset((opt.Page - 1) * opt.PageSize)
	}
	query = query.Limit(opt.PageSize)

	// 8. 取本页数据
	if err := query.Find(result).Error; err != nil {
		return nil, err
	}

	// 9. 组装响应(TotalPages 向上取整;HasMore = 当前页非末页;NextCursor = 末条 ID)
	totalPages := (int(total) + opt.PageSize - 1) / opt.PageSize
	return &PaginatedResult[T]{
		List:       *result,
		Total:      total,
		Page:       opt.Page,
		PageSize:   opt.PageSize,
		TotalPages: totalPages,
		HasMore:    opt.Page < totalPages,
		NextCursor: extractLastID(*result),
	}, nil
}

// WrapPageResponse 将分页数据包装为统一成功响应,供 Gin Handler 直接调用
func WrapPageResponse[T any](c *gin.Context, data *PaginatedResult[T]) {
	Success(c, data)
}

// ---------- 辅助函数 ----------

// buildOrder 安全组装排序子句
// 优先级: Orderby + Order > DefaultOrder > 空（不排序）
func buildOrder(opt *Option) string {
	if opt.Orderby != "" {
		if !fieldNameRe.MatchString(opt.Orderby) {
			return "id DESC" // 不合法字段名时回退
		}
		dir := strings.ToUpper(strings.TrimSpace(opt.Order))
		if dir != "ASC" && dir != "DESC" {
			dir = "DESC"
		}
		return opt.Orderby + " " + dir
	}
	if opt.DefaultOrder != "" {
		return opt.DefaultOrder
	}
	return ""
}

// fieldNameRe 校验排序字段名是否合法（仅字母、数字、下划线）
var fieldNameRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// extractLastID 通过反射提取结果集中最后一条记录的 ID 字段
// 用于自动填充 NextCursor。支持 uint 和 int64 类型的 ID。
func extractLastID[T any](items []T) int64 {
	if len(items) == 0 {
		return 0
	}
	v := reflect.ValueOf(items[len(items)-1])
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		if f := v.FieldByName("ID"); f.IsValid() {
			switch f.Kind() {
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return int64(f.Uint())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return f.Int()
			}
		}
	}
	return 0
}

// intVal 从查询参数读取整数,不存在或非法时返回默认值(纯函数,不依赖 gin)。
func intVal(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

// ---------- keyset 游标分页(id DESC) ----------
// 用途:深分页遍历(如前端递归下载逐层翻页)的 keyset 游标——以**末条记录主键**
// 作为 nextPageToken,配合 ORDER BY <prefix>id DESC 使用,避免 offset 深翻页性能
// 问题(游标即不透明令牌,客户端原样回传)。
// 与 common.PageQuery(id 升序, `id > cursor`)互补;本组为降序变体(`id < cursor`)。
// 注:keyset 快照式遍历——遍历中新插入的记录(id 更大)不进入剩余页,属预期语义。

// EncodeKeysetID 编码 keyset 游标(末条记录 ID 的十进制串)。
func EncodeKeysetID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

// DecodeKeysetID 解码 keyset 游标(EncodeKeysetID 的逆)。
// 空串/非法 → 0(调用方按"无游标/首页"处理)。
func DecodeKeysetID(s string) uint {
	id, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return uint(id)
}

// KeysetIDCond 构建 keyset 谓词(配合 ORDER BY <prefix>id DESC):
// 返回 `(<prefix>id < ?)` 与参数;prefix 为表别名(空 = 无前缀)。
// 游标空串/非法 → 空谓词与空参数(首页,调用方不加 WHERE 条件)。
func KeysetIDCond(prefix, cursor string) (string, []any) {
	id := DecodeKeysetID(cursor)
	if id == 0 {
		return "", nil
	}
	if prefix != "" {
		prefix += "."
	}
	return prefix + "id < ?", []any{id}
}
