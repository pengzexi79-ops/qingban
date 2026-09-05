package model

// 消息实体:messages 表(单聊与群聊统一模型)。
// 设计约束(BACKEND_HANDOFF §2):群聊 AI 消息必须携带 senderId;消息的 conversationType
// 区分 companion/group;流式回复结束事件携带最终 messageId/usageId/记忆处理状态。
// 全文检索:阶段实现时对 content(纯文本部分)建 FTS5 虚拟表(触发器同步),见 init 注释。

import "time"

// ConversationType:会话类型(消息归属维度)。
const (
	// ConvCompanion:单聊会话(conversationId = companionId)。
	ConvCompanion = "companion"
	// ConvGroup:群聊会话(conversationId = groupId)。
	ConvGroup = "group"
)

// MsgRole:消息角色。
const (
	// RoleUser:用户发送(本地单用户,senderId 统一为 userId)。
	RoleUser = "user"
	// RoleAssistant:AI 发送(senderId 必填=发言角色/群成员)。
	RoleAssistant = "assistant"
)

// ContentType:消息内容类型(纯媒体消息 content 可为空,由 contentType 表达)。
const (
	// ContentText:纯文本。
	ContentText = "text"
	// ContentImage:整条仅图片(引用标记承载)。
	ContentImage = "image"
	// ContentFile:整条仅文件附件。
	ContentFile = "file"
	// ContentVoice:语音消息(第二阶段能力)。
	ContentVoice = "voice"
	// ContentVideo:视频消息(第二阶段能力)。
	ContentVideo = "video"
	// ContentMixed:文本混引用(多图/图文/附件混排)。
	ContentMixed = "mixed"
)

// MessageRef:消息引用的文件资源(解析 content 中 ![img](id)/[name](id) 标记得到;
// 与 content 幂等可重建,随消息一起序列化返回前端)。
type MessageRef struct {
	// Kind:image/file(语音/视频同 file,由 File 的 kind 进一步区分)。
	Kind string `json:"kind"`
	// FileID:文件资源 id(归属校验与取流凭据)。
	FileID string `json:"fileId"`
	// FileName:展示文件名(取引用标记的名字,缺省用文件原名)。
	FileName string `json:"fileName,omitempty"`
	// MimeType:文件 MIME(如 image/jpeg)。
	MimeType string `json:"mimeType,omitempty"`
	// Size:文件字节数。
	Size int64 `json:"size,omitempty"`
	// Width/Height:图片宽高(图片类才有)。
	Width  *int `json:"width,omitempty"`
	Height *int `json:"height,omitempty"`
	// ThumbnailFileID:图片缩略图文件 id(可选)。
	ThumbnailFileID *string `json:"thumbnailFileId,omitempty"`
}

// Message:消息行。表:messages。
type Message struct {
	// ID:消息 id(可读前缀 message-)。
	ID string `json:"id" gorm:"primaryKey"`
	// ConversationType:companion/group。
	ConversationType string `json:"conversationType" gorm:"column:conversation_type;index:idx_conversation_time"`
	// ConversationId:单聊=companionId,群聊=groupId。
	ConversationId string `json:"conversationId" gorm:"column:conversation_id;index:idx_conversation_time"`
	// Role:user/assistant。
	Role string `json:"role"`
	// SenderId:单聊:AI 消息=companionId,用户消息=userId;群聊:必填(AI=发言角色)。
	SenderId string `json:"senderId" gorm:"column:sender_id;index"`
	// Content:正文(≤5000,支持引用标记;纯文本用户输入 ≤500,见 utils.ParseRefs)。
	Content string `json:"content" gorm:"type:text"`
	// ContentType:text/image/file/voice/video/mixed。
	ContentType string `json:"contentType" gorm:"column:content_type"`
	// Refs:解析后的引用数组(JSON 列;服务端保证与 content 幂等)。
	Refs []MessageRef `json:"refs,omitempty" gorm:"type:text;serializer:json"`
	// Timestamp:消息时间(RFC3339;列表/翻页的排序键)。
	Timestamp time.Time `json:"timestamp" gorm:"index"`
	// Proactive:是否主动消息产生的(第二阶段任务写入;聊天链路恒 false)。
	Proactive bool `json:"proactive,omitempty"`
	// Streamed:该 AI 消息是否以流式下发(审计/统计用)。
	Streamed bool `json:"streamed,omitempty"`
	// Fallback:模型失败时的本地兜底回复标记(前端展示"模拟/兜底"状态)。
	Fallback bool `json:"fallback,omitempty"`
	// UsageID:关联用量记录 id(仅 AI 消息有)。
	UsageID *string `json:"usageId,omitempty" gorm:"column:usage_id"`
}

// TableName:表名(messages)。
func (Message) TableName() string { return "messages" }
