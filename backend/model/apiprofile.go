package model

// API 配置实体:api_profiles 表。
// 原则(BACKEND_HANDOFF §2/§6):服务端返回绝不携带明文密钥,只返回 secretConfigured/
// maskedKey;密钥加密落库(utils.SecretBox,AES-256-GCM);删除受"至少保留一套"保护。

import "time"

// 协议类型常量(protocol 字段;决定 AI provider 用哪种载荷格式适配)。
const (
	// ProtoOpenAI:OpenAI 兼容协议(/v1/chat/completions;Ollama 亦提供该端点,默认即此)。
	ProtoOpenAI = "openai-compatible"
	// ProtoAnthropic:Anthropic Messages 协议(第二阶段能力接入再启用)。
	ProtoAnthropic = "anthropic"
	// ProtoOllama:Ollama 原生协议(/api/chat;选用时可直连,推荐走 openai-compatible)。
	ProtoOllama = "ollama"
)

// ApiProfile:API 配置行。表:api_profiles。
type ApiProfile struct {
	// ID:配置 id(可读前缀 profile-)。
	ID string `json:"id" gorm:"primaryKey"`
	// Name:配置名(≤30,如"主对话配置")。
	Name string `json:"name"`
	// Provider:服务商标识(openai/anthropic/ollama/qwen/custom,决定 UI 分组与默认参数)。
	Provider string `json:"provider"`
	// Region:展示分组(国内/国外/本地/自定义,前端过滤用)。
	Region string `json:"region"`
	// Protocol:载荷协议(openai-compatible/anthropic/ollama)。
	Protocol string `json:"protocol"`
	// Enabled:是否可用(阶段一保留字段:模型测试通过才置 true,后续做自动路由)。
	Enabled bool `json:"enabled"`
	// BaseURL:服务端地址(如 http://localhost:11434 或 https://api.example.com/v1)。
	BaseURL string `json:"baseUrl" gorm:"column:base_url"`
	// APIKeyEnc:加密后的密钥(base64;空=无需密钥,如本地 Ollama)。永不参与 JSON 输出。
	APIKeyEnc string `json:"-" gorm:"column:api_key_enc"`
	// ChatModel/VisionModel/...:七类能力分别绑定的模型 id(可留空=该配置不支持该能力)。
	ChatModel       string `json:"chatModel" gorm:"column:chat_model"`
	VisionModel     string `json:"visionModel" gorm:"column:vision_model"`
	HearingModel    string `json:"hearingModel" gorm:"column:hearing_model"`
	TTSModel        string `json:"ttsModel" gorm:"column:tts_model"`
	VoiceCloneModel string `json:"voiceCloneModel" gorm:"column:voice_clone_model"`
	VideoModel      string `json:"videoModel" gorm:"column:video_model"`
	ImageModel      string `json:"imageModel" gorm:"column:image_model"`
	// Temperature:采样温度(0~2,默认 0.8;为 nil 表示用服务商默认)。
	Temperature *float64 `json:"temperature"`
	// CreatedAt/UpdatedAt:创建/更新时间。
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`

	// 以下为响应派生字段(gorm:"-" 不入库,由 server/api_profiles.go 组装):
	// SecretConfigured:是否已配置密钥(apiKeyEnc 非空)。
	SecretConfigured bool `json:"secretConfigured" gorm:"-"`
	// MaskedKey:掩码(如 "sk-****abcd";无密钥为 nil)。
	MaskedKey *string `json:"maskedKey,omitempty" gorm:"-"`
}

// TableName:表名(api_profiles)。
func (ApiProfile) TableName() string { return "api_profiles" }

// ModelInfo:模型目录项(GET /api-profiles/{id}/models 响应元素)。
type ModelInfo struct {
	// ID:模型 id(如 qwen2.5:7b),写入各 *Model 字段的值。
	ID string `json:"id"`
	// Name:展示名(可读;缺省同 ID)。
	Name string `json:"name,omitempty"`
	// Capabilities:按模型名/元信息推断的能力标签(chat/streaming/vision/hearing/image 等)。
	Capabilities []string `json:"capabilities,omitempty"`
	// Serving:本地是否已加载(Ollama /api/ps 语义;远程恒 true)。
	Serving bool `json:"serving,omitempty"`
}
