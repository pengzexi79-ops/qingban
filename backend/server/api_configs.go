package server

// P3 API 配置端点:CRUD、连通测试、模型目录(端点:/api-configs;旧 /api-profiles、/model-configs 均已废弃)。
// 基准实体:model.APIConfig(表 api_configs)——"接口通道":连接信息 + 请求参数 + 能力开关 +
// ParentID 子能力路由(子行是普通行,以 parent 指向主行;"反哺"编排在智能体层)。
// 与 AI 用户解耦但可关联:companions.api_config_id 引用本表(每个角色各自绑定、独自引用一个 API,
// 配置删除 → 角色绑定 SET NULL);设置页维护本表,角色设置页做关联。
// 密钥纪律:apiKey 只在 POST/PATCH 请求体出现一次 → utils.SecretBox 加密落库;
// 读取/列表/导出永不回明文(响应层以 SecretConfigured 布尔替代)。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。
// v2 注记(openapi v2 重写要点):
//   - 端点路径 /api-configs、路径参数 apiConfigId;旧契约 /api-profiles/{profileId} 与其
//     ApiProfile 字段(provider/region/protocol/baseUrl/chatModel…/maskedKey)整段废弃;
//   - APIConfig 视图字段为 name/display_name/base_uri/api_type/temperature/max_tokens/
//     text_completion 等蛇形键(实体 json 直出);如 v2 走驼峰需视图映射:
//     displayName/baseUri/maxTokens/photoGeneration/…;
//   - parentId(子能力路由)与"角色可绑定一个 API"为旧契约没有的新语义;
//   - 首个配置自动记为默认配置(kv k:api_configs:default_id),角色未绑定回落;

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
	// 实现时按需恢复:"qingban/utils"(SecretBox 加密)等
)

// hListAPIConfigs:GET /api-configs —— API 配置列表(脱敏:不返回 APIKey)。
// 逻辑:
//
//	① rows := db.Find(&[]model.APIConfig{})
//	② 逐行脱敏视图:APIKey 置空,补 SecretConfigured = (原 APIKey != "")
//	③ 200 []脱敏视图(APIConfigView)
func hListAPIConfigs(c *gin.Context) {
	// respond(c, 200, rows)
}

// APIConfigCreateReq:创建请求体(与实体同构,但 APIKey 明文仅此出现;能力开关缺省 false)。
type APIConfigCreateReq struct {
	// Name 唯一 API 名(必填)。
	Name string `json:"name" binding:"required,max=100"`
	// DisplayName/Description/Version:展示信息。
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Version     string `json:"version"`
	// ParentID:父配置(子能力配置行指向主配置行,可空)。
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

// hCreateAPIConfig:POST /api-configs —— 新建(201)。
// 逻辑:
//
//	① bind;APIType 缺省 "openai";ParentID 非空时校验父配置存在(可空)
//	② APIKey 非空 → SecretBox.Encrypt 落 APIKey 字段(密文);空密钥(本地 Ollama)允许
//	③ 首个配置 → kv 记默认配置 id(model.KVDefaultAPIConfigID;角色未绑定配置时回落)
//	④ 落库;201 + 脱敏视图
func hCreateAPIConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hPatchAPIConfig:PATCH /api-configs/:apiConfigId —— 修改(200;未传键保持)。
// 逻辑:指针字段覆盖(含 ParentID);APIKey 指针非空才重加密;落库后返回脱敏视图。
func hPatchAPIConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hDeleteAPIConfig:DELETE /api-configs/:apiConfigId —— 删除(204;唯一配置受保护)。
// 逻辑:Count==1 → 409"至少保留一套";删除前把绑定的 companions.api_config_id 置空
// (子行 ParentID 由外键 SET NULL 自动脱离,变独立配置);204。
func hDeleteAPIConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hTestAPIConfig:POST /api-configs/:apiConfigId/test —— 连通测试(本地后端代理)。
// 逻辑:解密 key → ai.NewClient(配置行, 解密key).Test()(真实签名返回
// status/latencyMs/detail,映射契约 {status, latencyMs, detail});失败 → 502 PROVIDER_ERROR。
func hTestAPIConfig(c *gin.Context) {
	// TODO(实现):见函数注释
}

// hListAPIConfigModels:GET /api-configs/:apiConfigId/models —— 拉取该端点可用模型目录。
// 逻辑:ai.NewClient(配置行, 解密key).ListModels() → ai.ModelInfo 列表(能力按名推断)。
func hListAPIConfigModels(c *gin.Context) {
	// TODO(实现):见函数注释
}

// APIConfigView:脱敏响应视图(替代实体直接出参;禁止 APIKey 出网)。
type APIConfigView struct {
	model.APIConfig
	// SecretConfigured:是否已配置密钥(原 APIKey 非空)。
	SecretConfigured bool `json:"secretConfigured"`
}

// toAPIConfigView:实体 → 脱敏视图(APIKey 置空)。
func toAPIConfigView(mc model.APIConfig) APIConfigView {
	// has := mc.APIKey != ""
	// mc.APIKey = ""
	// return APIConfigView{APIConfig: mc, SecretConfigured: has}
	return APIConfigView{}
}
