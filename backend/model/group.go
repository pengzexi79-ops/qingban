package model

// 群聊:groups / group_members。
// 关系:一对一会话(group_id 唯一);成员多对多经 group_members(带各自入群时间);
// 成员读取用 many2many 预载,写入(加/移成员)用 GroupMember 显式行。

import (
	"time"

	"gorm.io/gorm"
)

// GroupStrategy:群聊调度策略(触发一轮时由 AI 包读取执行)。
type GroupStrategy struct {
	// Enabled:允许主动开一轮话题(主动任务阶段生效;手动"触发一轮"不受限)。
	Enabled bool `json:"enabled"`
	// Mode:random(随机)/turn(轮流),默认 random。
	Mode string `json:"mode"`
	// CooldownSeconds:两轮最小间隔秒(默认 30)。
	CooldownSeconds int `json:"cooldownSeconds"`
	// MaxSpeakers:每轮最多发言成员数(默认 2)。
	MaxSpeakers int `json:"maxSpeakers"`
	// Order:balanced(均衡)/member-order(按成员顺序),默认 member-order。
	Order string `json:"order"`
}

// Group:AI 群聊。
type Group struct {
	gorm.Model
	// Name:群名。
	Name string `json:"name" gorm:"size:40;not null"`
	// FileID:群头像文件引用(files.id,可空)。
	FileID *uint `json:"file_id,omitempty" gorm:"index"`
	// Color/Initial:文字头像展示(可选)。
	Color   string `json:"color" gorm:"size:16"`
	Initial string `json:"initial" gorm:"size:4"`
	// Announcement:群公告/群聊目的与调度说明。
	Announcement *string `json:"announcement,omitempty" gorm:"type:text"`
	// Strategy:调度策略(JSON 配置值;冷却等运行态经 core.Mem 缓存,见 AI 引擎注释)。
	Strategy GroupStrategy `json:"strategy" gorm:"type:text;serializer:json"`

	// Conversation:该群的会话(一对一;创建群时由服务层同步建行)。
	Conversation *Conversation `json:"-" gorm:"foreignKey:GroupID"`
	// Members:群成员(读取预载用;加/移成员请写 GroupMember 行;对外输出经服务层 memberIds)。
	Members []Companion `json:"-" gorm:"many2many:group_members;"`
	// 注:早期 LastRoundAt/Rounds(轮次表)已移除——轮次为运行期易失态,见 core/cache.go;
	// 旧库遗留列 last_round_at 与 rounds/round_speakers 表由迁移脚本 DROP(见 init/app.go)。
}

// GroupMember:群成员关系行(复合主键 = 群 + 成员,带各自入群时间)。
type GroupMember struct {
	// GroupID:群(groups.id)。
	GroupID uint `json:"group_id" gorm:"primaryKey"`
	// CompanionID:成员(companions.id)。
	CompanionID uint `json:"companion_id" gorm:"primaryKey"`
	// JoinedAt:入群时刻。
	JoinedAt time.Time `json:"joined_at"`

	// 关联(级联删除:群/角色删除时关系行一并清)。
	Group     *Group     `json:"-" gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Companion *Companion `json:"-" gorm:"foreignKey:CompanionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
