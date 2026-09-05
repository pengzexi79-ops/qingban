package server

// P3 群聊端点:群 CRUD、成员管理、手动触发一轮(第一阶段"同步执行一轮")。
// 分工:轮次调度(冷却/选人/逐成员调用/落库/广播)整体在 AI 包 RunGroupRound,
// 本文件只做 HTTP 校验与入参组装。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// GroupView:群响应视图(openapi Group;memberIds+memberSummaries 由 service 组装)。
type GroupView struct {
	model.Group
	// MemberSummaries:成员摘要(姓名/头像缩略,列表可选带出)。
	MemberSummaries []map[string]any `json:"memberSummaries,omitempty"`
}

// GroupCreateReq:POST /groups 请求体(契约:成员 ≥2 ≤20)。
type GroupCreateReq struct {
	// Name:群名(≤20,必填)。
	Name string `json:"name" binding:"required,max=20"`
	// AvatarFileID:群头像(可选)。
	AvatarFileID *string `json:"avatarFileId"`
	// Color/Initial:文字头像(可选)。
	Color   *string `json:"color"`
	Initial *string `json:"initial"`
	// Announcement:群公告(≤500)。
	Announcement *string `json:"announcement" binding:"omitempty,max=500"`
	// MemberIds:从通讯录选择的角色(≥2,必填)。
	MemberIds []string `json:"memberIds" binding:"required,min=2,max=20"`
	// Strategy:调度策略(必填;子键缺省 random/30/2/member-order)。
	Strategy model.GroupStrategy `json:"strategy" binding:"required"`
}

// hListGroups:GET /groups —— 群列表(成员摘要+策略;契约不分页)。
func hListGroups(c *gin.Context) {
	// groups := db.Find(Group{}, order: created_at DESC)
	// memberMap := batchLoadMembers(groups)                             // group_members+companions 联查
	// for g := range groups {                                            // 组装 memberIds + memberSummaries
	//     g.MemberIds = memberMap[g.ID].ids
	//     out  = append(out, GroupView{g, memberMap[g.ID].summaries})
	// }
	// respond(c, 200, out)
}

// hCreateGroup:POST /groups —— 建群(校验成员齐全,201)。
func hCreateGroup(c *gin.Context) {
	// var req GroupCreateReq; if !bind(c, &req) { return }
	// for id := range req.MemberIds { if !companionExists(id) { respondErr(422, "成员已删除: "+id); return } }
	// group := model.Group{ID: "group-"+uuid4(), Name: req.Name,                    // 默认补齐
	//     Initial: firstRune(req.Name), Color: paletteHash(req.Name),
	//     Announcement: req.Announcement, Strategy: defaultStrategy(req.Strategy)}
	// tx {
	//     db.Insert(&group)
	//     db.BatchInsert(groupMembers(group.ID, req.MemberIds, joinedAt: now))
	//     db.Insert(&Conversation{ID: group.ID, Kind: "group", LastActiveAt: now})  // 会话列表即时可见
	// }
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 201, groupWithMembers(group))
}

// hGetGroup:GET /groups/:groupId —— 群详情(含成员)。
func hGetGroup(c *gin.Context) {
	// group := db.Find(Group{id}); if nil { 404 }
	// respond(c, 200, groupWithMembers(group))
}

// GroupUpdateReq:PATCH /groups/:groupId 请求体(改名/头像/公告/策略)。
type GroupUpdateReq struct {
	Name         *string              `json:"name" binding:"omitempty,max=20"`
	AvatarFileID *string              `json:"avatarFileId"`
	Color        *string              `json:"color"`
	Initial      *string              `json:"initial"`
	Announcement *string              `json:"announcement" binding:"omitempty,max=500"`
	Strategy     *model.GroupStrategy `json:"strategy"`
}

// hPatchGroup:PATCH /groups/:groupId —— 分段更新。
func hPatchGroup(c *gin.Context) {
	// group := db.Find(Group{id}); if nil { 404 }
	// var req GroupUpdateReq; if !bind(c, &req) { return }
	// if req.Name != nil { group.Name = *req.Name }                     // 指针字段覆盖
	// if req.Strategy != nil { group.Strategy = *req.Strategy }         // 策略整体替换
	// ...同法覆盖 avatarFileId/color/initial/announcement
	// db.Save(&group)
	// hub.Publish(EventSettingsChanged, {scope: "group", id: group.ID}) // 名称实时联查,列表无需刷缓存行
	// respond(c, 200, groupWithMembers(group))
}

// hDeleteGroup:DELETE /groups/:groupId —— 解散群(消息保留=归档可导出)。
func hDeleteGroup(c *gin.Context) {
	// tx {
	//     db.Delete(Group{id}); db.Delete(GroupMember{group_id: id})
	//     db.Delete(Round{group_id: id}); db.Delete(Conversation{id})   // 会话缓存同删
	// }
	// // messages 保留(conversation_id=groupId,可经 /data/export 导出 → "归档可导出"语义)
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// GroupMembersReq:POST /groups/:groupId/members 请求体。
type GroupMembersReq struct {
	// MemberIds:新增角色(已在群内者忽略)。
	MemberIds []string `json:"memberIds" binding:"required,min=1"`
}

// hAddGroupMembers:POST /groups/:groupId/members —— 添加成员(已在内者忽略)。
func hAddGroupMembers(c *gin.Context) {
	// group := db.Find(Group{id}); if nil { 404 }
	// var req GroupMembersReq; if !bind(c, &req) { return }
	// existing := memberIDsOf(group)
	// for id := range req.MemberIds { if companionExists(id) && !existing[id] { 插入 group_members } }
	// respond(c, 200, groupWithMembers(group))
}

// hRemoveGroupMember:DELETE /groups/:groupId/members/:companionId —— 移出成员。
func hRemoveGroupMember(c *gin.Context) {
	// group := db.Find(Group{id}); if nil { 404 }
	// if len(memberIDsOf(group)) <= 2 { respondErr(422, "群成员少于 2 人,请解散群"); return }  // 契约保护
	// db.Delete(GroupMember{group_id, companion_id: param})
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// hListGroupRounds:GET /groups/:groupId/rounds —— 轮次记录。
// 说明(契约):第一阶段返回空列表占位;实现阶段按 triggered_at 倒序分页查 rounds 表
// (异步轮次明细第二阶段补)。404=群不存在。
func hListGroupRounds(c *gin.Context) {
	// if !groupExists(param) { respondErr(404, "群不存在"); return }
	// cur := parseCursor(c)
	// rows := db.Find(Round{group_id: param}, order: triggered_at DESC, 分页)
	// respond(c, 200, Page[Round]{items: rows, nextCursor: ...})
}

// hTriggerGroupRound:POST /groups/:groupId/rounds —— 触发一轮(同步执行,201)。
func hTriggerGroupRound(c *gin.Context) {
	// group := db.Find(Group{id}); if nil { 404 }
	// members := loadEnabledMembers(group)                              // 剔除已删角色
	// if len(members) < 2 { respondErr(422, "可用成员不足 2 人"); return }
	// userPrompt := bindJSON {prompt?}(可空;空=由调度选人起话题)
	// result := ai.RunGroupRound(ctx, GroupRoundArgs{group, members, userPrompt, convID: group.ID, now})  // 幂等中间件已挡重放
	// // 错误映射:COOLDOWN_ACTIVE→409 / PROACTIVE_DISABLED→409 / PROVIDER_ERROR→502
	// respond(c, 201, result)                                           // completed/cancelled(契约 201)
}
