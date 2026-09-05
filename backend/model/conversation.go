package model

// 会话:conversations(消息归属 + 列表状态,二合一)。
// 语义:一行 = 一段会话;单聊行 companion_id 非空且唯一,群聊行 group_id 非空且唯一,
// 二者必有其一;消息经 conversation_id 全部挂在本表;会话列表的置顶/静音/未读/
// 最后消息摘要也存本表。查询主路径:
//   会话列表 = 预载各角色/群的 Conversation;消息页 = 找会话行 → 按 conversation_id 查消息。

import (
	"time"

	"gorm.io/gorm"
)

// Conversation:会话行。
type Conversation struct {
	gorm.Model
	// CompanionID:单聊归属(companions.id,唯一)。
	CompanionID *uint `json:"companion_id,omitempty" gorm:"uniqueIndex"`
	// GroupID:群归属(groups.id,唯一)。
	GroupID *uint `json:"group_id,omitempty" gorm:"uniqueIndex"`
	// Pinned:是否置顶(置顶在前,按 last_active_at 倒序)。
	Pinned bool `json:"pinned" gorm:"not null;default:false"`
	// Muted:免打扰(是否抑制前端角标/提示)。
	Muted bool `json:"muted" gorm:"not null;default:false"`
	// Unread:未读数(对方消息 +1,自己的消息不 +1;标记已读清零)。
	Unread int `json:"unread" gorm:"not null;default:0"`
	// LastActiveAt:最后一条消息时间(=列表排序键)。
	LastActiveAt time.Time `json:"last_active_at" gorm:"index"`
	// LastMessageID:最后一条消息 id(点击跳转,可空)。
	LastMessageID *uint `json:"last_message_id,omitempty"`
	// LastContent:最后一条消息纯文本摘要(群聊带发送者前缀)。
	LastContent string `json:"-" gorm:"type:text"`
	// LastSenderName:最后消息发送者名(群聊展示用)。
	LastSenderName *string `json:"-"`

	// 关联(级联:角色/群删除 → 会话随删 → 消息等子数据随删):
	Companion *Companion `json:"-" gorm:"foreignKey:CompanionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Group     *Group     `json:"-" gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Messages  []Message  `json:"-" gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// LastMessageBrief:会话列表"最后一条消息摘要"响应结构(服务层 DTO,不落库)。
type LastMessageBrief struct {
	// Content:摘要文本(已截断/去引用标记)。
	Content string `json:"content"`
	// Timestamp:消息时间。
	Timestamp time.Time `json:"timestamp"`
	// SenderName:发送者名(群聊非空;单聊可为空或助手名)。
	SenderName *string `json:"senderName,omitempty"`
}
