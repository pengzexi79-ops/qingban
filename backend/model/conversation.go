package model

// 会话展示缓存:conversations 表。
// 定位(见 PHASE1_API.md §4 表说明):会话列表主数据(名称/头像/未读/置顶)来自
// companions/groups 实体,本表只存"展示摘要派生缓存":单机本地无一致性成本,
// 每次消息写入/已读/置顶时同步维护,避免列表页每次做大量 join 聚合。

import "time"

// Conversation:会话展示缓存行(表:conversations)。
// 主键即 conversationId:单聊=companionId,群聊=groupId(两实体删除时本行级联删除)。
type Conversation struct {
	// ID:会话 id(=companionId 或 groupId)。
	ID string `json:"id" gorm:"primaryKey"`
	// Kind:companion/group(冗余,列表聚合过滤用)。
	Kind string `json:"type" gorm:"column:kind;index"`
	// Pinned:是否置顶(置顶在前,按 last_active_at 倒序)。
	Pinned bool `json:"pinned"`
	// Muted:免打扰(消息落库/轮次广播时是否抑制前端角标提示)。
	Muted bool `json:"muted"`
	// Unread:未读数(标记已读时清零;写消息时对方消息 +1,自己的消息不 +1)。
	Unread int `json:"unread"`
	// LastActiveAt:最后一条消息时间(=列表排序键 lastTimestamp)。
	LastActiveAt time.Time `json:"lastTimestamp" gorm:"column:last_active_at;index"`
	// LastMessageID:最后一条消息 id(可选,用于点击跳转)。
	LastMessageID string `json:"-" gorm:"column:last_message_id"`
	// LastContent:最后一条消息纯文本摘要(截断 ≤120 字;群聊带发送者前缀,如 "沐沐: 今天……")。
	LastContent string `json:"-" gorm:"column:last_content"`
	// LastSenderName:最后消息发送者名(群聊展示用)。
	LastSenderName *string `json:"-" gorm:"column:last_sender_name"`
}

// TableName:表名(conversations)。
func (Conversation) TableName() string { return "conversations" }

// LastMessageBrief:会话列表的"最后消息摘要"响应结构(Thread.lastMessage,供前端展示)。
type LastMessageBrief struct {
	// Content:摘要文本(已截断/已去引用标记)。
	Content string `json:"content"`
	// Timestamp:消息时间。
	Timestamp time.Time `json:"timestamp"`
	// SenderName:发送者名(群聊非空;单聊可为空或助手名)。
	SenderName *string `json:"senderName,omitempty"`
}
