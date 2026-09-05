package model

// AI 好友/角色:表 companions。
// 关系:一对一会话(conversation,companion_id 唯一);头像引用文件登记表(可空);
// 每个角色"独自引用一个 API":绑定 api_configs 行(APIConfigID,可空=回落默认配置),可各自不同;
// 人设/记忆设置/聊天风格/批次调度/主动陪伴/能力开关为"整块配置值"(serializer:json),
// 不是关系;未读/置顶/记忆条数等列表派生字段不入库(服务层组装)。

import (
	"gorm.io/gorm"
)

// Persona:角色人设(基础资料,全部自由文本,注入系统提示)。
type Persona struct {
	// Identity:角色身份(如"27 岁的温柔电台主播")。
	Identity string `json:"identity"`
	// Relationship:与用户的关系(如"从小一起长大的青梅竹马")。
	Relationship string `json:"relationship"`
	// Personality:性格描述。
	Personality string `json:"personality"`
	// SpeakingStyle:表达风格(口头禅/语气/句式)。
	SpeakingStyle string `json:"speakingStyle"`
	// Boundaries:关系与安全边界(哪些话题不越界、不替代专业建议等)。
	Boundaries string `json:"boundaries"`
	// ForbiddenTopics:禁用内容与行为。
	ForbiddenTopics string `json:"forbiddenTopics"`
}

// MemorySettings:长期记忆设置(记忆召回/确认/摘要)。
type MemorySettings struct {
	// Enabled:是否启用长期记忆召回(默认 true)。
	Enabled bool `json:"enabled"`
	// Mode:hybrid(候选+确认)/curated(仅手动)/automatic(自动入库),默认 hybrid。
	Mode string `json:"mode"`
	// SummaryMode:rolling/session/manual,默认 rolling。
	SummaryMode string `json:"summaryMode"`
	// TimeRangeDays:召回时间范围(默认 365 天)。
	TimeRangeDays int `json:"timeRangeDays"`
	// SearchThreshold:召回相似度阈值(默认 0.65)。
	SearchThreshold float64 `json:"searchThreshold"`
	// MaxItems:每次最多载入记忆条数(默认 12)。
	MaxItems int `json:"maxItems"`
}

// ChatStyle:聊天交互样式(前端渲染参数 + 后端流式/延时行为参考)。
type ChatStyle struct {
	// Markdown:是否渲染 markdown。
	Markdown bool `json:"markdown"`
	// Streaming:是否流式输出。
	Streaming bool `json:"streaming"`
	// Typing:回复前是否显示"输入中"状态。
	Typing bool `json:"typing"`
	// SplitMessages:是否按段拆分消息气泡。
	SplitMessages bool `json:"splitMessages"`
	// ReplyDelay:回复延时 ms(默认 650)。
	ReplyDelay int `json:"replyDelay"`
	// BubbleStyle:气泡风格(soft/standard)。
	BubbleStyle string `json:"bubbleStyle"`
	// AllowSilent:允许"已读不回"(模型可选择不产出回复;false 时空输出按异常重试)。
	AllowSilent bool `json:"allowSilent"`
}

// ProactiveConfig:主动陪伴规则(主动消息任务属第二阶段;随角色配置持久化)。
type ProactiveConfig struct {
	// Enabled:该角色主动消息开关。
	Enabled bool `json:"enabled"`
	// Start:时间窗开始 "09:00"。
	Start string `json:"start"`
	// End:时间窗结束 "22:30"。
	End string `json:"end"`
	// Frequency:light/balanced/daily/off,默认 balanced。
	Frequency string `json:"frequency"`
	// MinMinutes:两次主动消息最小间隔(默认 45 分钟)。
	MinMinutes int `json:"minMinutes"`
	// MaxMinutes:最大间隔(默认 240 分钟)。
	MaxMinutes int `json:"maxMinutes"`
	// DailyLimit:每日上限(默认 4 条)。
	DailyLimit int `json:"dailyLimit"`
	// AvoidBusyTime:是否规避忙碌时段(默认 true)。
	AvoidBusyTime bool `json:"avoidBusyTime"`
}

// Companion:AI 好友/角色。
type Companion struct {
	gorm.Model
	// Name:角色名。
	Name string `json:"name" gorm:"size:32;not null"`
	// Initial:文字头像字(缺省取 name 首字符)。
	Initial string `json:"initial" gorm:"size:4"`
	// Color:主题色(#RRGGBB)。
	Color string `json:"color" gorm:"size:16"`
	// Category:分类(温柔陪伴/知心朋友/元气搭子/安静倾听/灵感导师 等,自由字符串)。
	Category string `json:"category" gorm:"size:50"`
	// Tagline:一句话介绍(列表副标题)。
	Tagline string `json:"tagline" gorm:"size:120"`
	// FileID:头像文件引用(files.id,可空;指针语义与 User.FileID/Group.FileID 一致)。
	FileID *uint `json:"file_id,omitempty" gorm:"index"`
	// APIConfigID:该角色独自引用的 API 配置(api_configs.id;空=回落默认配置)。
	APIConfigID *uint `json:"api_config_id,omitempty" gorm:"index"`

	// 下列为整块配置值(JSON 序列化存储,非关系):
	Persona        Persona          `json:"persona" gorm:"type:text;serializer:json"`
	MemorySettings MemorySettings   `json:"memory_settings" gorm:"type:text;serializer:json"`
	ChatStyle      ChatStyle        `json:"chat_style" gorm:"type:text;serializer:json"`
	Dispatch       DispatchSettings `json:"dispatch_settings" gorm:"type:text;serializer:json"`
	Proactive      ProactiveConfig  `json:"proactive" gorm:"type:text;serializer:json"`
	Capabilities   Capabilities     `json:"capabilities" gorm:"type:text;serializer:json"`

	// APIConfig:绑定的 API 配置行(关联;配置删除后本角色绑定置空)。
	APIConfig *APIConfig `json:"-" gorm:"foreignKey:APIConfigID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	// Conversation:该角色的单聊会话(一对一;创建角色时由服务层同步建行)。
	Conversation *Conversation `json:"-" gorm:"foreignKey:CompanionID"`

	// 派生字段(不入库,列表服务组装):
	MemoryCount int  `json:"memory_count,omitempty" gorm:"-"`
	Unread      int  `json:"unread,omitempty" gorm:"-"`
	Pinned      bool `json:"pinned,omitempty" gorm:"-"`
	Online      bool `json:"online,omitempty" gorm:"-"`
}
