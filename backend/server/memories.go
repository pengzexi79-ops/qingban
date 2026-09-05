package server

// P1/P3 长期记忆端点:数据台 CRUD、语义检索、候选确认、重建索引 + 记忆文本搜索。
// 原则:记忆可查看/编辑/删除/确认;删除后不得再被检索与注入(行删 + 向量行删同步)。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// MemoryListQuery:GET /memories 查询参数(数据台)。
type MemoryListQuery struct {
	// CompanionId:按角色过滤(省略=全部角色含全局)。
	CompanionId string `form:"companionId"`
	// Q:文本关键词(标题/正文 LIKE;走 FTS 时忽略引用标记)。
	Q string `form:"q"`
	// Type:preference/event/relationship/summary 过滤。
	Type string `form:"type"`
	// Before/Limit:游标分页(默认按 date 倒序、created_at 兜底)。
	Before string `form:"before"`
	Limit  int    `form:"limit"`
}

// hListMemories:GET /memories —— 记忆数据台列表(分页)。
func hListMemories(c *gin.Context) {
	// var q MemoryListQuery; c.ShouldBindQuery(&q)
	// cur := parseCursor(q.Before, q.Limit)
	// rows := db.Find(Memory{},
	//     where: companion_id==q.CompanionId? AND type==q.Type? AND (q.Q=="" OR title LIKE OR content LIKE),
	//     order: date DESC, created_at DESC, id DESC, 分页取 limit+1)
	// respond(c, 200, pageOf(rows))                                    // 含 status/embeddingStatus 原样展示
}

// MemoryCreateReq:POST /memories 请求体。
type MemoryCreateReq struct {
	// CompanionId:归属角色;省略=全局记忆(用户偏好)。
	CompanionId *string `json:"companionId"`
	// Type:记忆类型(必填枚举)。
	Type string `json:"type" binding:"required,oneof=preference event relationship summary"`
	// Title:标题(≤28,必填)。
	Title string `json:"title" binding:"required,max=28"`
	// Content:正文(≤5000,必填)。
	Content string `json:"content" binding:"required,max=5000"`
	// Date:归属日期 YYYY-MM-DD(缺省今天)。
	Date string `json:"date"`
	// Source:来源(≤50,缺省"手动添加")。
	Source string `json:"source" binding:"omitempty,max=50"`
	// Importance:重要度 0~1(缺省 0.5)。
	Importance *float64 `json:"importance"`
}

// hCreateMemory:POST /memories —— 手动添加(201)。
func hCreateMemory(c *gin.Context) {
	// var req MemoryCreateReq; if !bind(c, &req) { return }
	// if req.CompanionId != nil && !companionExists(*req.CompanionId) { respondErr(422, "角色不存在"); return }
	// importance := 0.5; if req.Importance != nil { importance = clamp(*req.Importance, 0, 1) }
	// m := Memory{ID: "memory-"+uuid4(), CompanionId: req.CompanionId, Type: req.Type,
	//            Title: req.Title, Content: req.Content, Date: req.Date or today(local),
	//            Source: req.Source or "手动添加", Importance: importance,
	//            Status: confirmed, EmbeddingStatus: pending}
	// db.Insert(&m)
	// enqueueIndex(m.ID)                                               // 异步:ai.EmbedTexts→写向量行→indexed/failed(队列限并发)
	// respond(c, 201, m)
}

// MemoryUpdateReq:PATCH /memories/:memoryId 请求体(编辑或确认候选)。
type MemoryUpdateReq struct {
	Type       *string  `json:"type" binding:"omitempty,oneof=preference event relationship summary"`
	Title      *string  `json:"title" binding:"omitempty,max=28"`
	Content    *string  `json:"content" binding:"omitempty,max=5000"`
	Date       *string  `json:"date"`
	Source     *string  `json:"source" binding:"omitempty,max=50"`
	Importance *float64 `json:"importance"`
	// Status:仅允许 "confirmed"(候选→正式;其余后端忽略/拒绝)。
	Status *string `json:"status" binding:"omitempty,oneof=confirmed"`
}

