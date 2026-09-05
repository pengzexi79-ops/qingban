package model

// 消息:messages + 附件(files/message_files)+ 点名(message_mentions)。
// 定位:消息挂会话(conversation_id),方向/发送者/已读/内容/回复引用在此;
// 附件与点名都拆成可联查的关系表,不把 id 数组塞进 JSON 列。

import (
	"time"

	"gorm.io/gorm"
)

// File:文件登记(表 files;二进制在数据目录,Path 为相对定位)。
type File struct {
	gorm.Model
	// FileName:原始文件名(展示与下载名)。
	FileName string `json:"file_name" gorm:"size:255;not null"`
	// FileType:文件类型(MIME 或扩展名)。
	FileType string `json:"file_type" gorm:"size:255;not null"`
	// Size:文件大小(字节)。
	Size int64 `json:"size" gorm:"not null"`
	// Path:存储路径(数据目录下相对路径)。
	Path string `json:"path" gorm:"size:500;not null"`
	// Width/Height:图片宽高(图片类填写,其余为空)。
	Width  *int `json:"width,omitempty"`
	Height *int `json:"height,omitempty"`
	// 说明:图片缩略图不入库,按约定以物理文件 {Path}.thumb 存放,读取时存在则用。
}

// MessageFile:消息附件关系行(复合主键;排序语义以 content 引用标记出现顺序为准)。
type MessageFile struct {
	// MessageID:消息(messages.id)。
	MessageID uint `json:"message_id" gorm:"primaryKey"`
	// FileID:文件(files.id)。
	FileID uint `json:"file_id" gorm:"primaryKey"`

	Message *Message `json:"-" gorm:"foreignKey:MessageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	File    *File    `json:"-" gorm:"foreignKey:FileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// MessageMention:点名关系行(群聊 @ 到具体成员;复合主键)。
type MessageMention struct {
	// MessageID:消息(messages.id)。
	MessageID uint `json:"message_id" gorm:"primaryKey"`
	// CompanionID:被点名成员(companions.id)。
	CompanionID uint `json:"companion_id" gorm:"primaryKey"`

	Message   *Message   `json:"-" gorm:"foreignKey:MessageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Companion *Companion `json:"-" gorm:"foreignKey:CompanionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Message:聊天消息(单聊与群聊统一)。
type Message struct {
	gorm.Model
	// ConversationID:归属会话(conversations.id;历史/未读/删除按此查询)。
	ConversationID uint `json:"conversation_id" gorm:"index;not null"`
	// Role:user(用户发送)/assistant(AI 发送)。
	Role string `json:"role" gorm:"size:16;index;default:user;not null"`
	// SenderCompanionID:发送的 AI 成员(companions.id;用户消息为空)。
	SenderCompanionID *uint `json:"sender_companion_id,omitempty" gorm:"index"`
	// MentionAll:@全员(群聊;点名到具体成员的走 Mentions 关系)。
	MentionAll bool `json:"mention_all" gorm:"not null;default:false"`
	// UserReadAt:用户阅读该消息的时刻(空=未读;会话红点以 conversations.unread 为准)。
	UserReadAt *time.Time `json:"user_read_at,omitempty" gorm:"index"`
	// Content:正文(纯文本或含引用标记)。
	Content string `json:"content" gorm:"type:text;not null"`
	// ReplyToID:回复引用的消息(messages.id,可空)。
	ReplyToID *uint `json:"reply_to_id,omitempty" gorm:"index"`

	// 关联:
	Conversation *Conversation `json:"-" gorm:"foreignKey:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// SenderCompanion:发送成员(成员删除后置空,消息保留)。
	SenderCompanion *Companion `json:"-" gorm:"foreignKey:SenderCompanionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	// ReplyTo:被回复的消息(消息删除后引用置空)。
	ReplyTo *Message `json:"-" gorm:"foreignKey:ReplyToID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	// Mentions:点名到的成员(读多写少,预载)。
	Mentions []MessageMention `json:"mentions,omitempty" gorm:"foreignKey:MessageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Files:消息附件(多对多经 message_files)。
	Files []File `json:"files,omitempty" gorm:"many2many:message_files;joinForeignKey:MessageID;joinReferences:FileID"`
	// FileIDs:附件 id 列表(仅响应辅助,不落库)。
	FileIDs []uint `json:"file_ids,omitempty" gorm:"-"`
}
