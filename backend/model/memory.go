package model

// 长期记忆实体:memories 表。
// 数据台语义(BACKEND_HANDOFF §2 + PHASE1_API §P3):记忆可查看/编辑/删除/确认候选;
// 删除后不得再被检索与注入;向量索引与 SQLite 行分离存储(见 init 的建表注释)。

import "time"

// MemoryType:记忆类型枚举(数据台与语义检索的类型维度)。
const (
	// MemPreference:用户偏好(喜欢雨天听歌……)。
	MemPreference = "preference"
	// MemEvent:关系事件/经历(一起做过什么、重要日子……)。
	MemEvent = "event"
	// MemRelationship:关系记忆(双方关系状态、称呼、相处模式……)。
	MemRelationship = "relationship"
	// MemSummary:对话摘要(滚动/会话级总结产物)。
	MemSummary = "summary"
)

// MemoryStatus:记忆状态(自动提取→候选→用户确认)。
const (
	// MemStatusCandidate:自动提取候选,等待用户确认(仅确认后参与注入/可编辑)。
	MemStatusCandidate = "candidate"
	// MemStatusConfirmed:已确认(手动新增默认 confirmed)。
	MemStatusConfirmed = "confirmed"
)

// EmbeddingStatus:向量索引状态。
const (
	// EmbedPending:待索引(新增/编辑后)。
	EmbedPending = "pending"
	// EmbedIndexed:已索引(参与语义检索)。
	EmbedIndexed = "indexed"
	// EmbedFailed:索引失败(embedding 模型不可用/超时),可 reindex 重试。
	EmbedFailed = "failed"
)

// Memory:长期记忆行。表:memories。
type Memory struct {
	// ID:记忆 id(可读前缀 memory-)。
	ID string `json:"id" gorm:"primaryKey"`
	// CompanionID:归属角色 id;空=全局记忆(用户偏好,所有角色可召回)。
	// 说明:表内冗余归属,数据台的全局/单角色过滤基于本字段。
	CompanionID *string `json:"companionId,omitempty" gorm:"column:companion_id;index"`
	// Type:preference/event/relationship/summary。
	Type string `json:"type" gorm:"index"`
	// Title:标题(≤28,数据台列表主文案)。
	Title string `json:"title"`
	// Content:记忆正文(≤5000,标准化后的内容)。
	Content string `json:"content" gorm:"type:text"`
	// Date:归属日期(YYYY-MM-DD;导入兼容与时间线排序)。
	Date string `json:"date" gorm:"column:date"`
	// Source:来源说明(≤50,如"手动添加 / 你告诉我的 / 自动提取")。
	Source string `json:"source"`
	// Importance:重要度 0~1(默认 0.5;候选提取时由模型给出,影响注入排序)。
	Importance float64 `json:"importance"`
	// Status:candidate/confirmed。
	Status string `json:"status" gorm:"index"`
	// EmbeddingStatus:pending/indexed/failed。
	EmbeddingStatus string `json:"embeddingStatus" gorm:"column:embedding_status"`
	// SourceMessageID:提取来源消息(候选提取时填;审计/回跳用)。
	SourceMessageID *string `json:"sourceMessageId,omitempty" gorm:"column:source_message_id"`
	// CreatedAt/UpdatedAt:创建/更新时间(数据台默认按 date+updated 倒序)。
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

// TableName:表名(memories)。
func (Memory) TableName() string { return "memories" }

// MemoryHit:记忆检索命中项(/memories/search hits[] 元素)。
type MemoryHit struct {
	// MemoryID:命中记忆 id。
	MemoryID string `json:"memoryId"`
	// Score:相关度 0~1(vector 为余弦相似度,keyword 为 FTS 折算分)。
	Score float64 `json:"score"`
	// Method:vector(向量命中)/keyword(FTS5 降级命中)。
	Method string `json:"method"`
	// Title:标题(列表回显)。
	Title string `json:"title"`
	// Content:内容(注入前可截断展示)。
	Content string `json:"content"`
	// Type:记忆类型。
	Type string `json:"type"`
	// CompanionID:归属角色(可为空=全局记忆)。
	CompanionID *string `json:"companionId,omitempty"`
}

// MemoryDraft:记忆候选草稿(自动提取产出,入库前形态;见 AI/candidates.go)。
// 作用:①流式 done 事件与同步响应中携带(前端先渲染"待确认"卡片);
//
//	②mode=automatic 时直接转 Memory 落库。
type MemoryDraft struct {
	// Type:提取出的记忆类型。
	Type string `json:"type"`
	// Title:候选标题。
	Title string `json:"title"`
	// Content:候选正文。
	Content string `json:"content"`
	// Importance:候选重要度。
	Importance float64 `json:"importance"`
	// SourceMessageID:来源消息。
	SourceMessageID string `json:"sourceMessageId,omitempty"`
}
