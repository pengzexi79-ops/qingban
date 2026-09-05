package server

// P2 消息端点:历史、发送(SSE 流式/同步)、删除、清空 + 跨会话搜索。
// 发送是阶段一最核心链路(PHASE1 §5.3 验收:浏览器+Ollama 完成
// "发消息→AI 回复→记忆候选"闭环);handler 只做协议编排,模型流水线进 AI 包。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// MessageCreateReq:POST .../messages 请求体。
type MessageCreateReq struct {
	// Content:正文(≤500 用户文本;可含引用标记;纯媒体消息 contentType 区分,content 可为空)。
	Content *string `json:"content" binding:"omitempty,max=5000"`
	// ContentType:text/image/file/voice/video/mixed,默认 text。
	ContentType string `json:"contentType" binding:"omitempty,oneof=text image file voice video mixed"`
}

// MessageSendResult:同步模式响应(openapi MessageSendResult)。
type MessageSendResult struct {
	// UserMessage:用户消息(已落库)。
	UserMessage model.Message `json:"userMessage"`
	// AssistantMessage:AI 回复(成功或 fallback;模型失败时 fallback=true)。
	AssistantMessage model.Message `json:"assistantMessage"`
	// MemoryCandidates:待确认记忆候选(空数组=无)。
	MemoryCandidates []model.MemoryDraft `json:"memoryCandidates"`
}

// hListMessages:GET /conversations/:conversationId/messages —— 历史(时间升序,向前翻页)。
func hListMessages(c *gin.Context) {
	// if !convExists(param) { respondErr(404, "会话不存在"); return }
	// cur := parseCursor(c)                                            // before+limit(夹取)
	// rows := db.Find(Message{},                                       // 向下取更旧
	//     where: conversation_id=param AND (cur.Before=="" OR id < cur.Before),
	//     order: timestamp DESC, id DESC, limit: cur.Limit+1)
	// hasMore := len(rows) > cur.Limit; if hasMore { rows = rows[:cur.Limit] }
	// slices.Reverse(rows)                                             // 升序输出(契约)
	// next := hasMore ? rows[0].ID : nil                               // nextCursor=本页最旧一条
	// respond(c, 200, Page[Message]{items: rows, nextCursor: next})
}

// hSendMessage:POST /conversations/:conversationId/messages —— 发送(核心链路,SSE 或同步)。
func hSendMessage(c *gin.Context) {
	// ---- 入参校验与用户消息落库 ----
	// var req MessageCreateReq; if !bind(c, &req) { return }
	// if req.ContentType == "" { req.ContentType = ContentText }
	// text := ""; if req.Content != nil { text = *req.Content }
	// refs, plain, err := utils.ParseRefs(text); if err != nil { 422 }
	// for r := range refs { assertFileOwned(r.FileID, ScopeMessage) }  // fileId 存在且 scope=message
	// conv := db.Find(Conversation{id}); if nil { 404 }
	// userMsg := Message{ID: message-, ConversationType: conv.Kind, ConversationId: conv.ID,
	//                    Role: user, SenderId: currentUserID, Content: text,
	//                    ContentType: req.ContentType, Refs: refs, Timestamp: now}
	// db.Insert(&userMsg)
	// refreshThreadCache(conv, last: &userMsg)                         // 缓存行 last_active/content/sender 刷新;自己的消息未读不加
	// hub.Publish(EventNewMessage, {conversationId, message: userMsg}) // 多窗口即时可见
	//
	// ---- 单聊 AI 回复(conv.Kind==companion 且角色存在) ----
	// if conv.Kind == "companion" && companionOK(conv.ID) {
	//     args := buildChatArgs(conv.ID, userMsg, plain)                // 读角色配置/用户画像/最近 N 轮历史
	//     stream := isStreamRequest(c)
	//     if stream {
	//         // 广播 typing 后走 AI 流水线,回调把 delta 转 SSE:
	//         //   SSE 序列:typing → delta* → message(完整助手消息) → memory_candidates → done
	//         // 落库 AI 消息(streamed=true)后广播 new_message(双窗口同步)
	//         streamChatReply(c, args, userMsg)
	//         return
	//     }
	//     // 同步:阻塞调用 → 落库 → 组装 MessageSendResult
	//     result := syncChatReply(args, userMsg)
	//     respond(c, 201, result)                                       // 契约 201
	//     return
	// }
	// // 群聊类型:本端点只收用户消息(发言经 /groups/rounds),无 AI 回复
	// respond(c, 201, MessageSendResult{UserMessage: userMsg})
}

// hDeleteMessage:DELETE .../messages/:messageId —— 删除单条消息。
func hDeleteMessage(c *gin.Context) {
	// msg := db.Find(Message{id: mid, conversation_id: cid}); if nil { 404 }
	// tx { db.Delete(&msg) }                                          // 解除该消息对文件的引用(不物理删)
	// if conv.LastMessageID == msg.ID {                                // 删了最后一条 → 摘要回退倒数第二条
	//     prev := lastMessageBefore(cid, msg.Timestamp)
	//     refreshThreadCache(conv, last: prev)                         // prev==nil → 清空摘要
	// }
	// hub.Publish("message_deleted", {conversationId, messageId})      // 多窗口同步删除
	// respond(c, 204)
}

// hClearMessages:DELETE .../messages?confirm=true —— 清空本会话历史。
func hClearMessages(c *gin.Context) {
	// if c.Query("confirm") != "true" { respondErr(400, "需 confirm=true"); return }
	// tx { db.Delete(Message{conversation_id: cid}); clearThreadSummary(conv) }  // 引用解除;保留会话本体
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// MessageHitView:搜索命中项(openapi MessageHit,含回跳定位)。
type MessageHitView struct {
	// MessageID:消息 id(回跳定位)。
	MessageID string `json:"messageId"`
	// ConversationType/ConversationId:归属会话。
	ConversationType string `json:"conversationType"`
	ConversationId   string `json:"conversationId"`
	// ConversationName:会话名回显。
	ConversationName string `json:"conversationName"`
	// SenderName:发送者名(可空)。
	SenderName *string `json:"senderName,omitempty"`
	// Content:命中片段(周边截断)。
	Content string `json:"content"`
	// Timestamp:消息时间。
	Timestamp string `json:"timestamp"`
}

// hSearchMessages:GET /search/messages —— 跨会话消息搜索。
func hSearchMessages(c *gin.Context) {
	// q := c.Query("q"); if q == "" { respondErr(422, "缺少关键词"); return }
	// plain := utils.StripRefs(q)                                      // 忽略引用标记只匹配纯文本
	// rows := db.Search("fts_messages", plain,                          // FTS5;中文可先退化 LIKE
	//     filter: conversation_id=c.Query("conversationId")?, 分页同列表)
	// for r := range rows {                                            // 联查会话名/发送者名
	//     hits = append(hits, MessageHitView{..., content: snippet(r, ±40字)})
	// }
	// respond(c, 200, Page[MessageHitView]{})
}

// ---- 内部链路内核(实现时随测试逐个落地) ----
// buildChatArgs(companionID, userMsg, plain) ai.ChatArgs // 组装:角色配置/用户画像/最近 contextTurns 轮历史
// streamChatReply(c, args, userMsg)                      // SSE:typing→delta*→message→memory_candidates→done
// syncChatReply(args, userMsg) MessageSendResult         // 同步:ai.ChatWithCompanion→落库→候选 PersistCandidates
//   两分支共用收尾:AI 消息落库(usageId 回填)→ usage_records 落行 → 缓存行/广播更新;
//   模型失败 → 落 fallback=true 本地兜底文案(不返回 502,前端不断链)。
