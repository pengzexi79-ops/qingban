package server

// P3 模型/API 配置端点:CRUD、连通测试、模型目录(端点:/model-configs;旧 /api-profiles 已废弃)。
// 基准实体:model.ModelConfig —— 一个配置 ≡ 一个模型入口(连接信息 + 请求参数 + 能力开关 +
// ParentID 子能力路由:子配置行是普通行,以 parent 指向主配置行,主行遇不擅长的能力时调度到子行,
// 结果可回传主行续写"反哺";编排在智能体层)。
// 密钥纪律:apiKey 只在 POST/PATCH 请求体出现一次 → utils.SecretBox 加密落库;
// 读取/列表/导出永不回明文(响应层以 SecretConfigured 布尔替代)。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
	// 实现时按需恢复:"qingban/utils"(SecretBox 加密)等
)

// hListModelConfigs:GET /model-configs —— 模型配置列表(脱敏:不返回 APIKey)。
// 逻辑:
//
//	① rows := db.Find(&[]model.ModelConfig{})
//	② 逐行脱敏视图:APIKey 置空,补 SecretConfigured = (原 APIKey != "")
//	③ 200 []脱敏视图(ModelConfigView)
func hListModelConfigs(c *gin.Context) {
	// respond(c, 200, rows)
}

// ModelConfigCreateReq:创建请求体(与实体同构,但 APIKey 明文仅此出现;能力开关缺省 false)。
type ModelConfigCreateReq struct {
	// Name 唯一名(必填)。
	Name string `json:"name" binding:"required,max=100"`
	// DisplayName/Description/Version:展示信息。
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Version     string `json:"version"`
	// ParentID:父配置(子能力配置行指向主配置行,可空;见文件头"反哺"说明)。
	ParentID *uint `json:"parentId"`
	// BaseURI/APIKey/APIType:连接配置(BaseURI 必填;APIKey 密文落库)。
	BaseURI string `json:"baseUri" binding:"required"`
	APIKey  string `json:"apiKey"`
	APIType string `json:"apiType"`
	// 请求参数(缺省取实体默认:0.7/1.0/2048/0/0)。
	Temperature      *float64 `json:"temperature"`
	TopP             *float64 `json:"topP"`
	MaxTokens        *int     `json:"maxTokens"`
	FrequencyPenalty *float64 `json:"frequencyPenalty"`
	PresencePenalty  *float64 `json:"presencePenalty"`
	// 能力开关。
	TextCompletion     bool `json:"textCompletion"`
	PhotoGeneration    bool `json:"photoGeneration"`
	VideoGeneration    bool `json:"videoGeneration"`
	AudioGeneration    bool `json:"audioGeneration"`
	TextToSpeech       bool `json:"textToSpeech"`
	ImageUnderstanding bool `json:"imageUnderstanding"`
	VideoUnderstanding bool `json:"videoUnderstanding"`
	AudioUnderstanding bool `json:"audioUnderstanding"`
}

// hCreateModelConfig:POST /model-configs —— 新建(201)。
// 逻辑:
//
//	① bind;APIType 缺省 "openai";ParentID 非空时校验父配置存在(可空)
//	② APIKey 非空 → SecretBox.Encrypt 落 APIKey 字段(密文);空密钥(本地 Ollama)允许
//	③ 首个配置 → kv 记默认配置 id(model.ConfigKVs KVDefaultModelConfigID;角色无绑定时回落)
//	④ 落库;201 + 脱敏视图
func hCreateModelConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hPatchModelConfig:PATCH /model-configs/:modelConfigId —— 修改(200;未传键保持)。
// 逻辑:指针字段覆盖(含 ParentID);APIKey 指针非空才重加密;落库后返回脱敏视图。
func hPatchModelConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hDeleteModelConfig:DELETE /model-configs/:modelConfigId —— 删除(204;唯一配置受保护)。
// 逻辑:Count==1 → 409"至少保留一套";删除前把绑定的 companions.model_config_id 置空
// (子行 ParentID 由外键 SET NULL 自动脱离,变独立配置);204。
func hDeleteModelConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hTestModelConfig:POST /model-configs/:modelConfigId/test —— 连通测试(本地后端代理)。
// 逻辑:解密 key → 按 APIType 走协议探测(openai:GET /models 或 1-token chat;
// ollama:/api/tags)→ {status, latencyMs, detail};成功语义供后续自动路由。
func hTestModelConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hListModelConfigModels:GET /model-configs/:modelConfigId/models —— 拉取该端点可用模型目录。
// 逻辑:ai.NewClient(配置行, 解密key).ListModels() → ai.ModelInfo 列表(能力按名推断)。
func hListModelConfigModels(c *gin.Context) {
	// TODO(实现):见函数注释
}

// ModelConfigView:脱敏响应视图(替代实体直接出参;禁止 APIKey 出网)。
type ModelConfigView struct {
	model.APIConfig
	// SecretConfigured:是否已配置密钥(原 APIKey 非空)。
	SecretConfigured bool `json:"secretConfigured"`
}

// toView:实体 → 脱敏视图(APIKey 置空)。
func toView(mc model.APIConfig) ModelConfigView {
	// has := mc.APIKey != ""
	// mc.APIKey = ""
	// return ModelConfigView{ModelConfig: mc, SecretConfigured: has}
	return ModelConfigView{}
}
