package model

// 用量账本:usage_records。每次模型调用真实落一条(本地记账/诊断对账)。
// 注:结构字段 Model 与嵌入类型名冲突,故不整体嵌入 gorm.Model,
// 时间戳/软删字段显式声明(主键仍为数字自增)。

import (
	"time"

	"gorm.io/gorm"
)

// CapabilityName:能力维度标识(记录调用的是哪类模型能力)。
const (
	CapChat       = "chat"       // 对话
	CapVision     = "vision"     // 视觉理解
	CapHearing    = "hearing"    // 听觉/转写
	CapTTS        = "tts"        // 语音合成
	CapVoiceClone = "voiceClone" // 声音复刻
	CapVideo      = "video"      // 视频理解
	CapImage      = "image"      // 文生图
	CapEmbedding  = "embedding"  // 向量化
)

// UsageRecord:一次模型调用的用量明细。
type UsageRecord struct {
	// ID:主键(自增)。
	ID uint `json:"id" gorm:"primaryKey"`
	// CreatedAt:记录时间(按日聚合的分桶依据;建议迁移补 created_at 索引)。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt:最近更新。
	UpdatedAt time.Time `json:"-"`
	// DeletedAt:软删标记。
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	// Provider:服务商标识(openai/anthropic/ollama/qwen/custom)。
	Provider string `json:"provider" gorm:"size:32;not null"`
	// Model:实际调用模型 id。
	Model string `json:"model" gorm:"size:128;not null"`
	// Capability:chat/vision/hearing/tts/voiceClone/video/image/embedding。
	Capability string `json:"capability" gorm:"size:24;index"`
	// ConversationID:归属会话(conversations.id;可空=能力调用)。
	ConversationID *uint `json:"conversation_id,omitempty" gorm:"index"`
	// CompanionID:服务角色(companions.id;可空)。
	CompanionID *uint `json:"companion_id,omitempty" gorm:"index"`
	// ModelConfigID:使用的模型配置(model_configs.id;可空)。
	ModelConfigID *uint `json:"model_config_id,omitempty" gorm:"index"`
	// InputTokens/OutputTokens/CachedTokens:输入/输出/缓存 token。
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	CachedTokens int `json:"cached_tokens"`
	// LatencyMs:调用总耗时毫秒。
	LatencyMs int64 `json:"latency_ms"`
	// Status:success/failed(失败也记录,便于诊断)。
	Status string `json:"status" gorm:"size:16;not null"`
	// EstimatedCost:估算费用(仅远程计费服务商,Ollama 为 0)。
	EstimatedCost float64 `json:"estimated_cost" gorm:"type:real;default:0"`
	// ProviderRequestID:供应商侧请求 id(排障)。
	ProviderRequestID string `json:"-" gorm:"size:255"`
	// ErrorCode:失败类目(如 PROVIDER_ERROR/超时;成功为空)。
	ErrorCode string `json:"-" gorm:"size:64"`
}
