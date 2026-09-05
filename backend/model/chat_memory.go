package model

// 短期记忆:chat_short_memories 表。
// 定位(见 docs/batch_dispatch_design.md §5):AI 成员的"消费即归档"上下文压缩副本。
// 动机:厂商对相同前缀请求有 KV 缓存,过期后旧对话即失忆;本表在 KV 过期前把
// "上一期短期记忆 + 本批已读消息"压缩为新一期,下次投喂注入 system 前缀,保证连续性。
// 与长期记忆(memories)语义不同:本表是"滚动压缩的工作记忆"(一角色一会话一行,
// 覆盖更新),不参与语义召回;memories 走召回/确认流程。

import "time"

// ShortMemory:某 AI 成员在某会话内的最新一期短期记忆。表:chat_short_memories。
// 复合主键 = (companion_id, conversation_id):每个成员在单聊/每个群聊各有一行。
type ShortMemory struct {
	// CompanionID 归属 AI 成员 id(记忆所有者;群聊中每个成员各自压缩自己视角的内容)。
	CompanionID string `gorm:"column:companion_id;primaryKey;size:64" json:"companion_id"`
	// ConversationID 归属会话 id(单聊=companionId,群聊=groupId)。
	ConversationID string `gorm:"column:conversation_id;primaryKey;size:64" json:"conversation_id"`
	// Content 本期压缩结果(≤ 目标字数上限,见 AI 包 SummaryTargetChars)。
	Content string `gorm:"type:text;not null" json:"content"`
	// Generation 压缩期数(每次覆盖 +1;审计连续性与覆盖窗口)。
	Generation int `gorm:"column:generation;not null" json:"generation"`
	// CoveredFromMessageID 本记忆覆盖的源消息起点 id(含;0=上一期续压无新增源消息)。
	CoveredFromMessageID uint `gorm:"column:covered_from_message_id;not null" json:"covered_from_message_id"`
	// CoveredToMessageID 本记忆覆盖的源消息终点 id(含);下一期从本值 +1 起压。
	CoveredToMessageID uint `gorm:"column:covered_to_message_id;not null" json:"covered_to_message_id"`
	// RawChars 源文本总字数(上一期记忆 + 区间内消息;公式输入,诊断/审计)。
	RawChars int `gorm:"column:raw_chars;not null" json:"raw_chars"`
	// TargetChars 字数公式给出的目标字数(公式输出,诊断/审计)。
	TargetChars int `gorm:"column:target_chars;not null" json:"target_chars"`
	// DateFactor 公式的日期因子快照(按本批最旧消息年龄)。
	DateFactor float64 `gorm:"column:date_factor;not null" json:"date_factor"`
	// CharRatio 公式的换算因子(压缩率)快照(取角色配置 dispatch_settings.charRatio)。
	CharRatio float64 `gorm:"column:char_ratio;not null" json:"char_ratio"`
	// SummarizedAt 本期压缩完成时刻。
	SummarizedAt time.Time `gorm:"column:summarized_at" json:"summarized_at"`
	// CreatedAt/UpdatedAt 行创建/更新时间。
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`
}

// TableName:表名(chat_short_memories)。
func (ShortMemory) TableName() string { return "chat_short_memories" }
