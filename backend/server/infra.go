package server

// P0 基础设施端点:/health、/bootstrap、/bootstrap/init、/events(SSE)、/refresh。
// 阶段验收线(PHASE1 §5.1):"进程能起、前端能连、能收事件"由本批端点承载。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"
)

// apiVersion:对外暴露的 API 版本(与 openapi.phase1.yaml version 对齐)。
const apiVersion = "0.2.0-phase1"

// Health:GET /health 响应体(契约 schema)。
type Health struct {
	// Status:ok/degraded(DB 不可达时 degraded)。
	Status string `json:"status"`
	// APIVersion:契约版本。
	APIVersion string `json:"apiVersion"`
	// DbOK:数据库探活结果。
	DbOK bool `json:"dbOk"`
	// ServerTime:本机后端进程当前时间(RFC3339 UTC)。
	ServerTime string `json:"serverTime"`
}

// hHealth:GET /health —— 健康检查。
func hHealth(c *gin.Context) {
	// dbOK := pingDB() == nil                        // SQLite 探活(SELECT 1)
	// status := "ok"; if !dbOK { status = "degraded" }
	// respond(c, 200, Health{status, apiVersion, dbOK, time.Now().UTC().Format(RFC3339)})
}

// BootstrapCounts:引导态计数摘要(/bootstrap 响应用)。
type BootstrapCounts struct {
	// Companions:角色数。
	Companions int64 `json:"companions"`
	// Messages:消息数。
	Messages int64 `json:"messages"`
	// Memories:记忆数。
	Memories int64 `json:"memories"`
}

// BootstrapResp:GET /bootstrap 响应体。
type BootstrapResp struct {
	// FirstRun:true=显示首次引导;false=直接进主界面。
	FirstRun bool `json:"firstRun"`
	// UserID:当前用户 id(未初始化时省略)。
	UserID string `json:"userId,omitempty"`
	// DataVersion:数据版本(未初始化时省略)。
	DataVersion string `json:"dataVersion,omitempty"`
	// Counts:实体计数(未初始化时省略)。
	Counts *BootstrapCounts `json:"counts,omitempty"`
}

// hGetBootstrap:GET /bootstrap —— 首次引导状态(开放端点)。
func hGetBootstrap(c *gin.Context) {
	// done := kvGet(KVBootstrapDone) == "1"           // kv 表单键读
	// if !done { respond(c, 200, BootstrapResp{FirstRun: true}); return }
	// resp := BootstrapResp{FirstRun: false}
	// resp.UserID = firstUser().ID
	// resp.DataVersion = kvGet(KVDataVersion)          // 缺省 "qingban_api_v1"
	// resp.Counts = &BootstrapCounts{                  // 三表 COUNT
	//     Companions: count(companions), Messages: count(messages), Memories: count(memories)}
	// respond(c, 200, resp)
}

// BootstrapInitReq:POST /bootstrap/init 请求体。
type BootstrapInitReq struct {
	// Mode:empty(空白空间+种子配置)/import-demo(再迁移演示数据)。
	Mode string `json:"mode" binding:"required,oneof=empty import-demo"`
	// ImportPayload:mode=import-demo 时携带演示导出 JSON(前端 localStorage 导出格式)。
	ImportPayload map[string]any `json:"importPayload,omitempty"`
}

// hPostBootstrapInit:POST /bootstrap/init —— 初始化本地数据空间(幂等:二次调用 409)。
func hPostBootstrapInit(c *gin.Context) {
	// var req BootstrapInitReq; if !bind(c, &req) { return }        // 422
	// if kvGet(KVBootstrapDone) == "1" { respondErr(409, CodeConflict, "本地空间已初始化"); return }
	// tx {
	//     db.Insert(&User{ID: "user-"+uuid4(), Nickname: "我", Settings: defaultSettings})   // ① 单行用户
	//     seed := ApiProfile{ID: profile-, provider: "ollama", baseUrl: "http://localhost:11434",   // ② 种子 Ollama 配置
	//                        protocol: ProtoOpenAI, name: "本地模型(Ollama)", region: "本地"}
	//     db.Insert(&seed); kvSet(KVDefaultProfileID, seed.ID)
	//     if req.Mode == "import-demo" {                               // ③ 演示数据迁移(内核见 data.go)
	//         if req.ImportPayload == nil { 422 "缺少导入文件"; rollback }
	//         stats = importPayloadCore(req.ImportPayload)             // 不写迁移 kv
	//     }
	//     kvSet(KVBootstrapDone, "1"); kvSet(KVDataVersion, "qingban_api_v1")
	// }
	// hub.Publish(EventSettingsChanged, {scope: "bootstrap"})          // ④ 已订阅前端刷新
	// respond(c, 200, {userId, dataVersion: "qingban_api_v1", importStats: stats?, defaultApiProfileId})
}

// hSSEEvents:GET /events —— 本地实时事件订阅(SSE 长连接)。
// 注意:不能走 gin 压缩/缓冲,需禁用 ResponseWriter 缓冲(见注释尾)。
func hSSEEvents(c *gin.Context) {
	// c.Header("Content-Type", "text/event-stream")
	// c.Header("Cache-Control", "no-cache"); c.Header("X-Accel-Buffering", "no")
	// lastID := parseLastEventID(c.GetHeader("Last-Event-ID"))        // -1=全新连接
	// sub := core.Hub.Subscribe(lastID)                                // ① 注册订阅者(含补发历史)
	// defer closeSub(sub)                                              // ② 断开时清理
	// sendSSE(c, "snapshot", {unreadTotal: sumUnread(), presence: onlineCompanionIDs()})  // ③ 首帧快照
	// for {                                                           // ④ 事件泵
	//     select {
	//     case frame := <-sub.Events(): c.Writer.Write(frame); c.Writer.Flush()
	//     case <-c.Request.Context().Done(): return                   // 客户端断开
	//     case <-sub.Done(): return                                   // Hub 关闭
	//     }
	// }
}

// hRefresh:GET /refresh?include=threads,presence —— 会话状态重建(下拉刷新)。
func hRefresh(c *gin.Context) {
	// includes := split(c.Query("include"))                           // 逗号分隔,默认 threads
	// unreadTotal := sumUnread()                                       // conversations SUM(unread)
	// resp := {unreadTotal}
	// if contains(includes, "threads") { resp.threads = listThreadsCore(filter: all, q: "", noPaging: true) }
	// if contains(includes, "presence") { resp.presence = onlineCompanionIDs() }  // 本地=全部启用角色;第二阶段接可用性
	// respond(c, 200, resp)
}
