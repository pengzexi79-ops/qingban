package model

// 群轮次:rounds / round_speakers。
// 原设计把"本轮谁发言"整块序列化进 round.speakers JSON 列,不可查询;
// 拆成 round_speakers 每行一次发言,发言成员/产出消息均可直接关联。

import (
	"time"

	"gorm.io/gorm"
)

// RoundStatus:轮次状态常量。
const (
	// RoundRunning:轮次执行中。
	RoundRunning = "running"
	// RoundCompleted:轮次正常完成(达到 maxSpeakers)。
	RoundCompleted = "completed"
	// RoundCancelled:轮次被取消(冷却校验失败/成员不足/用户取消/过程错误)。
	RoundCancelled = "cancelled"
)

// Round:群聊轮次(表 rounds)。第一阶段"同步执行一轮",明细落库供第二阶段异步展示。
type Round struct {
	gorm.Model
	// GroupID:所属群(groups.id)。
	GroupID uint `json:"group_id" gorm:"index;not null"`
	// TriggeredAt:触发时刻(冷却判断依据)。
	TriggeredAt time.Time `json:"triggered_at"`
	// EndedAt:结束时刻(空=未结束)。
	EndedAt *time.Time `json:"ended_at,omitempty"`
	// Status:running/completed/cancelled。
	Status string `json:"status" gorm:"size:16;default:running"`
	// CancelReason:取消原因(仅 cancelled,如 COOLDOWN_ACTIVE)。
	CancelReason *string `json:"cancel_reason,omitempty"`

	// Group:所属群(群删则轮次级联删)。
	Group *Group `json:"-" gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// Speakers:本轮每次发言(拆行)。
	Speakers []RoundSpeaker `json:"speakers,omitempty" gorm:"foreignKey:RoundID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// RoundSpeaker:轮次内一次发言(表 round_speakers)。
type RoundSpeaker struct {
	gorm.Model
	// RoundID:所属轮次(rounds.id)。
	RoundID uint `json:"round_id" gorm:"index;not null"`
	// CompanionID:发言成员(companions.id)。
	CompanionID uint `json:"companion_id" gorm:"index;not null"`
	// MessageID:该成员本轮产出的消息(messages.id)。
	MessageID uint `json:"message_id" gorm:"index;not null"`

	// 关联(均级联删:轮次/成员/消息删除时对应发言行清理)。
	Round     *Round     `json:"-" gorm:"foreignKey:RoundID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Companion *Companion `json:"-" gorm:"foreignKey:CompanionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Message   *Message   `json:"-" gorm:"foreignKey:MessageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
