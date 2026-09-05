package model

// 本地单用户实体:表 users(GORM 默认命名)。
// 语义:单机本地空间仅 1 行(bootstrap 创建,固定主键 1);主题/字号等纯展示项透传保存。
// 说明:settings 是整块"设置值"JSON 列(serializer),不拆关系表;头像引用文件登记表。

import (
	"time"

	"gorm.io/gorm"
)

// User:当前用户资料与设置。
type User struct {
	gorm.Model
	// Nickname:昵称(默认"我")。
	Nickname string `json:"nickname" gorm:"size:32"`
	// Signature:个性签名。
	Signature string `json:"signature" gorm:"size:120"`
	// UserPersona:用户画像("希望如何被理解和回应"),注入 AI 系统提示。
	UserPersona string `json:"user_persona" gorm:"size:2000"`
	// FileID:头像文件引用(files.id,可空)。
	FileID *uint `json:"file_id,omitempty" gorm:"index"`
	// Settings:设置对象(嵌套 JSON;整块存储,结构见 UserSettings)。
	Settings UserSettings `json:"settings" gorm:"type:text;serializer:json"`
}

// AdvancedSettings:全局高级请求参数(设置页"高级"分组)。
type AdvancedSettings struct {
	// ContextTurns:每次回复注入的最近对话轮数(默认 20)。
	ContextTurns int `json:"contextTurns"`
	// SummaryMode:上下文摘要模式(hybrid/rolling/session,默认 hybrid)。
	SummaryMode string `json:"summaryMode"`
	// MemoryThreshold:记忆召回相似度阈值(默认 0.65)。
	MemoryThreshold float64 `json:"memoryThreshold"`
}

// BackupSettings:云端备份设置(第三阶段生效;字段随设置保存与导入兼容)。
type BackupSettings struct {
	// Enabled:是否开启自动备份(默认 false)。
	Enabled bool `json:"enabled"`
	// IntervalHours:备份间隔小时(默认 24)。
	IntervalHours int `json:"intervalHours"`
	// LastBackupAt:最近一次备份时刻(可为空)。
	LastBackupAt *time.Time `json:"lastBackupAt,omitempty"`
}

// UserSettings:users.settings 嵌套对象。说明:后端只解释影响自身行为的子键
// (autoMessages/notifications/quietHours 供第二阶段主动消息任务使用);其余透传保存。
type UserSettings struct {
	// AutoMessages:允许 AI 主动消息(第二阶段任务执行前必须为 true)。
	AutoMessages bool `json:"autoMessages"`
	// Notifications:允许本地通知。
	Notifications bool `json:"notifications"`
	// QuietHours:启用免打扰时段。
	QuietHours bool `json:"quietHours"`
	// QuietStart:免打扰开始(如 "23:00")。
	QuietStart string `json:"quietStart"`
	// QuietEnd:免打扰结束(如 "08:00")。
	QuietEnd string `json:"quietEnd"`
	// GlobalCapabilities:用户级能力开关(透传,本阶段仅持久化)。
	GlobalCapabilities *Capabilities `json:"globalCapabilities,omitempty"`
	// MomentAutoPost:朋友圈自动发布开关(朋友圈模块推迟,字段兼容保存)。
	MomentAutoPost bool `json:"momentAutoPost"`
	// MomentFrequency:朋友圈发布频率文案(推迟模块,兼容保存)。
	MomentFrequency string `json:"momentFrequency"`
	// Advanced:高级请求参数。
	Advanced *AdvancedSettings `json:"advanced,omitempty"`
	// Backup:云端备份设置(第三阶段)。
	Backup *BackupSettings `json:"backup,omitempty"`
	// Theme:浅色/深色主题(纯前端展示,后端透传)。
	Theme string `json:"theme"`
	// FontSize:字体(compact/comfortable/large,前端透传)。
	FontSize string `json:"fontSize"`
	// BubbleRadius:气泡圆角 px(前端透传)。
	BubbleRadius int `json:"bubbleRadius"`
	// MessageGap:消息间距 px(前端透传)。
	MessageGap int `json:"messageGap"`
}
