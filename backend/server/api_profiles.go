package server

// P3 API 配置端点:CRUD、连通测试、模型目录(密钥本机加密,浏览器不直连)。
// 密钥纪律:apiKey 只在 POST/PATCH 请求体出现一次 → SecretBox 加密落库;
// 任何读取响应只带 secretConfigured/maskedKey(见 model/apiprofile.go)。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
	"qingban/utils"
)

// ApiProfileCreateReq:POST /api-profiles 请求体(契约必填:name/provider/baseUrl)。
type ApiProfileCreateReq struct {
	// Name:配置名(≤30,必填)。
	Name string `json:"name" binding:"required,max=30"`
	// Provider:服务商标识(必填:openai/anthropic/ollama/qwen/custom)。
	Provider string `json:"provider" binding:"required"`
	// Region:展示分组(国内/国外/本地/自定义)。
	Region string `json:"region"`
	// Protocol:载荷协议(缺省 openai-compatible)。
	Protocol string `json:"protocol" binding:"omitempty,oneof=openai-compatible anthropic ollama"`
	// BaseUrl:服务端地址(必填)。
	BaseUrl string `json:"baseUrl" binding:"required,url"`
	// ApiKey:明文密钥(可选;仅本次请求,服务端加密保存)。
	ApiKey string `json:"apiKey"`
	// 七类模型(均可选,由 /models 目录下拉填充):
	ChatModel       string `json:"chatModel"`
	VisionModel     string `json:"visionModel"`
	HearingModel    string `json:"hearingModel"`
	TTSModel        string `json:"ttsModel"`
	VoiceCloneModel string `json:"voiceCloneModel"`
	VideoModel      string `json:"videoModel"`
	ImageModel      string `json:"imageModel"`
	// Temperature:采样温度(0~2,缺省 0.8)。
	Temperature *float64 `json:"temperature"`
}

// sanitizeProfile:响应脱敏视图——清空 APIKeyEnc、补 SecretConfigured/MaskedKey。
// 说明:库中只存密文;掩码选择"解密后即时计算再丢弃"(不落掩码列,避免衍生明文)。
func sanitizeProfile(p *model.ApiProfile, box *utils.SecretBox) model.ApiProfile {
	// out := *p; out.APIKeyEnc = ""
	// if p.APIKeyEnc != "" {
	//     out.SecretConfigured = true
	//     if plain, err := box.Decrypt(p.APIKeyEnc); err == nil { out.MaskedKey = &utils.MaskKey(plain) }
	// }
	// return out
	return model.ApiProfile{}
}

// hListApiProfiles:GET /api-profiles —— 配置列表(脱敏)。
func hListApiProfiles(c *gin.Context) {
	// rows := db.Find(ApiProfile{})
	// out := make([]model.ApiProfile, 0, len(rows))
	// for i := range rows { out = append(out, sanitizeProfile(&rows[i], box)) }
	// respond(c, 200, out)
}

// hCreateApiProfile:POST /api-profiles —— 新建(201)。
func hCreateApiProfile(c *gin.Context) {
	// var req ApiProfileCreateReq; if !bind(c, &req) { return }
	// if req.Protocol == "" { req.Protocol = model.ProtoOpenAI }
	// p := model.ApiProfile{ID: "profile-" + uuid4(), ..., Protocol: req.Protocol}
	// if req.ApiKey != "" { p.APIKeyEnc = box.Encrypt(req.ApiKey) }      // ② 加密;空密钥(本地 Ollama)允许
	// if req.Temperature == nil { t := 0.8; p.Temperature = &t }
	// if db.Count(ApiProfile{}) == 0 { kvSet(KVDefaultProfileID, p.ID) } // ③ 首个=默认
	// db.Insert(&p)
	// respond(c, 201, sanitizeProfile(&p, box))
}

