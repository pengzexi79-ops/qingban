package model

// 本地文件实体:files 表。
// 存储布局:{core.DataDir}/files/{fileId}(二进制);缩略图:{fileId}.thumb(图片类)。
// 引用模型:消息/头像只存 fileId,删除受引用保护(409 FILE_REFERENCED,见 server/files.go)。

import "time"

// FileKind:文件大类(上传 multipart 的 kind 参数)。
const (
	// FileKindImage:图片(服务端生成缩略图)。
	FileKindImage = "image"
	// FileKindFile:普通附件。
	FileKindFile = "file"
	// FileKindVoice:语音(第二阶段能力上传)。
	FileKindVoice = "voice"
	// FileKindVideo:视频(第二阶段能力上传)。
	FileKindVideo = "video"
)

// FileScope:文件用途域(上传 multipart 的 scope 参数;孤儿清理/引用统计按域区分)。
const (
	// ScopeMessage:消息附件(可被消息引用标记引用)。
	ScopeMessage = "message"
	// ScopeMoment:朋友圈媒体(朋友圈模块推迟,字段保留)。
	ScopeMoment = "moment"
	// ScopeAvatar:头像(用户/角色/群;同一 fileId 可被多方引用)。
	ScopeAvatar = "avatar"
)

// File:本地文件行。表:files。
type File struct {
	// ID:文件 id(可读前缀 file-;即磁盘文件名与访问凭据)。
	ID string `json:"fileId" gorm:"primaryKey;column:id"`
	// Kind:image/file/voice/video(上传时声明,服务端会嗅探 MIME 复核)。
	Kind string `json:"kind"`
	// Scope:message/moment/avatar(用途域)。
	Scope string `json:"scope"`
	// OrigName:原始文件名(展示与下载名)。
	OrigName string `json:"fileName" gorm:"column:orig_name"`
	// MimeType:存储 MIME(由 mimetype 嗅探,不信任客户端头)。
	MimeType string `json:"mimeType" gorm:"column:mime_type"`
	// Size:字节数(限制:图片≤10MB,附件≤100MB,语音/视频见能力阶段)。
	Size int64 `json:"size"`
	// Width/Height:图片宽高(解码得到,图片类填充)。
	Width  *int `json:"width,omitempty"`
	Height *int `json:"height,omitempty"`
	// DurationMs:语音/视频时长毫秒(能力阶段填充)。
	DurationMs *int64 `json:"duration,omitempty" gorm:"column:duration_ms"`
	// SHA256:文件校验和(导入导出清单/完整性校验)。
	SHA256 string `json:"-" gorm:"column:sha256"`
	// ThumbFileID:图片缩略图文件 id(独立 file 行,kind=image,scope=avatar 之外不做引用计数)。
	ThumbFileID *string `json:"-" gorm:"column:thumb_file_id"`
	// CreatedAt:上传时间。
	CreatedAt time.Time `json:"createdAt"`
}

// TableName:表名(files)。
func (File) TableName() string { return "files" }

// FileRef:文件引用响应(/files 上传与消息 refs 的公共返回形态)。
type FileRef struct {
	// FileID:文件 id(资源定位:GET /api/v1/files/{fileId})。
	FileID string `json:"fileId"`
	// URL:本地访问地址(由 server 组装,如 /api/v1/files/file-xxx)。
	URL string `json:"url"`
	// ThumbnailFileID:缩略图文件 id(图片类才有)。
	ThumbnailFileID *string `json:"thumbnailFileId,omitempty"`
	// FileName:文件名(展示)。
	FileName string `json:"fileName"`
	// MimeType:图片直接渲染或附件卡片展示。
	MimeType string `json:"mimeType"`
	// Size:字节数。
	Size int64 `json:"size"`
	// Width/Height:图片宽高。
	Width  *int `json:"width,omitempty"`
	Height *int `json:"height,omitempty"`
	// Duration:语音/视频毫秒数(如有)。
	Duration *int64 `json:"duration,omitempty"`
}