// hPatchMemory:PATCH /memories/:memoryId —— 编辑 / 确认候选。
func hPatchMemory(c *gin.Context) {
	// m := db.Find(Memory{id}); if nil { 404 }
	// var req MemoryUpdateReq; if !bind(c, &req) { return }
	// if req.Status != nil && *req.Status == "confirmed" && m.Status == candidate {
	//     m.Status = confirmed                                         // 候选转正式 → 参与后续注入
	// }
	// if req.Type != nil { m.Type = *req.Type }; ... // title/content/date/source/importance 指针覆盖
	// if 任一内容字段变更 { m.EmbeddingStatus = pending; enqueueIndex(m.ID) }  // 编辑触发重建
	// m.UpdatedAt = now; db.Save(&m)
	// respond(c, 200, m)
}

// hDeleteMemory:DELETE /memories/:memoryId —— 删除并移除向量/索引。
func hDeleteMemory(c *gin.Context) {
	// m := db.Find(Memory{id}); if nil { 404 }
	// tx { db.Delete(&m); delVectorRow(m.ID); delFTSRow(m.ID) }        // 删除后不得再被检索/注入
	// respond(c, 204)
}

// MemorySearchReq:POST /memories/search 请求体。
type MemorySearchReq struct {
	// Query:检索原文(≤500,必填)。
	Query string `json:"query" binding:"required,max=500"`
	// CompanionId:限定角色(省略=全部)。
	CompanionId string `json:"companionId"`
	// TopK:返回条数(默认 8,最大 50)。
	TopK int `json:"topK" binding:"omitempty,min=1,max=50"`
	// Threshold:相似度阈值(默认 0.65)。
	Threshold float64 `json:"threshold"`
	// Include:限定记忆类型(省略=全部)。
	Include []string `json:"include" binding:"omitempty,dive,oneof=preference event relationship summary"`
}

// MemorySearchResp:POST /memories/search 响应(openapi MemorySearchResponse)。
type MemorySearchResp struct {
	// TraceId:追踪号(链路复现)。
	TraceId string `json:"traceId"`
	// Query:原样回显。
	Query string `json:"query"`
	// Hits:命中(score/method 见 MemoryHit)。
	Hits []model.MemoryHit `json:"hits"`
	// Summary:注入前压缩摘要(可为空)。
	Summary *string `json:"summary"`
}

// hSearchMemories:POST /memories/search —— 语义检索(向量优先,FTS5 降级)。
func hSearchMemories(c *gin.Context) {
	// var req MemorySearchReq; if !bind(c, &req) { return }
	// topK := req.TopK; if topK == 0 { topK = 8 }
	// res, err := ai.RecallMemories(RecallQuery{                       // 向量/降级全在 AI 包
	//     Query: req.Query, CompanionID: req.CompanionId, TopK: topK,
	//     Threshold: req.Threshold /*缺省 0.65*/, Types: req.Include, OnlyConfirmed: false})
	// if err != nil { respondErr(c, 503, CodeProviderError, "检索后端不可用"); return }
	// respond(c, 200, MemorySearchResp{traceId: res.TraceID, query: req.Query,
	//                                  hits: res.Hits, summary: &res.Summary})
}

// hReindexMemory:POST /memories/:memoryId/reindex —— 重建 embedding(pending→indexed)。
func hReindexMemory(c *gin.Context) {
	// m := db.Find(Memory{id}); if nil { 404 }
	// vec, err := ai.EmbedTexts([]string{m.Content}, "" /*模型推断*/)
	// if err != nil { m.EmbeddingStatus = failed; db.Save(&m); respondErr(503, "embedding 模型不可用"); return }
	// tx { 删旧向量行; 插新行(vec[0]); m.EmbeddingStatus = indexed; db.Save(&m) }   // 维度校验防错乱
	// respond(c, 200, m)
}

// hSearchMemoryTexts:GET /search/memories?q=(必填)—— 记忆文本搜索(数据台搜索框)。
func hSearchMemoryTexts(c *gin.Context) {
	// q := c.Query("q"); if q == "" { respondErr(422, "缺少关键词"); return }
	// page := listMemoriesCore(q: q, ...)                              // 复用 hListMemories 内核(仅 q 过滤)
	// respond(c, 200, page)
}
