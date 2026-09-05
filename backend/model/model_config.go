package model

// 模型/接口配置:model_configs —— 配置的基准实体(取代早期 ApiProfile)。
// 一条记录 = 一个可被角色绑定的"模型入口"(连接信息 + 请求参数 + 能力声明)。
// 能力不擅长的场景走"子配置委托":子能力配置行也是普通行,以 parent 指向主配置行;
// 编排语义:如文字模型遇图,调度层找其图片子配置行执行理解,再把结果回传文字模型续写。
// 约束:APIKey 仅密文落库(utils.SecretBox);读取/导出永不回明文。

import "gorm.io/gorm"

// ModelConfig:模型(API)配置行。
type ModelConfig struct {
	gorm.Model
	// Name:配置唯一名(角色绑定引用)。
	Name string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	// DisplayName:显示名(设置页展示)。
	DisplayName string `gorm:"size:200" json:"display_name"`
	// Description:描述(备注用途)。
	Description string `gorm:"type:text" json:"description"`
	// Version:模型版本(可选,如 7b/1.5)。
	Version string `gorm:"size:50" json:"version"`
	// ParentID:父配置(model_configs.id,自引用;子能力配置行指向主配置行,可空)。
	ParentID *uint `gorm:"index" json:"parent_id,omitempty"`

	// ---- 连接配置 ----
	// BaseURI:接口基础地址(如 http://localhost:11434/v1)。
	BaseURI string `gorm:"size:500" json:"base_uri"`
	// APIKey:接口密钥(密文保存;仅进程内解密使用)。
	APIKey string `gorm:"size:500" json:"-"`
	// APIType:协议类型(openai/anthropic/gemini/ollama,默认 openai)。
	APIType string `gorm:"size:50;default:openai" json:"api_type"`

	// ---- 请求参数(高级设置) ----
	Temperature      float64 `gorm:"type:real;default:0.7" json:"temperature"`
	TopP             float64 `gorm:"type:real;default:1.0" json:"top_p"`
	MaxTokens        int     `gorm:"default:2048" json:"max_tokens"`
	FrequencyPenalty float64 `gorm:"type:real;default:0" json:"frequency_penalty"`
	PresencePenalty  float64 `gorm:"type:real;default:0" json:"presence_penalty"`

	// ---- 模型输出能力开关(本配置支持的模态) ----
	TextCompletion     bool `gorm:"not null;default:false" json:"text_completion"`
	PhotoGeneration    bool `gorm:"not null;default:false" json:"photo_generation"`
	VideoGeneration    bool `gorm:"not null;default:false" json:"video_generation"`
	AudioGeneration    bool `gorm:"not null;default:false" json:"audio_generation"`
	TextToSpeech       bool `gorm:"not null;default:false" json:"text_to_speech"`
	ImageUnderstanding bool `gorm:"not null;default:false" json:"image_understanding"`
	VideoUnderstanding bool `gorm:"not null;default:false" json:"video_understanding"`
	AudioUnderstanding bool `gorm:"not null;default:false" json:"audio_understanding"`

	// Parent:父配置行(子行委托主行;主行被删时子行脱离,变独立配置)。
	Parent *ModelConfig `json:"-" gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
