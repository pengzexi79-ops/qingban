package model

// 短期记忆(滚动压缩的工作记忆):表 chat_short_memories。
// 语义(见 docs/batch_dispatch_design.md §5):AI 成员"消费即归档"的上下文压缩副本,
// 一成员一会话一行、覆盖更新,不参与语义召回(长期记忆走 memories)。
// 复合主键 (companion_id, conversation_id),均为数值外键。

import "time"

// ShortMemory:某 AI 成员在某会话内的最新一期短期记忆。
type ShortMemory struct {
	// CompanionID:归属 AI 成员(companions.id)。
	CompanionID uint `gorm:"primaryKey;not null" json:"companion_id"`
	// ConversationID:归属会话(conversations.id)。
	ConversationID uint `gorm:"primaryKey;not null" json:"conversation_id"`
	// Content:本期压缩结果。
	Content string `gorm:"type:text;not null" json:"content"`
	// Generation:压缩期数(每次覆盖 +1)。
	Generation int `gorm:"not null" json:"generation"`
	// CoveredFromMessageID:覆盖的源消息起点(含;0=上一期续压无新增源消息)。
	CoveredFromMessageID uint `gorm:"not null;default:0" json:"covered_from_message_id"`
	// CoveredToMessageID:覆盖的源消息终点(含);下一期从本值 +1 起压。
	CoveredToMessageID uint `gorm:"not null;default:0" json:"covered_to_message_id"`
	// RawChars:源文本总字数(公式输入)。
	RawChars int `gorm:"not null" json:"raw_chars"`
	// TargetChars:字数公式目标字数(公式输出)。
	TargetChars int `gorm:"not null" json:"target_chars"`
	// DateFactor:公式日期因子快照。
	DateFactor float64 `gorm:"type:real;not null" json:"date_factor"`
	// CharRatio:换算因子快照(取角色配置 dispatch_settings.char_ratio)。
	CharRatio float64 `gorm:"type:real;not null" json:"char_ratio"`
	// SummarizedAt:本期压缩完成时刻。
	SummarizedAt time.Time `json:"summarized_at"`
	// CreatedAt/UpdatedAt:行时间。
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// 关联(父级删除时本行随删):
	Companion    *Companion    `json:"-" gorm:"foreignKey:CompanionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Conversation *Conversation `json:"-" gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
