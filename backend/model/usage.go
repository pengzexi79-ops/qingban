package model

// 用量记录实体:usage_records 表。
// 语义(PHASE1 §6/P2):每次模型调用真实落一条(本地记账,不分析也可对账);
// "API 消耗诊断"页改读本表统计,不再用前端估算。Ollama 本地模型仅记 token,费用为 0。

import "time"

// CapabilityName:用量/能力维度的标识常量(记录调用的是哪类模型能力)。
const (
	// CapChat:对话。
	CapChat = "chat"
	// CapVision:视觉理解。
	CapVision = "vision"
	// CapHearing:听觉/ASR。
	CapHearing = "hearing"
	// CapTTS:语音合成。
	CapTTS = "tts"
	// CapVoiceClone:声音复刻。
	CapVoiceClone = "voiceClone"
	// CapVideo:视频理解。
	CapVideo = "video"
	// CapImage:文生图。
	CapImage = "image"
	// CapEmbedding:向量化(记忆索引/召回)。
	CapEmbedding = "embedding"
)

// UsageRecord:一次模型调用的用量明细。表:usage_records。
type UsageRecord struct {
	// ID:用量记录 id(可读前缀 usage-;AI 消息 usageId 指向本字段)。
	ID string `json:"usageId" gorm:"primaryKey;column:id"`
	// Provider:服务商标识(openai/anthropic/ollama/qwen/custom)。
	Provider string `json:"provider"`
	// Model:实际调用模型 id。
	Model string `json:"model"`
	// Capability:chat/vision/hearing/tts/voiceClone/video/image/embedding 等。
	Capability string `json:"capability" gorm:"index"`
	// ConversationID:所属会话(单聊消息场景有值;可为空=能力调用)。
	ConversationID *string `json:"conversationId,omitempty" gorm:"column:conversation_id"`
	// CompanionID:服务角色(群聊里属于某角色调用)。
	CompanionID *string `json:"companionId,omitempty" gorm:"column:companion_id"`
	// APIProfileID:使用的 API 配置(审计与成本对账维度)。
	APIProfileID string `json:"apiProfileId" gorm:"column:api_profile_id"`
	// InputTokens/OutputTokens/CachedTokens:输入/输出/缓存 token(供应商回传;缺失为 0)。
	InputTokens  int `json:"inputTokens" gorm:"column:input_tokens"`
	OutputTokens int `json:"outputTokens" gorm:"column:output_tokens"`
	CachedTokens int `json:"cachedTokens" gorm:"column:cached_tokens"`
	// LatencyMs:调用总耗时毫秒(首字节/全量,按能力记录口径)。
	LatencyMs int64 `json:"latencyMs" gorm:"column:latency_ms"`
	// Status:success/failed(失败也记录,便于诊断与重试统计)。
	Status string `json:"status"`
	// EstimatedCost:估算费用(仅远程服务商计费,Ollama 为 0)。
	EstimatedCost float64 `json:"estimatedCost" gorm:"column:estimated_cost"`
	// ProviderRequestID:供应商侧请求 id(排障引用;失败时记错误类目)。
	ProviderRequestID string `json:"-" gorm:"column:provider_request_id"`
	// ErrorCode:失败时的类目(如 PROVIDER_ERROR/超时;成功为空)。
	ErrorCode string `json:"-" gorm:"column:error_code"`
	// CreatedAt:记录时间(按日聚合趋势的分桶依据)。
	CreatedAt time.Time `json:"createdAt" gorm:"index"`
}

// TableName:表名(usage_records)。
func (UsageRecord) TableName() string { return "usage_records" }
