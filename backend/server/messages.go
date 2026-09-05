package server

// P2 消息端点:历史、发送、删除、清空 + 跨会话搜索。
// 发送语义(v3 + docs/batch_dispatch_design.md):POST 只负责"用户消息落库 + 广播",
// 立即返回 201;AI 回复由异步批次调度引擎产出,经 DispatchHooks 落库并发 SSE(时序:
// typing → delta* → message),单聊与群聊同一入口。附件走"先 /files 上传再在正文引用",
// 落库时由引用标记建立 message_files 关系;点名写 message_mentions 关系 + MentionAll 布尔。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// MessageCreateReq:POST .../messages 请求体。
type MessageCreateReq struct {
	// Content:正文(可含 ![图](fileId)/[名](fileId) 引用标记)。
	Content string `json:"content" binding:"max=5000"`
	// MentionIDs:显式点名成员 id(群聊;前端已解析时可传,与后端 @ 解析结果并集)。
	MentionIDs []uint `json:"mentionIds,omitempty"`
}

// MessageSendResult:发送响应(openapi MessageSendResult)。
type MessageSendResult struct {
	// UserMessage:用户消息(已落库)。
	UserMessage model.Message `json:"userMessage"`
	// AssistantMessage:AI 回复(异步引擎产出,本次响应恒为空;经 SSE 到达)。
	AssistantMessage *model.Message `json:"assistantMessage,omitempty"`
}

// hListMessages:GET /conversations/:conversationId/messages —— 历史(升序,向前翻页)。
func hListMessages(c *gin.Context) {
	// id := parseUintParam(c, "conversationId")
	// if !convExists(id) { respondErr(404, "会话不存在"); return }
	// cur := parseCursor(c)                                            // before(id)+limit(夹取)
	// rows := db.Preload("Files").Where("conversation_id = ? AND id < ?", id, cur.Before).
	//     Order("id DESC").Limit(cur.Limit + 1).Find(&[]model.Message{}) // 时间序 = id 序(created_at 自增)
	// hasMore := len(rows) > cur.Limit; if hasMore { rows = rows[:cur.Limit] }
	// slices.Reverse(rows)                                             // 升序输出(契约)
	// next := hasMore ? rows[0].ID : 0                                 // nextCursor=本页最旧一条 id
	// respond(c, 200, Page[model.Message]{Items: rows, NextCursor: next})
}

// hSendMessage:POST /conversations/:conversationId/messages —— 发送(201 即返,异步回复)。
func hSendMessage(c *gin.Context) {
	// id := parseUintParam(c, "conversationId")
	// var req MessageCreateReq; if !bind(c, &req) { return }
	// conv := db.Preload("Messages").First(&model.Conversation{}, id); if nil { 404 }
	// plain := req.Content
	// ① 点名解析:单聊 → all=false,nil;群聊 → allFlag, hits := utils.ParseMentions(plain, 群成员名→id)
	//    并集 req.MentionIDs;ids 中不存在的成员过滤
	// ② 用户消息落库(role=user,sender_companion_id 为空):
	// userMsg := model.Message{ConversationID: conv.ID, Role: "user",
	//     MentionAll: allFlag, Content: plain}
	// db.Create(&userMsg)
	// ③ 点名关系行:for _, cid := range ids { db.Create(&model.MessageMention{userMsg.ID, cid}) }
	// ④ 附件关系:refs := utils.ParseRefsIds(plain); for fid := range refs {
	//    db.Create(&model.MessageFile{MessageID: userMsg.ID, FileID: fid}) }  // 忽略不存在/重复
	// ⑤ 会话行摘要前滚(自己消息未读不加):refreshConversation(conv, &userMsg, unread:false)
	// hub.Publish(EventNewMessage, {conversationId: id, message: userMsg})
	// respond(c, 201, MessageSendResult{UserMessage: userMsg})
	//
	// ---- 调度:交棒异步批次引擎(单聊/群聊统一) ----
	// ai.Dispatch.NotifyMessage(ctx, userMsg)
	//     // 引擎按设计攒批/点名投喂;AI 回复经 server 装配的 DispatchHooks:
	//     //   OnReply → 落库 assistant 消息(message_files/usage 回填)→ refreshConversation(+1)
	//     //             → hub.Publish(new_message/typing/delta);记忆候选事件由引擎另发
	//     //   OnSilent → 已读不回(仅推进水位,不落库不广播)
	//     // 群聊成员回复对其他成员是"新消息",自然进入各自未读池
}

