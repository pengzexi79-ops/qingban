package model

// AI 通讯录实体:companions 表 + 各配置子结构。
// 字段全集对齐 BACKEND_HANDOFF.md Companion;展示派生字段(memoryCount/unread/pinned/online)
// 不入库,由 service 组装(见 server/contacts.go)。

import "time"

// Persona:角色人设(基础资料 Tab,全部为自由文本;注入系统提示)。
type Persona struct {
	// Identity:角色身份(≤500,如"27 岁的温柔电台主播")。
	Identity string `json:"identity"`
	// Relationship:与用户的关系(≤200,如"从小一起长大的青梅竹马")。
	Relationship string `json:"relationship"`
	// Personality:性格描述(≤2000)。
	Personality string `json:"personality"`
	// SpeakingStyle:表达风格(≤1000,口头禅/语气/句式)。
	SpeakingStyle string `json:"speakingStyle"`
	// Boundaries:关系与安全边界(≤1000,哪些话题不越界、不替代专业建议等)。
	Boundaries string `json:"boundaries"`
	// ForbiddenTopics:禁用内容与行为(≤1000,如不讨论的禁忌、不自称人类等)。
	ForbiddenTopics string `json:"forbiddenTopics"`
}

// MemorySettings:长期记忆设置(Tab:长期记忆)。
type MemorySettings struct {
	// Enabled:是否启用长期记忆召回(默认 true)。
	Enabled bool `json:"enabled"`
	// Mode:记忆写入模式:hybrid(候选+确认)/curated(仅手动)/automatic(自动入库),默认 hybrid。
	Mode string `json:"mode"`
	// SummaryMode:摘要模式:rolling/session/manual,默认 rolling。
	SummaryMode string `json:"summaryMode"`
	// TimeRangeDays:召回时间范围(默认 365 天)。
	TimeRangeDays int `json:"timeRangeDays"`
	// SearchThreshold:召回相似度阈值(默认 0.65)。
	SearchThreshold float64 `json:"searchThreshold"`
	// MaxItems:每次最多载入记忆条数(默认 12)。
	MaxItems int `json:"maxItems"`
}

// ChatStyle:聊天交互样式(Tab:消息与风格;前端渲染参数 + 后端流式/延时行为参考)。
type ChatStyle struct {
	// Markdown:是否渲染 markdown。
	Markdown bool `json:"markdown"`
	// Streaming:是否流式输出(为 false 时后端也应一次性返回,但保留字段驱动前端)。
	Streaming bool `json:"streaming"`
	// Typing:回复前是否显示"输入中"状态。
	Typing bool `json:"typing"`
	// SplitMessages:是否按段拆分消息气泡(前端渲染)。
	SplitMessages bool `json:"splitMessages"`
	// ReplyDelay:回复延时 ms(前端渲染节奏,默认 650)。
	ReplyDelay int `json:"replyDelay"`
	// BubbleStyle:气泡风格(soft/standard,前端渲染)。
	BubbleStyle string `json:"bubbleStyle"`
}

// ProactiveConfig:主动陪伴规则(主动消息任务属第二阶段;字段随角色配置一并持久化)。
type ProactiveConfig struct {
	// Enabled:该角色主动消息开关(默认 true;第二阶段的调度器逐层校验用户总开关+本开关)。
	Enabled bool `json:"enabled"`
	// Start:时间窗开始 "09:00"。
	Start string `json:"start"`
	// End:时间窗结束 "22:30"。
	End string `json:"end"`
	// Frequency:频率档位 light/balanced/daily/off,默认 balanced。
	Frequency string `json:"frequency"`
	// MinMinutes:两次主动消息最小间隔(默认 45 分钟)。
	MinMinutes int `json:"minMinutes"`
	// MaxMinutes:最大间隔(默认 240 分钟)。
	MaxMinutes int `json:"maxMinutes"`
	// DailyLimit:每日上限(默认 4 条)。
	DailyLimit int `json:"dailyLimit"`
	// AvoidBusyTime:是否规避忙碌时段(默认 true,结合用户最近互动判断)。
	AvoidBusyTime bool `json:"avoidBusyTime"`
}

// Companion:AI 好友/角色(数据通讯录主实体)。
// 表:companions。嵌套配置子对象全部以 TEXT JSON 列持久化(serializer:json)。
type Companion struct {
	// ID:角色 id(后端生成,可读前缀 companion-,如 companion-mumu)。
	ID string `json:"id" gorm:"primaryKey"`
	// Name:角色名(≤12)。
	Name string `json:"name"`
	// Initial:文字头像字(≤1,缺省取 name 首字符)。
	Initial string `json:"initial"`
	// Color:主题色(#RRGGBB,如 #776ee8)。
	Color string `json:"color"`
	// AvatarFileID:头像文件引用(/files 上传;可选)。
	AvatarFileID *string `json:"avatarFileId,omitempty" gorm:"column:avatar_file_id"`
	// AvatarImage:base64 头像(演示数据导入兼容,迁移后废弃)。
	AvatarImage *string `json:"avatarImage,omitempty" gorm:"column:avatar_image"`
	// Category:分类(温柔陪伴/知心朋友/元气搭子/安静倾听/灵感导师 等,自由字符串)。
	Category string `json:"category"`
	// Tagline:一句话介绍(≤40,列表副标题)。
	Tagline string `json:"tagline"`
	// APIProfileID:绑定的 API 配置 id(为空=回落默认本地配置)。
	APIProfileID *string `json:"apiProfileId,omitempty" gorm:"column:api_profile_id;index"`
	// Persona:角色人设(Tab 分段更新)。
	Persona Persona `json:"persona" gorm:"type:text;serializer:json"`
	// MemorySettings:长期记忆设置。
	MemorySettings MemorySettings `json:"memorySettings" gorm:"type:text;serializer:json"`
	// ChatStyle:消息与风格设置。
	ChatStyle ChatStyle `json:"chatStyle" gorm:"type:text;serializer:json"`
	// Proactive:主动陪伴设置。
	Proactive ProactiveConfig `json:"proactive" gorm:"type:text;serializer:json"`
	// Capabilities:该角色能力开关(是否真实可用由后端验证)。
	Capabilities Capabilities `json:"capabilities" gorm:"type:text;serializer:json"`
	// CreatedAt/UpdatedAt:创建/更新时间。
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`

	// 下方为"列表展示派生字段"(gorm:"-" 不入库),由 List 服务填充:
	// MemoryCount:该角色记忆条数(服务端统计)。
	MemoryCount int `json:"memoryCount,omitempty" gorm:"-"`
	// Unread:未读数(会话缓存表统计)。
	Unread int `json:"unread,omitempty" gorm:"-"`
	// Pinned:会话是否置顶(缓存表读取)。
	Pinned bool `json:"pinned,omitempty" gorm:"-"`
	// Online:可用状态(本地形态恒为 true 占位;第二阶段接模型可用性)。
	Online bool `json:"online,omitempty" gorm:"-"`
}

// TableName:表名(companions)。
func (Companion) TableName() string { return "companions" }
