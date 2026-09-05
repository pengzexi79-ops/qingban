package server

// P2 消息端点:历史、发送、删除、清空 + 跨会话搜索。
// 发送语义(v3 + docs/batch_dispatch_design.md):POST 只负责"用户消息落库 + 广播",
// 立即返回 201;AI 回复由异步批次调度引擎产出(AI 包 Dispatcher,装配见 init/app.go:
// server 提供 DispatchRepo 的 GORM 实现 + TurnRunner + DispatchHooks),经 SSE 推送
// (时序:typing → delta* → new_message,单聊与群聊同一入口)。
// 附件:先 POST /files 再在正文写入引用标记,落库时由 utils.ParseRefs 解析 id 建
// message_files 关系;点名(群聊 @)经 utils.ParseMentions 写 message_mentions + MentionAll 布尔。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。
// v2 注记:Message 出参(contentType/refs/senderId 等)为视图派生,勿直出实体——
// 实体键为 snake_case(conversation_id/sender_companion_id),且实体无 contentType/refs 列。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// MessageCreateReq:POST .../messages 请求体。
// 校验:content 上限 5000 字(含引用标记);去标记纯文本 ≤500、图片 ≤9、附件 ≤20
// 由 utils.ParseRefs 在解析时统一校验(见 utils/refs.go 常量)。
type MessageCreateReq struct {
	// Content:正文(可含 ![图](fileId)/[名](fileId) 引用标记)。
	Content string `json:"content" binding:"max=5000"`
	// MentionIDs:显式点名成员 id(群聊;数字 companions.id,与后端 @ 解析结果并集)。
	MentionIDs []uint `json:"mentionIds,omitempty"`
}

// MessageSendResult:发送响应(v3 语义:201 即返;AI 回复异步经 SSE 到达,assistantMessage 恒空)。
// 结构兼容旧契约(openapi MessageSendResult)保留字段位;memoryCandidates 同样改经
// memory_candidates SSE 事件送达,同步响应体仅作占位。
type MessageSendResult struct {
	// UserMessage:用户消息(已落库)。
	UserMessage model.Message `json:"userMessage"`
	// AssistantMessage:AI 回复(异步引擎产出,本次响应恒为空;经 SSE 到达)。
	AssistantMessage *model.Message `json:"assistantMessage,omitempty"`
	// MemoryCandidates:记忆候选(异步事件承载;保留字段兼容旧契约)。
	MemoryCandidates []model.MemoryDraft `json:"memoryCandidates,omitempty"`
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
	// refIDs, plain, err := utils.ParseRefs(req.Content)            // ① 引用校验(纯文本500/图9/附件20)
	// if err != nil { respondErr(422, err.Error()); return }        //    附:plain 供长度/搜索/摘要统计
	// ② 点名解析:单聊 → all=false,nil;群聊 → all, hits := utils.ParseMentions(plain, 群成员名→id)
	//    并集 req.MentionIDs(数字 id;ids 中不存在的成员过滤);all=@所有人/@all
	// ③ 用户消息落库(role=user,sender_companion_id 为空;content 保留原文引用标记):
	// userMsg := model.Message{ConversationID: conv.ID, Role: "user",
	//     MentionAll: all, Content: req.Content}
	// db.Create(&userMsg)
	// ④ 点名关系行:for _, m := range hits { db.Create(&model.MessageMention{userMsg.ID, m.ID}) }  // 去重
	//    请求体显式 mentionIds 同法(与解析结果并集,均过滤非成员)
	// ⑤ 附件关系:for _, fid := range refIDs { db.Create(&model.MessageFile{MessageID: userMsg.ID,
	//    FileID: uint(fid)}) }                                      // 忽略不存在;重复允许(去重落库)
	// ⑥ 会话行摘要前滚(自己消息未读不加):refreshConversation(conv, &userMsg, unread:false)
	// hub.Publish(EventNewMessage, {conversationId: id, message: userMsg})
	// respond(c, 201, MessageSendResult{UserMessage: userMsg})
	//
	// ---- 调度:交棒异步批次引擎(单聊/群聊统一,见 AI/dispatch.go)----
	// aiDispatcher.NotifyMessage(ctx, userMsg)
	//     // 引擎按"点名即时/闲话攒批"决策投喂(静默窗 core.IdleWindow / 攒批硬顶 / 群冷却);
	//     // 引擎产出经装配层注入的 DispatchHooks 回调:
	//     //   OnReply(convID, msg)    → hub.Publish(new_message)(AI 回复已由引擎经
	//     //                              repo.SaveAssistantMessage 落库并刷新会话摘要/未读)
	//     //   OnRoundStart(convID, readerIDs) → hub.Publish(round_start)
	//     //   OnSilent/OnConsumed     → 已读不回/水位推进(仅审计日志,无前端事件)
	//     // 记忆候选:引擎产出 ChatOutcome.Candidates → ai.PersistCandidates 按角色记忆模式
	//     //          确认/入库 → hub.Publish(memory_candidates, {conversationId, candidates})
	//     // 装配处(init)构造 ai.NewDispatcher(gormRepo, turnRunner, hooks),非包级单例。
}

