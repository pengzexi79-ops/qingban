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
	// MentionIDs:显式点名成员 id 列表(群聊;前端已 @ 解析的场景可传此覆盖/并集,
	// 否则后端按 utils.ParseMentions 从 content 解析)。
	MentionIDs []string `json:"mentionIds,omitempty"`
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

// hSendMessage:POST /conversations/:conversationId/messages —— 发送(批次调度语义)。
// 流程(与 docs/batch_dispatch_design.md 对齐):
//
//	① 入参校验、@点名解析(utils.ParseMentions,显式 mentionIds 可覆盖),用户消息落库
//	   (mentions 随行存储),刷新会话缓存行并广播 new_message —— 立即返回 201;
//	② AI 的回复不再在本请求内同步产出:落库后经调度引擎(ai.Dispatcher)异步"攒批/点名
//	   投喂",产出消息由 hooks 再广播 new_message(SSE 时序不变,前端无感);
//	③ 单聊与群聊走同一入口:单聊无点名语义(全量攒批),群聊 @ 谁由引擎决定立即调度谁。
func hSendMessage(c *gin.Context) {
	// ---- 入参校验与用户消息落库 ----
	// var req MessageCreateReq; if !bind(c, &req) { return }
	// if req.ContentType == "" { req.ContentType = ContentText }
	// text := ""; if req.Content != nil { text = *req.Content }
	// refs, plain, err := utils.ParseRefs(text); if err != nil { 422 }   // 附件引用标记
	// conv := db.Find(Conversation{id}); if nil { 404 }
	// mentions := resolveMentions(conv, plain, req.MentionIDs)           // ① 群聊点名解析:
	//     //   单聊 → nil;群聊 → utils.ParseMentions(plain, 成员名→id 映射)
	//     //   解析结果与请求体显式 mentionIds 并集;非法 id 过滤
	// userMsg := Message{ConversationID: conv.ID, Role: RoleUser, Mentions: mentions,
	//                    Content: text, ContentType: req.ContentType, Timestamp: now}
	// db.Insert(&userMsg)
	// refreshThreadCache(conv, last: &userMsg)                          // 自己的消息未读不加
	// hub.Publish(EventNewMessage, {conversationId, message: userMsg})
	// respond(c, 201, MessageSendResult{UserMessage: userMsg})          // 立即返回
	//
	// ---- 调度:交棒给异步批次引擎(单聊/群聊统一) ----
	// ai.Dispatch.NotifyMessage(ctx, userMsg)
	//     // 引擎按设计 §3/§4 攒批或点名投喂;AI 回复落库后 hooks.OnReply →
	//     // hub.Publish(new_message)(assistant 消息,流式与否由角色 ChatStyle 决定,
	//     // 时序仍为 typing → delta* → message,但为异步到达)
	//     // AI 允许静默(silent):不落库不广播(诊断走 OnSilent)
	//     // 群聊 AI 成员的回复对其他成员是"新消息",自然进入他们的未读池
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
