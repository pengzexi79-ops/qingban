package model

// 群聊实体:groups / group_members / rounds 三表。
// 会话语义:群聊的 conversationId 即 group.id(与单聊共用同一 id 命名空间,见 openapi 契约说明)。

import "time"

// GroupStrategy:群聊调度策略(触发一轮时由 AI 包读取执行)。
type GroupStrategy struct {
	// Enabled:允许主动开一轮话题(主动任务阶段生效;手动"触发一轮"不受此开关限制)。
	Enabled bool `json:"enabled"`
	// Mode:发言成员选择模式:random(随机)/turn(轮流),默认 random。
	Mode string `json:"mode"`
	// CooldownSeconds:两轮之间的最小间隔秒(默认 30)。
	CooldownSeconds int `json:"cooldownSeconds"`
	// MaxSpeakers:每轮最多发言成员数(默认 2)。
	MaxSpeakers int `json:"maxSpeakers"`
	// Order:发言顺序:balanced(均衡,优先发言少者)/member-order(按成员顺序),默认 member-order。
	Order string `json:"order"`
}

// Group:AI 群聊。表:groups。
type Group struct {
	// ID:群 id(可读前缀 group-;同时作为该群会话的 conversationId)。
	ID string `json:"id" gorm:"primaryKey"`
	// Name:群名(≤20)。
	Name string `json:"name"`
	// AvatarFileID:群头像文件引用(可选)。
	AvatarFileID *string `json:"avatarFileId,omitempty" gorm:"column:avatar_file_id"`
	// Color/Initial:文字头像展示(前端渲染;可选)。
	Color   string `json:"color"`
	Initial string `json:"initial"`
	// Announcement:群公告/群聊目的与调度说明(≤500)。
	Announcement *string `json:"announcement" gorm:"column:announcement"`
	// Strategy:调度策略(JSON 列)。
	Strategy GroupStrategy `json:"strategy" gorm:"type:text;serializer:json"`
	// LastRoundAt:最近一次轮次触发时刻(冷却判断用;本阶段由 AI 包维护)。
	LastRoundAt *time.Time `json:"-" gorm:"column:last_round_at"`
	// CreatedAt/UpdatedAt:创建/更新时间。
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`

	// MemberIDs:群成员 id 列表(不入库;读取时联查 group_members 组装,由 service 填充)。
	MemberIDs []string `json:"memberIds" gorm:"-"`
}

// TableName:表名(groups)。
func (Group) TableName() string { return "groups" }

// GroupMember:群成员关系行(表:group_members,一对多)。
// 字段作用:记录某角色何时加入某群(将来用于排序/活跃度统计与移除)。
type GroupMember struct {
	// GroupID:群 id(复合主键 1)。
	GroupID string `json:"groupId" gorm:"primaryKey"`
	// CompanionID:角色 id(复合主键 2)。
	CompanionID string `json:"companionId" gorm:"primaryKey"`
	// JoinedAt:入群时刻。
	JoinedAt time.Time `json:"joinedAt"`
}

// TableName:表名(group_members)。
func (GroupMember) TableName() string { return "group_members" }

// RoundStatus:轮次状态常量。
const (
	// RoundRunning:轮次执行中。
	RoundRunning = "running"
	// RoundCompleted:轮次正常完成(达到 maxSpeakers)。
	RoundCompleted = "completed"
	// RoundCancelled:轮次被取消(冷却校验失败/成员不足/用户取消/过程错误)。
	RoundCancelled = "cancelled"
)

// RoundSpeaker:轮次中某成员的发言记录(序列化在 round.speakers JSON 内)。
type RoundSpeaker struct {
	// CompanionID:发言角色。
	CompanionID string `json:"companionId"`
	// MessageID:该成员本轮产生的消息 id(便于跳转与审计)。
	MessageID string `json:"messageId"`
}

// Round:群聊轮次记录(表:rounds)。第一阶段"同步执行一轮",明细同时落库供第二阶段做异步展示。
type Round struct {
	// RoundID:轮次唯一 id(可读前缀 round-)。
	RoundID string `json:"roundId" gorm:"primaryKey;column:round_id"`
	// GroupID:所属群。
	GroupID string `json:"groupId" gorm:"column:group_id;index"`
	// TriggeredAt:触发时刻(冷却判断依据:距上轮 triggeredAt ≥ cooldownSeconds)。
	TriggeredAt time.Time `json:"triggeredAt" gorm:"column:triggered_at"`
	// EndedAt:结束时刻(为空=尚未结束)。
	EndedAt *time.Time `json:"endedAt,omitempty" gorm:"column:ended_at"`
	// Status:running/completed/cancelled。
	Status string `json:"status"`
	// CancelReason:取消原因(仅 cancelled 有意义,如 COOLDOWN_ACTIVE)。
	CancelReason *string `json:"cancelReason,omitempty" gorm:"column:cancel_reason"`
	// Speakers:本轮成员发言序列(JSON 列)。
	Speakers []RoundSpeaker `json:"speakers" gorm:"type:text;serializer:json"`
}

// TableName:表名(rounds)。
func (Round) TableName() string { return "rounds" }
