package model

// 成员已读消费水位:member_cursors 表。
// 语义(见 docs/batch_dispatch_design.md §2.1):每条消息对"人"和"AI"两侧各维护已读状态。
//   - 人侧:会话级未读计数在 conversations.unread + 消息级 user_read_at(见 Message);
//   - AI 侧:以"接收方消费水位游标"表达——(会话, 接收方) 记录该接收方已读到哪条消息。
//
// 消费语义:某成员的"未读消息" = 本会话内 id > last_read_message_id 且非该成员自己发送的消息;
// 一次批次投喂(或用户标记已读)后游标前移,已读消息永不重投(只进短期记忆,见 chat_short_memories)。
// 群聊中每个 AI 成员各自独立一行:成员 A 的发言对成员 B 而言仍是 B 的未读。
// 该方案替代早期"每条消息存一份 bitmap 已读行"的草稿(见 git 历史 message.go):
// 写放大更小、批次消费天然按 id 序前移、重启后可从 DB 恢复水位。

import "time"

// ReaderUser:用户作为接收方时的 reader_id 常量(本地单用户)。
const ReaderUser = "user"

// MentionEveryone:群聊"@所有人"保留点名 id(解析结果写入 Message.Mentions,
// 调度层展开为群内全部成员;避免与真实成员 id 撞名,成员 id 恒为 "companion-" 前缀)。
const MentionEveryone = "everyone"

// MemberCursor:某会话内某接收方(人 / AI 成员)的已读消费水位。表:member_cursors。
type MemberCursor struct {
	// ConversationID 归属会话 id(单聊=companionId,群聊=groupId)。
	ConversationID string `gorm:"column:conversation_id;primaryKey;size:64" json:"conversation_id"`
	// ReaderID 接收方 id:AI 成员=companionId,用户=ReaderUser。
	ReaderID string `gorm:"column:reader_id;primaryKey;size:64" json:"reader_id"`
	// LastReadMessageID 该接收方已消费到的最后一条消息 id(含本值;0=一条未读)。
	LastReadMessageID uint `gorm:"column:last_read_message_id;not null" json:"last_read_message_id"`
	// UpdatedAt 水位最近推进时刻(调度器冷却/诊断用)。
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName:表名(member_cursors)。
func (MemberCursor) TableName() string { return "member_cursors" }
