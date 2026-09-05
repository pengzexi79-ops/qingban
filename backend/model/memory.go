package model

// 长期记忆:memories(向量索引表 memory_vectors 独立,迁移时原生 SQL 建)。
// 归属:角色或全局(空);角色删除时其记忆级联删除(评审结论)。
// 来源消息仅审计/回跳引用:消息删除后置空保留记忆。

import "gorm.io/gorm"

// MemoryType:记忆类型。
const (
	MemPreference   = "preference"   // 用户偏好
	MemEvent        = "event"        // 关系事件/经历
	MemRelationship = "relationship" // 关系记忆
	MemSummary      = "summary"      // 对话摘要
)

// MemoryStatus:记忆状态。
const (
	MemStatusCandidate = "candidate" // 自动提取候选,待用户确认
	MemStatusConfirmed = "confirmed" // 已确认
)

// EmbeddingStatus:向量索引状态。
const (
	EmbedPending = "pending" // 待索引
	EmbedIndexed = "indexed" // 已索引(参与语义检索)
	EmbedFailed  = "failed"  // 索引失败,可重试
)

// Memory:长期记忆行。
type Memory struct {
	gorm.Model
	// CompanionID:归属角色(companions.id;空=全局记忆)。
	CompanionID *uint `json:"companion_id,omitempty" gorm:"index"`
	// SourceMessageID:提取来源消息(messages.id;候选提取时填,可空)。
	SourceMessageID *uint `json:"source_message_id,omitempty" gorm:"index"`
	// Type:preference/event/relationship/summary。
	Type string `json:"type" gorm:"size:24;index;not null"`
	// Title:标题(数据台列表主文案)。
	Title string `json:"title" gorm:"size:120;not null"`
	// Content:记忆正文(标准化后内容)。
	Content string `json:"content" gorm:"type:text;not null"`
	// Date:归属日期(YYYY-MM-DD;时间线排序)。
	Date string `json:"date" gorm:"size:16"`
	// Source:来源说明("手动添加/你告诉我的/自动提取"等)。
	Source string `json:"source" gorm:"size:200"`
	// Status:candidate/confirmed。
	Status string `json:"status" gorm:"size:16;index;default:confirmed"`
	// EmbeddingStatus:pending/indexed/failed。
	EmbeddingStatus string `json:"embedding_status" gorm:"size:16;default:pending"`

	// 关联:
	// Companion:归属角色(级联删:角色删,其记忆随删)。
	Companion *Companion `json:"-" gorm:"foreignKey:CompanionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Message:来源消息(置空保留)。
	Message *Message `json:"-" gorm:"foreignKey:SourceMessageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// MemoryHit:记忆检索命中项(/memories/search hits[] 元素)。
type MemoryHit struct {
	// MemoryID:命中记忆 id。
	MemoryID uint `json:"memory_id"`
	// Score:相关度 0~1。
	Score float64 `json:"score"`
	// Method:vector(向量命中)/keyword(FTS5 降级命中)。
	Method string `json:"method"`
	// Title:标题(回显)。
	Title string `json:"title"`
	// Content:内容(注入前可截断展示)。
	Content string `json:"content"`
	// Type:记忆类型。
	Type string `json:"type"`
	// CompanionID:归属角色(空=全局记忆)。
	CompanionID *uint `json:"companion_id,omitempty"`
}

// MemoryDraft:记忆候选草稿(自动提取产出,入库前形态)。
// ①流式 done 事件/同步响应携带(前端渲染"待确认"卡片);②automatic 模式直接转 Memory 落库。
type MemoryDraft struct {
	// Type:提取出的记忆类型。
	Type string `json:"type"`
	// Title:候选标题。
	Title string `json:"title"`
	// Content:候选正文。
	Content string `json:"content"`
	// Importance:候选重要度。
	Importance float64 `json:"importance"`
	// SourceMessageID:来源消息 id(0=无来源)。
	SourceMessageID uint `json:"source_message_id,omitempty"`
}