// hDeleteMessage:DELETE .../messages/:messageId —— 删除单条消息(级联清理点名/附件关系)。
func hDeleteMessage(c *gin.Context) {
	// cid := parseUintParam(c, "conversationId"); mid := parseUintParam(c, "messageId")
	// msg := db.Where("id = ? AND conversation_id = ?", mid, cid).First(&model.Message{}); if nil { 404 }
	// tx {
	//     db.Delete(&msg)                                           // message_mentions/message_files 级联
	//     delFTSMessageRow(mid)                                     // 消息 FTS 行非外键,手动清(若有)
	// }
	// refreshConversationSummary(conv)                              // 摘要回退倒数第二条(取 id 次大)
	// // 被删除消息若被其他消息回复引用,其 ReplyToID 由外键 SET NULL;AI 消费水位不回退
	// // (已读即归档语义下,删历史消息不影响后续投喂)
	// hub.Publish(EventThreadsChanged, {})                          // 会话列表摘要变更(无专用删除事件)
	// respond(c, 204)
}

// hClearMessages:DELETE .../messages?confirm=true —— 清空本会话历史。
func hClearMessages(c *gin.Context) {
	// id := parseUintParam(c, "conversationId")
	// if c.Query("confirm") != "true" { respondErr(400, "需 confirm=true"); return }
	// conv := db.First(&model.Conversation{}, id); if nil { 404 }
	// tx {
	//     db.Where("conversation_id = ?", id).Delete(&model.Message{})   // 级联点名/附件关系
	//     db.Where("conversation_id = ?", id).Delete(&model.MemberCursor{}) // 人/AI 水位一并复位
	//     delFTSMessagesOf(id)                                          // 消息 FTS 行非外键,手动清(若有)
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
// refreshConversation(conv, msg, unread bool)        // 会话行摘要/LastActiveAt/LastMessageID 前滚(自己的消息 unread=false)
// refreshConversationSummary(conv)                    // 删除/清空后摘要回退(取 id 次大消息)
// gormRepo:实现 AI.DispatchRepo 全部方法(会话/成员/未读批/水位/短期记忆/AI 消息落库)
//           —— 落库 SaveAssistantMessage 时补 msg.ID 并刷新会话行,再交由引擎触发 OnReply 广播
// turnRunner:ai.TurnRunner——一次成员发言(装配期接 ai.RunTurn 或 ai.ChatWithCompanion 流水线)
// hooks:ai.DispatchHooks{OnRoundStart/OnReply/OnSilent/OnConsumed/LogError}
//        —— OnReply 内 hub.Publish(new_message)+usage 关联回填;记忆候选见 hSendMessage 注释
// buildChatArgs(companionID, userMsg) ai.ChatArgs    // 组装:角色配置/用户画像/最近轮历史(供 TurnRunner)
