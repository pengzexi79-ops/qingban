package model

import "gorm.io/gorm"

// Message 聊天消息(单聊与群聊统一实体)
// 定位:以"会话归属 + 文本内容 + 关联文件"为最小模型(用户基准版);
// 会话/未读/历史等维度经 ConversationID 挂在会话实体上,消息本身不做展示缓存。
// 附件采用"双轨":①content 保留前端引用标记语法(见 utils.ParseRefs);
// ②Files 多对多(message_files)落库关联,渲染与清理以多对多为准。
// 说明:若未来需要区分单/群与消息方向,可加 conversation_type/role 扩展列(当前由会话实体与 SenderID 表达)。
type Message struct {
	gorm.Model
	// ConversationID 归属会话 id(单聊=companionId,群聊=groupId);历史/未读/删除按此查询。
	ConversationID string `gorm:"size:64;index;not null" json:"conversation_id"`
	// Role 消息方向:user(用户发送)/assistant(AI 发送);未读数据此判断是否+1。
	Role string `gorm:"size:16;index;default:user" json:"role"`
	// SenderID 发送者:单聊助手消息=companionId,群聊=发言角色 id,用户消息=空(本地单用户)。
	SenderID string `gorm:"size:64;index" json:"sender_id"`
	// Content 消息内容(纯文本,或含 ![图](fileID)/[文件名](fileID) 引用标记)。
	Content string `gorm:"type:text;not null" json:"content"`
	// ReplyToID 引用的消息ID(回复消息;可空)。
	ReplyToID *uint `gorm:"index" json:"reply_to_id"`
	// ReplyTo 引用的消息(自关联)。
	ReplyTo *Message `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
	// FileIDs 关联文件ID列表(仅响应辅助,不落库)。
	FileIDs []uint `gorm:"-" json:"file_ids,omitempty"`
	// Files 关联的文件(多对多,落库于 message_files)。
	Files []File `gorm:"many2many:message_files" json:"files,omitempty"`
}

// File 文件元数据
// 定位:本地附件(消息图片/附件/头像)的通用登记;二进制存放于数据目录 files/{id},
// Path 存相对文件名(以 ID 为名),具体用途(消息/头像)由引用方关系表达。
type File struct {
	gorm.Model
	// FileName 原始文件名(展示与下载名)。
	FileName string `gorm:"size:255;not null" json:"file_name"`
	// FileType 文件类型(MIME 或扩展名;读取响应 Content-Type 依据)。
	FileType string `gorm:"size:50;not null" json:"file_type"`
	// Size 文件大小(字节)。
	Size int64 `gorm:"not null" json:"size"`
	// Path 存储路径(数据目录下的相对路径,如 files/12 或原名哈希)。
	Path string `gorm:"size:500;not null" json:"path"`
}

type 群聊消息是否已读 struct {
	ID     int64  `gorm:"primary_key"` //引用的群聊消息ID
	bitmap []byte //xxx已读(使用可变位图表示)
}
