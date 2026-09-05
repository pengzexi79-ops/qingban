package server

// P2 用量端点:汇总 / 明细 / 近 N 日趋势(读 usage_records 本地真实数据)。
// 原则(BACKEND_HANDOFF §6.7):API 损耗页读后端真实用量,不以前端估算作账单。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"
	// 实现时按需恢复:"qingban/model"(UsageRecord 实体)等
)

// UsageSummaryView:GET /usage/summary 响应(openapi UsageSummary)。
type UsageSummaryView struct {
	// Period:统计区间(from/to)。
	Period map[string]string `json:"period"`
	// TotalTokens:总 token(输入+输出+缓存)。
	TotalTokens int64 `json:"totalTokens"`
	// InputTokens/OutputTokens/CachedTokens:分项。
	InputTokens  int64 `json:"inputTokens"`
	OutputTokens int64 `json:"outputTokens"`
	CachedTokens int64 `json:"cachedTokens"`
	// EstimatedCost:估算费用(仅远程计费,Ollama 恒 0)。
	EstimatedCost float64 `json:"estimatedCost"`
	// ByCapability:按能力分项(chat/vision/…)。
	ByCapability []UsageCapSum `json:"byCapability"`
}

// UsageCapSum:单能力分项。
type UsageCapSum struct {
	// Capability:能力名。
	Capability string `json:"capability"`
	// Tokens:该能力 token 合计。
	Tokens int64 `json:"tokens"`
	// EstimatedCost:费用合计。
	EstimatedCost float64 `json:"estimatedCost"`
}

// UsageTrendPoint:趋势单点(openapi 结构)。
type UsageTrendPoint struct {
	// Date:YYYY-MM-DD。
	Date string `json:"date"`
	// InputTokens/OutputTokens/EstimatedCost:当日聚合。
	InputTokens   int64   `json:"inputTokens"`
	OutputTokens  int64   `json:"outputTokens"`
	EstimatedCost float64 `json:"estimatedCost"`
}

// UsageTrendResp:GET /usage/trend 响应。
type UsageTrendResp struct {
	// Points:按天升序。
	Points []UsageTrendPoint `json:"points"`
}

// hGetUsageSummary:GET /usage/summary?from=&to= —— 汇总。
func hGetUsageSummary(c *gin.Context) {
	// from, to := parseRange(c.Query("from"), c.Query("to"))            // 缺省近 30 天;本地时区日期
	// where := created_at ∈ [from 00:00, to 23:59:59] (本地时区)
	// rows := db.Query(`SELECT capability,
	//                        SUM(input_tokens), SUM(output_tokens), SUM(cached_tokens),
	//                        SUM(estimated_cost)
	//                   FROM usage_records {where} GROUP BY capability`)
	// // 语义决定:status=failed 仍计 estimated_cost(供应商可能计费),输出 token 按 0
	// v := UsageSummaryView{Period: {from, to}, ByCapability: ...}
	// v.TotalTokens = Σ 分项;v.EstimatedCost = Σ cost
	// respond(c, 200, v)
}

// UsageRecordQuery:GET /usage/records 查询参数。
type UsageRecordQuery struct {
	// Model:模型过滤。
	Model string `form:"model"`
	// Capability:能力过滤。
	Capability string `form:"capability"`
	// Before/Limit:游标分页(created_at 倒序)。
	Before string `form:"before"`
	Limit  int    `form:"limit"`
}

// hListUsageRecords:GET /usage/records —— 明细(分页)。
func hListUsageRecords(c *gin.Context) {
	// var q UsageRecordQuery; c.ShouldBindQuery(&q)
	// rows := db.Find(UsageRecord{},
	//     where: model==q.Model? AND capability==q.Capability? AND 游标 id < before?,
	//     order: created_at DESC, id DESC, limit+1 翻页)
	// respond(c, 200, pageOf(rows))
}

// hGetUsageTrend:GET /usage/trend?days=(默认 7,最大 90)—— 近 N 日趋势。
func hGetUsageTrend(c *gin.Context) {
	// days := clamp(c.Query("days") or 7, 1, 90)
	// buckets := genDays(days)                                          // 补零日期桶(缺失日 0,折线不跳空)
	// rows := db.Query(`SELECT date(created_at) d, SUM(input_tokens), SUM(output_tokens),
	//                          SUM(estimated_cost)
	//                    FROM usage_records WHERE created_at >= now-days GROUP BY d`)
	// for b := range buckets { b.填入 rows[d] }                          // 本地时区按日聚合
	// respond(c, 200, UsageTrendResp{Points: buckets})
}
