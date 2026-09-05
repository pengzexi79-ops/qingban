package server

// P1 统一会话列表端点(Thread = 单聊+群聊的统一展示对象)。
// 数据来源(v3):conversations 会话行为主,单聊经 companion_id、群聊经 group_id
// 联回 companions/groups 实体取名称/头像/成员;排序:置顶优先 → last_active_at 倒序。
// 消息/未读/置顶/静音状态都只在会话行上,不再有"主键即业务 id"的旧缓存行。
// 伪代码草稿:逻辑以函数体内伪代码注释占位。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// ThreadView:会话列表/摘要响应视图(openapi Thread;id 一律为 conversations.id 数字直出)。
type ThreadView struct {
	// ID:会话行 id(conversations.id;消息页/已读等路径参数同源)。
	ID uint `json:"id"`
	// Type:companion/group(由会话行归属推断,非存列)。
	Type string `json:"type"`
	// CompanionID/GroupID:会话归属的目标实体(单聊/群二选一,可空)。
	CompanionID *uint `json:"companionId,omitempty"`
	GroupID     *uint `json:"groupId,omitempty"`
	// Name:角色名或群名。
	Name string `json:"name"`
	// FileID/Color/Initial:头像展示(实体带出,数字文件 id)。
	FileID  *uint  `json:"fileId,omitempty"`
	Color   string `json:"color"`
	Initial string `json:"initial"`
	// Unread:未读数。
	Unread int `json:"unread"`
	// Pinned/Muted:置顶/免打扰(会话行字段)。
	Pinned bool `json:"pinned"`
	Muted  bool `json:"muted"`
	// Online:单聊可用状态(本地占位 true)。
	Online bool `json:"online,omitempty"`
	// MemberCount:群聊成员数。
	MemberCount int `json:"memberCount,omitempty"`
	// MemberIds:群成员 id(群详情扩展;单聊不含)。
	MemberIds []uint `json:"memberIds,omitempty"`
	// LastMessage:最后消息摘要(可空)。
	LastMessage *model.LastMessageBrief `json:"lastMessage,omitempty"`
	// LastTimestamp:最后消息时间(展示键,与 last_message 同时刻,RFC3339 文本)。
	LastTimestamp string `json:"lastTimestamp"`
	// Announcement:群公告(ThreadDetail 扩展)。
	Announcement *string `json:"announcement,omitempty"`
}

// threadEntity:列表组装的实体侧摘要(由 companions/groups 批量加载后填充)。
type threadEntity struct {
	// Name:角色名或群名。
	Name string
	// FileID/Color/Initial:头像展示。
	FileID  *uint
	Color   string
	Initial string
	// MemberCount/MemberIDs:群聊成员数/成员 id(单聊为空)。
	MemberCount int
	MemberIDs   []uint
}

// ThreadListQuery:GET /conversations 查询参数。
type ThreadListQuery struct {
	// Filter:all/unread/companion/group,默认 all。
	Filter string `form:"filter"`
	// Q:匹配会话名称或最后消息摘要(名称走实体 LIKE,摘要走会话行 LIKE)。
	Q string `form:"q"`
	// Before:游标(按 last_active_at 倒序翻页;id 辅助稳定排序;值=上页末行 id)。
	Before uint `form:"before"`
	// Limit:页大小(1~100,默认 20)。
	Limit int `form:"limit"`
}

// hListConversations:GET /conversations —— 会话列表(单聊+群聊聚合,游标分页)。
func hListConversations(c *gin.Context) {
	// var q ThreadListQuery; c.ShouldBindQuery(&q)
	// limit := clamp(q.Limit, 1, 100); if q.Limit == 0 { limit = 20 }
	// tx := db.Model(&model.Conversation{})
	// switch q.Filter {                                                   // ① kind 由归属列表达
	// case "companion": tx = tx.Where("companion_id IS NOT NULL")
	// case "group":     tx = tx.Where("group_id IS NOT NULL")
	// case "unread":    tx = tx.Where("unread > 0")
	// }
	// if q.Q != "" { tx = tx.Where("last_content LIKE ? OR id IN (实体名匹配的会话子查询)", ...) }
	// rows := tx.Order("pinned DESC, last_active_at DESC, id DESC").
	//            Where("last_active_at < (?) OR (last_active_at = (?) AND id < ?)", 游标条件...).
	//            Limit(limit + 1).Find(&[]model.Conversation{})            // 多取 1 判 hasMore
	// entities := loadCompanionsAndGroups(rows)                           // ② 批量加载实体(名称/头像/成员数)
	// for _, row := range rows {                                          // ③ 组装 ThreadView
	//     v := threadViewOf(row, entities[row.id])
	//     out = append(out, v)
	// }
	// respond(c, 200, Page[ThreadView]{Items: out, NextCursor: hasMore ? last(rows).ID : 0})
}