// hDeleteMessage:DELETE .../messages/:messageId —— 删除单条消息(级联清理点名/附件关系)。
func hDeleteMessage(c *gin.Context) {
	// cid := parseUintParam(c, "conversationId"); mid := parseUintParam(c, "messageId")
	// msg := db.Where("id = ? AND conversation_id = ?", mid, cid).First(&model.Message{}); if nil { 404 }
	// tx { db.Delete(&msg) }                                           // message_mentions/message_files 级联
	// refreshConversationSummary(conv)                                  // 摘要回退倒数第二条(取 id 次大)
	// hub.Publish("message_deleted", {conversationId: cid, messageId: mid})
	// respond(c, 204)
}

// hClearMessages:DELETE .../messages?confirm=true —— 清空本会话历史。
func hClearMessages(c *gin.Context) {
	// id := parseUintParam(c, "conversationId")
	// if c.Query("confirm") != "true" { respondErr(400, "需 confirm=true"); return }
	// conv := db.First(&model.Conversation{}, id); if nil { 404 }
	// tx {
	//     db.Where("conversation_id = ?", id).Delete(&model.Message{})   // 级联点名/附件关系
	//     conv.LastMessageID = nil; conv.LastContent = ""; conv.LastSenderName = nil; conv.Unread = 0
	//     db.Save(&conv)
	// }
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// MessageHitView:搜索命中项(openapi MessageHit,含回跳定位)。
type MessageHitView struct {
	// MessageID:消息 id(回跳定位)。
	MessageID uint `json:"messageId"`
	// ConversationID/ConversationType:归属会话(id 数字;类型由会话行归属推断)。
	ConversationID   uint   `json:"conversationId"`
	ConversationType string `json:"conversationType"`
	// ConversationName:会话名回显。
	ConversationName string `json:"conversationName"`
	// SenderName:发送者名(可空)。
	SenderName *string `json:"senderName,omitempty"`
	// Content:命中片段(周边截断)。
	Content string `json:"content"`
	// Timestamp:消息时间(created_at,RFC3339 文本)。
	Timestamp string `json:"timestamp"`
}

// hSearchMessages:GET /search/messages —— 跨会话消息搜索。
func hSearchMessages(c *gin.Context) {
	// q := c.Query("q"); if q == "" { respondErr(422, "缺少关键词"); return }
	// rows := db.Search("fts_messages", q,                                // FTS5;中文可先退化 LIKE
	//     filter: conversation_id=c.Query("conversationId")?(uint), 分页同列表)
	// for r := range rows {                                            // 联查会话名/发送者名
	//     hits = append(hits, MessageHitView{MessageID: r.ID,
	//         ConversationID: r.ConversationID, ConversationType: typeOfConv(r.ConversationID),
	//         Content: snippet(r.Content, ±40字), ...})
	// }
	// respond(c, 200, Page[MessageHitView]{})
}

// ---- 内部链路(实现时随测试逐个落地) ----
// refreshConversation(conv, msg, unread bool)        // 会话行摘要/LastActiveAt/LastMessageID 前滚
// refreshConversationSummary(conv)                    // 删除/清空后摘要回退(取 id 次大消息)
// DispatchHooks(OnUserMsg/OnReply/OnSilent/OnCandidate)// 引擎事件 → 落库/会话维护/SSE 广播(server 装配)
// buildChatArgs(companionID, userMsg) ai.ChatArgs    // 组装:角色配置/用户画像/最近轮历史(供 TurnRunner)
