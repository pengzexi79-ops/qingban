package server

// P1 统一会话列表端点(Thread = 单聊+群聊的统展示对象)。
// 数据来源:conversations 缓存行 + companions/groups 实体联查组装;
// 排序:置顶优先 → last_active_at 倒序(消息落库时刷新缓存行)。
// 伪代码草稿:逻辑以函数体内伪代码注释占位。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// ThreadView:会话列表/摘要响应视图(openapi Thread;详情扩展字段见 ThreadDetail)。
type ThreadView struct {
	// Type:companion/group。
	Type string `json:"type"`
	// ID:单聊=companionId,群聊=groupId(即 conversationId)。
	ID string `json:"id"`
	// Name:角色名或群名。
	Name string `json:"name"`
	// AvatarFileID/Color/Initial:头像展示(从实体带出)。
	AvatarFileID *string `json:"avatarFileId,omitempty"`
	Color        string  `json:"color"`
	Initial      string  `json:"initial"`
	// Unread:未读数。
	Unread int `json:"unread"`
	// Pinned/Muted:置顶/免打扰(缓存行字段)。
	Pinned bool `json:"pinned"`
	Muted  bool `json:"muted"`
	// Online:单聊可用状态(本地占位 true)。
	Online bool `json:"online,omitempty"`
	// MemberCount:群聊成员数。
	MemberCount int `json:"memberCount,omitempty"`
	// LastMessage:最后消息摘要(可空)。
	LastMessage *model.LastMessageBrief `json:"lastMessage,omitempty"`
	// LastTimestamp:最后消息时间(RFC3339,排序展示键)。
	LastTimestamp string `json:"lastTimestamp"`
	// Announcement:群公告(ThreadDetail 扩展)。
	Announcement *string `json:"announcement,omitempty"`
	// MemberIds:群成员 id(ThreadDetail 扩展;单聊=[companionId])。
	MemberIds []string `json:"memberIds,omitempty"`
}

// ThreadListQuery:GET /conversations 查询参数。
type ThreadListQuery struct {
	// Filter:all/unread/companion/group,默认 all。
	Filter string `form:"filter"`
	// Q:匹配会话名称或最后消息摘要(名称走实体 LIKE,摘要走缓存行 LIKE)。
	Q string `form:"q"`
	// Before:游标(按 last_active_at 倒序翻页;id 辅助稳定排序)。
	Before string `form:"before"`
	// Limit:页大小(1~100,默认 20)。
	Limit int `form:"limit"`
}

// hListConversations:GET /conversations —— 会话列表(单聊+群聊聚合,游标分页)。
func hListConversations(c *gin.Context) {
	// var q ThreadListQuery; c.ShouldBindQuery(&q)
	// cur := common.ParseCursor(q.Before, q.Limit)                        // limit 夹取 1~100 默认 20
	// rows := db.Find(Conversation{},                                      // ① 查缓存行
	//     where: kindByFilter(q.Filter) /* all|companion|group|unread>0 */,
	//     and:  (q.Q 匹配 → 实体名 LIKE 子查询 OR last_content LIKE),
	//     order: pinned DESC, last_active_at DESC, id DESC,
	//     limit: cur.Limit+1, offset: byCursor(cur.Before))                // 多取 1 判 hasMore
	// if len(rows) > cur.Limit { rows = rows[:cur.Limit]; hasMore = true }
	// next := hasMore ? last(rows).ID : nil
	// entities := loadCompanionsAndGroups(rows)                           // ② 批量加载实体 → 名称/头像/色/成员数
	// for each row:                                                        // ③ 组装 ThreadView
	//     v := ThreadView{Type, ID, Unread, Pinned, Muted, Online: true(单聊), ...}
	//     if kind==group { v.MemberCount = len(memberIDs); v.Announcement = ... }
	//     if row.LastContent != "" { v.LastMessage = {Content: row.LastContent,
	//                                                   Timestamp: row.LastActiveAt, SenderName: row.LastSenderName} }
	//     v.LastTimestamp = row.LastActiveAt.RFC3339()
	// respond(c, 200, common.Page[ThreadView]{Items: v, NextCursor: next})
}

// hGetConversation:GET /conversations/:conversationId —— 会话摘要(顶栏信息/群公告/成员)。
func hGetConversation(c *gin.Context) {
	// conv := db.Find(Conversation{id}); if nil { 404 }
	// v := threadViewOf(conv)                                              // 复用列表组装内核
	// if conv.Kind == "group" {
	//     group := db.Find(Group{id}); v.MemberIds = memberIDsOf(group)
	//     v.Announcement = group.Announcement
	// } else {
	//     v.MemberIds = []string{conv.ID}                                  // 单聊详情=[companionId]
	// }
	// respond(c, 200, v)
}

// ConversationUpdateReq:PATCH /conversations/:conversationId 请求体(至少一个键)。
type ConversationUpdateReq struct {
	// Pinned:置顶开关。
	Pinned *bool `json:"pinned"`
	// Muted:免打扰开关。
	Muted *bool `json:"muted"`
}

// hPatchConversation:PATCH /conversations/:conversationId —— 置顶/免打扰。
func hPatchConversation(c *gin.Context) {
	// var req ConversationUpdateReq; if !bind(c, &req) { return }          // 空 body → 422(minProperties)
	// conv := db.Find(Conversation{id}); if nil { 404 }
	// if req.Pinned != nil { conv.Pinned = *req.Pinned }                   // 未传键不动
	// if req.Muted != nil { conv.Muted = *req.Muted }
	// db.Save(&conv)
	// respond(c, 200, threadViewOf(conv))
}

// hDeleteConversation:DELETE /conversations/:conversationId —— 删除会话与历史(本地即删)。
func hDeleteConversation(c *gin.Context) {
	// tx {
	//     db.Delete(Conversation{id: param})                                // 删缓存行
	//     db.Delete(Message{conversation_id: param})                        // 删该会话全部消息
	// }
	// // 注意:companion/group 实体保留——下次发消息自动重建缓存行
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// hMarkConversationRead:POST /conversations/:conversationId/read —— 标记已读(人侧)。
func hMarkConversationRead(c *gin.Context) {
	// conv := db.Find(Conversation{id})
	// if conv != nil { conv.Unread = 0; db.Save(&conv) }                  // 会话级红点清零
	// ---- 消息级"人已读"回执(设计 §2.1,可选实现)----
	//  ① 推进人侧消费水位:Upsert(member_cursors{conv, ReaderUser,
	//     LastReadMessageID: 该会话最后一条消息 id})(供"未读拼接"的对称语义,见 AI 包);
	//  ② 批量回填 Message.user_read_at = now(role=assistant 且 user_read_at 为空的行);
	//  ③ 群聊中 AI 成员视角不受影响(它们各自水位独立)。
	// unreadTotal := db.Sum("SELECT COALESCE(SUM(unread),0) FROM conversations")
	// hub.Publish(EventRead, {conversationId: param, unreadTotal})         // 多窗口即时同步
	// respond(c, 200, {unreadTotal})
}

// hSearchThreads:GET /search/threads?q=(必填)—— 会话页顶栏搜索(名称/最后消息命中)。
func hSearchThreads(c *gin.Context) {
	// q := c.Query("q"); if q == "" { respondErr(422, "缺少关键词"); return }
	// v := listThreadsCore(filter: all, q: q, limit: 20)                   // 复用 hListConversations 查询内核
	// respond(c, 200, v)                                                   // []ThreadView(契约不分页)
}
