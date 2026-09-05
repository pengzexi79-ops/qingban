package server

// 路由注册总表(P0~P3 全量端点,对应 openapi.phase1.yaml paths)。
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
	protect.GET("/groups/:groupId/rounds", hListGroupRounds) // 阶段一空列表占位
	round := api.Group("/groups/:groupId/rounds", MiddlewareTrace(), MiddlewareLocalToken(), MiddlewareIdempotency())
	round.POST("", hTriggerGroupRound) // 同步一轮(幂等)

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

	// ---- P3 API/模型配置(ModelConfig 基准;原 ApiProfile 系已取代) ----
	protect.GET("/api-profiles", hListModelConfigs)
	protect.POST("/api-profiles", hCreateModelConfig)
	protect.PATCH("/api-profiles/:profileId", hPatchModelConfig)
	protect.DELETE("/api-profiles/:profileId", hDeleteModelConfig)
	protect.POST("/api-profiles/:profileId/test", hTestModelConfig)
	protect.GET("/api-profiles/:profileId/models", hListModelConfigModels)

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
