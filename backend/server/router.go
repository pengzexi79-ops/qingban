package server

// 路由注册总表(P0~P3 全量端点)。路由集合自 openapi.phase1.yaml 冻结后已按
// Model v3(数字自增主键/会话行 id/api_configs)演进,与旧契约的差异总清单
// (openapi v2 重写时以本注记 + 各文件"v2 注记"为准,详见 docs/model_design.md):
//   ① id 一律数字自增(文档仍写 uuid4 字符串;companionId/conversationId 等路径参数同);
//   ② API 配置端点 /api-configs(文档旧路径 /api-profiles、字段 provider/baseUrl/chatModel 等
//      旧模型已废弃,现为 APIConfig 实体:name/base_uri/api_type/能力开关/parent_id);
//   ③ 不提供 GET /groups/:groupId/rounds——轮次为运行期易失态,无历史回放接口;
//   ④ 会话以 conversations.id 定位(旧契约 conversationId=companionId/groupId 归属语义已废,
//      现归属由会话行 companion_id/group_id 列表达);
//   ⑤ /me、/bootstrap 的 userId 为数字主键;分页游标 nextCursor 为数字 id(文档为字符串)。
// 设计约定:
//   - 全部业务端点挂在 /api/v1 分组(Base URL 见契约);
//   - 中间件分层:开放端点(health/bootstrap/events/refresh)只挂 trace;
//     其余挂 token;幂等中间件按发送类端点逐个挂;
//   - /events(SSE)保持长连接,不受普通响应缓冲影响(handler 内处理);
//   - 501 占位路由已注册(前端可安全调用),返回 NOT_IMPLEMENTED;
//   - 本文件只做"路径→handler/中间件"的静态映射,不含业务逻辑。

import "github.com/gin-gonic/gin"

// RegisterRoutes:注册全部路由(initapp.NewApp 第 8 步调用)。
func RegisterRoutes(g *gin.Engine) {
	// 分组:open=免令牌;protect=令牌保护;idem 类端点另挂幂等中间件
	api := g.Group("/api/v1")

	open := api.Group("")
	open.Use(MiddlewareTrace())

	protect := api.Group("")
	protect.Use(MiddlewareTrace(), MiddlewareLocalToken())

	// ---- P0 基础设施(开放) ----
	open.GET("/health", hHealth)
	open.GET("/bootstrap", hGetBootstrap)
	open.POST("/bootstrap/init", hPostBootstrapInit)
	open.GET("/events", hSSEEvents) // SSE 长连接(不挂令牌:浏览器调试;正式壳注入同源令牌后改为 protect)
	open.GET("/refresh", hRefresh)

	// ---- P1 我的资料 ----
	protect.GET("/me", hGetMe)
	protect.PATCH("/me", hPatchMe)
	protect.POST("/me/avatar", hPostMyAvatar) // multipart
	protect.GET("/me/stats", hGetMyStats)

	// ---- P1/P3 AI 通讯录 ----
	protect.GET("/companions", hListCompanions)
	protect.POST("/companions", hCreateCompanion)
	protect.GET("/companions/:companionId", hGetCompanion)
	protect.PATCH("/companions/:companionId", hPatchCompanion)
	protect.DELETE("/companions/:companionId", hDeleteCompanion)
	protect.GET("/companions/:companionId/memories", hListCompanionMemories)
	protect.POST("/companions/:companionId/proactive", hTriggerProactive501) // 501 占位(第二阶段)

	// ---- P1 统一会话列表 ----
	protect.GET("/conversations", hListConversations)
	protect.GET("/conversations/:conversationId", hGetConversation)
	protect.PATCH("/conversations/:conversationId", hPatchConversation)
	protect.DELETE("/conversations/:conversationId", hDeleteConversation)
	protect.POST("/conversations/:conversationId/read", hMarkConversationRead)

	// ---- P2 消息 ----
	protect.GET("/conversations/:conversationId/messages", hListMessages)
	send := api.Group("/conversations/:conversationId/messages", MiddlewareTrace(), MiddlewareLocalToken(), MiddlewareIdempotency())
	send.POST("", hSendMessage)                                               // SSE 或同步(幂等)
	protect.DELETE("/conversations/:conversationId/messages", hClearMessages) // ?confirm=true
	protect.DELETE("/conversations/:conversationId/messages/:messageId", hDeleteMessage)

	// ---- P3 群聊 ----
	protect.GET("/groups", hListGroups)
	protect.POST("/groups", hCreateGroup)
	protect.GET("/groups/:groupId", hGetGroup)
	protect.PATCH("/groups/:groupId", hPatchGroup)
	protect.DELETE("/groups/:groupId", hDeleteGroup)
	protect.POST("/groups/:groupId/members", hAddGroupMembers)
	protect.DELETE("/groups/:groupId/members/:companionId", hRemoveGroupMember)
	// 注:不再提供 GET /groups/:groupId/rounds——轮次为运行期易失态,无历史回放接口;
	// 手动"触发一轮"见下方 POST(幂等分组)。
	round := api.Group("/groups/:groupId/rounds", MiddlewareTrace(), MiddlewareLocalToken(), MiddlewareIdempotency())
	round.POST("", hTriggerGroupRound) // 运行期动态回合(幂等)

	// ---- P1/P3 长期记忆 ----
	protect.GET("/memories", hListMemories)
	protect.POST("/memories", hCreateMemory)
	protect.PATCH("/memories/:memoryId", hPatchMemory)
	protect.DELETE("/memories/:memoryId", hDeleteMemory)
	protect.POST("/memories/search", hSearchMemories)
	protect.POST("/memories/:memoryId/reindex", hReindexMemory)

	// ---- 搜索 ----
	protect.GET("/search/companions", hSearchCompanions)
	protect.GET("/search/messages", hSearchMessages)
	protect.GET("/search/threads", hSearchThreads)
	protect.GET("/search/memories", hSearchMemoryTexts)

	// ---- P1 本地文件 ----
	protect.POST("/files", hUploadFile) // multipart
	protect.GET("/files/:fileId", hGetFile)
	protect.DELETE("/files/:fileId", hDeleteFile)
	protect.POST("/files/purge-orphans", hPurgeOrphans)

	// ---- P3 API 配置(APIConfig 基准;旧 /api-profiles、/model-configs 均已废弃) ----
	protect.GET("/api-configs", hListAPIConfigs)
	protect.POST("/api-configs", hCreateAPIConfig)
	protect.PATCH("/api-configs/:apiConfigId", hPatchAPIConfig)
	protect.DELETE("/api-configs/:apiConfigId", hDeleteAPIConfig)
	protect.POST("/api-configs/:apiConfigId/test", hTestAPIConfig)
	protect.GET("/api-configs/:apiConfigId/models", hListAPIConfigModels)

	// ---- P2 用量 ----
	protect.GET("/usage/summary", hGetUsageSummary)
	protect.GET("/usage/records", hListUsageRecords)
	protect.GET("/usage/trend", hGetUsageTrend)

	// ---- P1 数据迁移(import 幂等) ----
	protect.GET("/data/export", hExportData)
	dataImport := api.Group("/data/import", MiddlewareTrace(), MiddlewareLocalToken(), MiddlewareIdempotency())
	dataImport.POST("", hImportData) // 兼容 api_v1 与前端演示格式
	protect.DELETE("/data", hClearData)

	// 兜底:未知路径由 gin 默认 404;panic 由 Recovery 处理(均以 JSON 错误体返回,统一在 middleware 内注记)
}
