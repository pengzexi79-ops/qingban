package server

// P3 API/模型配置端点:CRUD、连通测试、模型目录(密钥本机加密,浏览器不直连)。
// 基准实体:model.ModelConfig(一个 API 配置 ≡ 一个模型配置,能力开关+子模型自引用);
// ApiProfile 旧契约行仅作迁移兼容(见 model/apiprofile.go Deprecated 注记)。
// 密钥纪律:apiKey 只在 POST/PATCH 请求体出现一次 → utils.SecretBox 加密落库;
// 读取/列表/导出永不回明文(响应层以 SecretConfigured 布尔替代)。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
	// 实现时按需恢复:"qingban/utils"(SecretBox 加密)等
)

// hListModelConfigs:GET /api-profiles —— 模型配置列表(脱敏:不返回 APIKey)。
// 逻辑:
//
//	① rows := db.Find(ModelConfig{})
//	② 逐行脱敏视图:APIKey 置空,补 SecretConfigured = (原 APIKey != "")
//	③ 200 []脱敏视图(后续响应结构统一用 ModelConfigView,见文件尾)
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
	// BaseURI/APIType:连接配置(BaseURI 必填)。
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

// hCreateModelConfig:POST /api-profiles —— 新建(201)。
// 逻辑:
//
//	① bind;APIType 缺省 "openai"
//	② APIKey 非空 → SecretBox.Encrypt 落 APIKey 字段(密文);空密钥(本地 Ollama)允许
//	③ 首个配置 → kv 记默认配置 id(角色无绑定时回落)
//	④ 落库;201 + 脱敏视图
func hCreateModelConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hPatchModelConfig:PATCH /api-profiles/:profileId —— 修改(200;未传键保持)。
// 逻辑:指针字段覆盖;APIKey 指针非空才重加密;落库后返回脱敏视图。
func hPatchModelConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hDeleteModelConfig:DELETE /api-profiles/:profileId —— 删除(204;唯一配置受保护)。
// 逻辑:Count==1 → 409"至少保留一套";删除后解除角色绑定(companion 回退默认);204。
func hDeleteModelConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hTestModelConfig:POST /api-profiles/:profileId/test —— 连通测试(本地后端代理)。
// 逻辑:解密 key → 按 APIType 走协议探测(openai:GET /models 或 1-token chat)
// → {status, latencyMs, detail};成功置 enabled 语义(供后续自动路由)。
func hTestModelConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hListModelConfigModels:GET /api-profiles/:profileId/models —— 拉取该端点可用模型目录。
// 逻辑:GET {BaseURI}/models(Ollama:/api/tags)→ 映射 ModelInfo 列表(能力按名推断)。
func hListModelConfigModels(c *gin.Context) {
	// TODO(实现):见函数注释
}

// ModelConfigView:脱敏响应视图(替代实体直接出参;禁止 APIKey 出网)。
type ModelConfigView struct {
	model.ModelConfig
	// SecretConfigured:是否已配置密钥(原 APIKey 非空)。
	SecretConfigured bool `json:"secretConfigured"`
}

// toView:实体 → 脱敏视图(APIKey 置空)。
func toView(mc model.ModelConfig) ModelConfigView {
	// has := mc.APIKey != ""
	// mc.APIKey = ""
	// return ModelConfigView{ModelConfig: mc, SecretConfigured: has}
	return ModelConfigView{}
}

// 注:原 ApiProfile 系 handler(hListApiProfiles/hCreateApiProfile/hPatchApiProfile/
// hDeleteApiProfile/hTestApiProfile/hListApiProfileModels)已被上述 ModelConfig 系取代,
// 路由映射名将在 router.go 同步更新(见 router.go 中 /api-profiles 段)。
