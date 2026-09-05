package model

// ModelConfig 模型(API)配置 —— 青伴模型配置的基准实体。
// 语义(以本实体为基准):一条记录 = 一个可被角色(companion)绑定的"模型/API 配置",
// 等价"一个模型入口":连接信息 + 请求参数 + 能力声明;
// 能力声明(TextCompletion/PhotoGeneration/...)表达该配置"能做什么",
// 无对应能力时通过 Photo/Video/Audio 自引用指定"承担该能力的子模型配置"。
// 取代早期 ApiProfile 扁平七模型设计(见 apiprofile.go 顶部 Deprecated 注记)。
// 约束:APIKey 仅密文落库(utils.SecretBox),读取/导出永不回明文。
import "gorm.io/gorm"

type ModelConfig struct {
	gorm.Model
	// Name 配置唯一名(绑定角色用;如 companion.api_profile_id 指向本行 ID)。
	Name string `gorm:"size:100;uniqueIndex;not null;comment:配置唯一名"`
	// DisplayName 显示名(设置页展示)。
	DisplayName string `gorm:"size:200;comment:显示名称"`
	// Description 描述(备注用途)。
	Description string `gorm:"type:text;comment:模型描述"`
	// Version 模型版本(可选,如 7b/1.5)。
	Version string `gorm:"size:50;comment:模型版本"`

	// ---- 连接配置 ----
	// BaseURI API 基础地址(如 http://localhost:11434/v1 或 https://api.deepseek.com)。
	BaseURI string `gorm:"size:500;comment:API基础地址"`
	// APIKey API 密钥(密文保存;仅进程内解密使用)。
	APIKey string `gorm:"size:500;comment:API密钥(密文)"`
	// APIType 协议类型:openai/anthropic/gemini/ollama 等(默认 openai)。
	APIType string `gorm:"size:50;default:openai;comment:API类型(openai/anthropic/gemini等)"`

	// ---- 请求参数(高级设置) ----
	Temperature      float64 `gorm:"type:decimal(3,2);default:0.7;comment:温度参数"`
	TopP             float64 `gorm:"type:decimal(3,2);default:1.0;comment:Top-P采样参数"`
	MaxTokens        int     `gorm:"default:2048;comment:最大输出token数"`
	FrequencyPenalty float64 `gorm:"type:decimal(3,2);default:0;comment:频率惩罚"`
	PresencePenalty  float64 `gorm:"type:decimal(3,2);default:0;comment:存在惩罚"`

	// ---- 模型输出能力(本配置支持的模态能力开关) ----
	// TextCompletion 文本生成。
	TextCompletion bool `gorm:"default:false;comment:文本生成能力"`
	// PhotoGeneration 图像生成(命名:Photo 缩写历史拼写已修正)。
	PhotoGeneration bool `gorm:"default:false;comment:图像生成能力"`
	// VideoGeneration 视频生成。
	VideoGeneration bool `gorm:"default:false;comment:视频生成能力"`
	// AudioGeneration 音频生成。
	AudioGeneration bool `gorm:"default:false;comment:音频生成能力"`
	// TextToSpeech 文本转语音。
	TextToSpeech bool `gorm:"default:false;comment:文本转语音能力"`
	// ImageUnderstanding 图像理解。
	ImageUnderstanding bool `gorm:"default:false;comment:图像理解能力"`
	// VideoUnderstanding 视频理解。
	VideoUnderstanding bool `gorm:"default:false;comment:视频理解能力"`
	// AudioUnderstanding 音频理解。
	AudioUnderstanding bool `gorm:"default:false;comment:音频理解能力"`

	// ---- 子能力模型(可选自引用)----
	// 语义:当本配置不具备某能力时,委托给对应子模型配置执行(能力路由)。
	// Photo 图像生成/理解的子模型配置(可空)。
	Photo *ModelConfig `gorm:"-" json:"photo,omitempty"`
	// Video 视频相关子模型配置(可空)。
	Video *ModelConfig `gorm:"-" json:"video,omitempty"`
	// Audio 音频相关子模型配置(可空)。
	Audio *ModelConfig `gorm:"-" json:"audio,omitempty"`
}

// TableName 表名(model_configs)。
func (ModelConfig) TableName() string { return "model_configs" }
