package model

// 成员已读消费水位:member_cursors。
// 语义(见 docs/batch_dispatch_design.md §2.1):人侧未读用 conversations.unread +
// messages.user_read_at;AI 侧以"接收方消费水位"表达——(会话, 接收方) 记录读到哪条。
// 某成员"未读消息" = 会话内 id > last_read_message_id 且非自己发送;批次投喂后游标前移,
// 已读消息永不重投(只进短期记忆)。群聊每 AI 成员一行;重启可从库恢复水位。
// 读取方:AI 成员 = companions.id;用户 = 保留值 0(本地单用户)。

import "time"

// ReaderUserID:用户作为接收方时的 reader_id 保留值(与 companions.id ≥1 不冲突)。
// 注:本列不对 companions 建外键约束(保留值 0 无对应行);引用完整性由服务层保证。
const ReaderUserID = uint(0)

// MemberCursor:某会话内某接收方(人/AI 成员)的已读消费水位。
type MemberCursor struct {
	// ConversationID:归属会话(conversations.id)。
	ConversationID uint `json:"conversation_id" gorm:"primaryKey;not null"`
	// ReaderID:接收方(companions.id;用户=ReaderUserID)。
	ReaderID uint `json:"reader_id" gorm:"primaryKey;not null"`
	// LastReadMessageID:已消费到的最后一条消息 id(含;0=一条未读)。
	LastReadMessageID uint `json:"last_read_message_id" gorm:"not null;default:0"`
	// UpdatedAt:水位最近推进时刻。
	UpdatedAt time.Time `json:"updated_at"`

	// 关联(会话删则水位随删):
	Conversation *Conversation `json:"-" gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