// threadViewOf:会话行 + 实体 → ThreadView 组装内核(列表/详情/搜索共用)。
// 说明:Type 由 CompanionID/GroupID 归属推断;LastMessage/LastTimestamp 直接取会话行摘要。
func threadViewOf(row *model.Conversation, entity *threadEntity) *ThreadView {
	// v := &ThreadView{ID: row.ID, Unread: row.Unread, Pinned: row.Pinned, Muted: row.Muted,
	//     LastTimestamp: timeJSON(row.LastActiveAt)}
	// if row.CompanionID != nil { v.Type, v.CompanionID, v.Online = "companion", row.CompanionID, true }
	// if row.GroupID != nil { v.Type, v.GroupID = "group", row.GroupID }
	// if entity != nil { v.Name = entity.Name; v.FileID = entity.FileID; v.Color = entity.Color; v.Initial = entity.Initial }
	// if v.Type == "group" && entity != nil { v.MemberCount = entity.MemberCount; v.MemberIds = entity.MemberIDs }
	// if row.LastContent != "" { v.LastMessage = &model.LastMessageBrief{Content: row.LastContent,
	//     Timestamp: row.LastActiveAt, SenderName: row.LastSenderName} }
	// return v
	return nil // TODO(实现):见函数注释
}

// hGetConversation:GET /conversations/:conversationId —— 会话摘要(顶栏信息/群公告/成员)。
func hGetConversation(c *gin.Context) {
	// id := parseUintParam(c, "conversationId")
	// conv := db.First(&model.Conversation{}, id); if nil { 404 }
	// v := threadViewOf(conv, entityOf(conv))                              // 复用组装内核
	// if v.Type == "group" { group := db.First(&model.Group{}, *conv.GroupID); v.Announcement = group.Announcement }
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
	// id := parseUintParam(c, "conversationId")
	// conv := db.First(&model.Conversation{}, id); if nil { 404 }
	// if req.Pinned != nil { conv.Pinned = *req.Pinned }                   // 未传键不动
	// if req.Muted != nil { conv.Muted = *req.Muted }
	// db.Save(&conv)
	// respond(c, 200, threadViewOf(conv, entityOf(conv)))
}

// hDeleteConversation:DELETE /conversations/:conversationId —— 删除会话及其消息(204)。
// 语义:删 conversations 行即级联删该会话全部消息/点名/附件关系/水位/短期记忆;
// companion/group 实体保留。会话行删除后若再次发消息,服务层需按归属(companion/group)
// 重建一行空会话(等同"清空记录后重新开始")。
func hDeleteConversation(c *gin.Context) {
	// id := parseUintParam(c, "conversationId")
	// conv := db.First(&model.Conversation{}, id); if nil { 404 }
	// tx { db.Delete(&conv) }                                             // 级联见上方说明
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// hMarkConversationRead:POST /conversations/:conversationId/read —— 标记已读(人侧)。
func hMarkConversationRead(c *gin.Context) {
	// id := parseUintParam(c, "conversationId")
	// conv := db.First(&model.Conversation{}, id)
	// if conv != nil {
	//     conv.Unread = 0; db.Save(&conv)                                  // ① 会话级红点清零
	// }
	// ---- 消息级"人已读"回执(设计 §2.1)----
	// last := 该会话最后一条消息 id
	//  ② 推进人侧水位:Upsert(model.MemberCursor{ConversationID: id, ReaderID: model.ReaderUserID,
	//     LastReadMessageID: last})
	//  ③ 批量回填 user_read_at = now:role='assistant' 且 user_read_at IS NULL 的行
	//  ④ 群聊中 AI 成员视角不受影响(它们各自水位独立)
	// unreadTotal := db.Model(&model.Conversation{}).Select("COALESCE(SUM(unread),0)").Scan(&n)
	// hub.Publish(EventRead, {conversationId: id, unreadTotal})            // 多窗口即时同步
	// respond(c, 200, {unreadTotal})
}

// hSearchThreads:GET /search/threads?q=(必填)—— 会话页顶栏搜索(名称/最后消息命中)。
func hSearchThreads(c *gin.Context) {
	// q := c.Query("q"); if q == "" { respondErr(422, "缺少关键词"); return }
	// v := listThreadsCore(filter: all, q: q, limit: 20)                   // 复用 hListConversations 查询内核
	// respond(c, 200, v)                                                   // []ThreadView(契约不分页)
}