// ApiProfileUpdateReq:PATCH /api-profiles/:profileId 请求体(apiKey 缺省=不更新)。
type ApiProfileUpdateReq struct {
	Name     *string `json:"name" binding:"omitempty,max=30"`
	Provider *string `json:"provider"`
	Region   *string `json:"region"`
	Protocol *string `json:"protocol" binding:"omitempty,oneof=openai-compatible anthropic ollama"`
	BaseUrl  *string `json:"baseUrl"`
	// ApiKey:仅非 nil 且非空时更新(缺省/空串均不动原密钥,契约语义保守处理)。
	ApiKey          *string  `json:"apiKey"`
	ChatModel       *string  `json:"chatModel"`
	VisionModel     *string  `json:"visionModel"`
	HearingModel    *string  `json:"hearingModel"`
	TTSModel        *string  `json:"ttsModel"`
	VoiceCloneModel *string  `json:"voiceCloneModel"`
	VideoModel      *string  `json:"videoModel"`
	ImageModel      *string  `json:"imageModel"`
	Temperature     *float64 `json:"temperature"`
}

// hPatchApiProfile:PATCH /api-profiles/:profileId —— 修改(200)。
func hPatchApiProfile(c *gin.Context) {
	// p := db.Find(ApiProfile{id}); if p == nil { respondErr(404, "配置不存在"); return }
	// var req ApiProfileUpdateReq; if !bind(c, &req) { return }
	// if req.Name != nil { p.Name = *req.Name }                          // 指针字段覆盖(未传键不动)
	// ...同法:provider/region/protocol/baseUrl/七模型/temperature
	// if req.ApiKey != nil && *req.ApiKey != "" { p.APIKeyEnc = box.Encrypt(*req.ApiKey) }  // 重加密
	// db.Save(&p)
	// respond(c, 200, sanitizeProfile(p, box))
}

// hDeleteApiProfile:DELETE /api-profiles/:profileId —— 删除(204)。
func hDeleteApiProfile(c *gin.Context) {
	// p := db.Find(ApiProfile{id}); if p == nil { respondErr(404, "配置不存在"); return }
	// if db.Count(ApiProfile{}) <= 1 { respondErr(409, CodeConflict, "至少保留一套配置"); return }  // ①
	// tx {
	//     db.Delete(&p)
	//     db.Update(Companion{}, {api_profile_id: NULL}, where: api_profile_id == p.ID)  // ② 解除角色绑定(回落默认)
	//     if kvGet(KVDefaultProfileID) == p.ID {                          // ③ 默认指针改指任一剩余配置
	//         any := db.First(ApiProfile{}); kvSet(KVDefaultProfileID, any.ID)
	//     }
	// }
	// respond(c, 204)
}

// hTestApiProfile:POST /api-profiles/:profileId/test —— 连通测试。
func hTestApiProfile(c *gin.Context) {
	// p := db.Find(ApiProfile{id}); if p == nil { respondErr(404, "配置不存在"); return }
	// secret := ""; if p.APIKeyEnc != "" { secret, _ = box.Decrypt(p.APIKeyEnc) }   // 解密仅进程内
	// status, latency, detail := ai.NewClient(p, secret).Test()
	// if status == "success" { p.Enabled = true; db.Save(&p) }
	// if status == "failed" { respondErr(502, CodeProviderError, detail); return }
	// respond(c, 200, {status, latencyMs: latency, detail?})              // 失败原因摘要,不泄密钥
}

// hListApiProfileModels:GET /api-profiles/:profileId/models —— 模型目录+能力推断。
func hListApiProfileModels(c *gin.Context) {
	// p := db.Find(ApiProfile{id}); if p == nil { respondErr(404, "配置不存在"); return }
	// secret := ""; if p.APIKeyEnc != "" { secret, _ = box.Decrypt(p.APIKeyEnc) }
	// models, err := ai.NewClient(p, secret).ListModels()
	// if err != nil { respondErr(502, CodeProviderError, "服务商不可达"); return }
	// respond(c, 200, {models})
}
