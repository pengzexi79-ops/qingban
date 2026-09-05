package server

// P3 群聊端点:群 CRUD、成员管理、手动触发一轮(第一阶段"同步执行一轮")。
// 分工:轮次调度(冷却/选人/逐成员调用/落库/广播)整体在 AI 包 RunGroupRound,
// 本文件只做 HTTP 校验与入参组装。
// 建表语义(v3):创建群 → 同事务落 groups 行 + group_members 行(JoinedAt)+ 一行群会话
// (Conversation{GroupID: &group.ID});解散群 → 删 groups 行,会话/消息/点名/轮次/发言
// 经外键级联清理(如需归档请先 /data/export 再删)。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// GroupView:群响应视图(openapi Group;memberIds+memberSummaries 由 service 组装)。
type GroupView struct {
	model.Group
	// MemberIds:群成员 id(数字直出,取自 group_members)。
	MemberIds []uint `json:"memberIds,omitempty"`
	// MemberSummaries:成员摘要(姓名/头像缩略,列表可选带出)。
	MemberSummaries []map[string]any `json:"memberSummaries,omitempty"`
}

// GroupCreateReq:POST /groups 请求体(契约:成员 ≥2 ≤20)。
type GroupCreateReq struct {
	// Name:群名(≤40,必填)。
	Name string `json:"name" binding:"required,max=40"`
	// FileID:群头像(files.id,可选)。
	FileID *uint `json:"fileId"`
	// Color/Initial:文字头像(可选)。
	Color   *string `json:"color"`
	Initial *string `json:"initial"`
	// Announcement:群公告(≤500)。
	Announcement *string `json:"announcement" binding:"omitempty,max=500"`
	// MemberIds:从通讯录选择的角色(≥2,必填)。
	MemberIds []uint `json:"memberIds" binding:"required,min=2,max=20"`
	// Strategy:调度策略(必填;子键缺省 random/30/2/member-order)。
	Strategy model.GroupStrategy `json:"strategy" binding:"required"`
}

// hListGroups:GET /groups —— 群列表(成员摘要+策略;契约不分页)。
func hListGroups(c *gin.Context) {
	// groups := db.Preload("Conversation").Order("created_at DESC").Find(&[]model.Group{})
	// memberMap := batchLoadMembers(groups)                               // group_members+companions 联查
	// for g := range groups {                                             // 组装 memberIds + memberSummaries
	//     out = append(out, GroupView{*g, memberMap[g.ID].ids, memberMap[g.ID].summaries})
	// }
	// respond(c, 200, out)
}

// hCreateGroup:POST /groups —— 建群(校验成员齐全,201)。
func hCreateGroup(c *gin.Context) {
	// var req GroupCreateReq; if !bind(c, &req) { return }
	// for _, id := range req.MemberIds { if !companionExists(id) { respondErr(422, "成员不存在"); return } }
	// group := model.Group{Name: req.Name,                               // 默认补齐;ID 自增不赋
	//     Initial: firstRune(req.Name), Color: paletteHash(req.Name),
	//     FileID: req.FileID, Announcement: req.Announcement, Strategy: defaultStrategy(req.Strategy)}
	// tx {
	//     db.Create(&group)
	//     rows := 每组员 = model.GroupMember{GroupID: group.ID, CompanionID: id, JoinedAt: now}
	//     db.CreateInBatches(rows, 50)
	//     db.Create(&model.Conversation{GroupID: &group.ID})              // 群会话行(group_id 唯一)
	// }
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 201, groupWithMembers(group))
}

// hGetGroup:GET /groups/:groupId —— 群详情(含成员)。
func hGetGroup(c *gin.Context) {
	// id := parseUintParam(c, "groupId")
	// group := db.Preload("Conversation").First(&model.Group{}, id); if nil { 404 }
	// respond(c, 200, groupWithMembers(*group))
}

// GroupUpdateReq:PATCH /groups/:groupId 请求体(改名/头像/公告/策略)。
type GroupUpdateReq struct {
	Name         *string              `json:"name" binding:"omitempty,max=40"`
	FileID       *uint                `json:"fileId"`
	Color        *string              `json:"color"`
	Initial      *string              `json:"initial"`
	Announcement *string              `json:"announcement" binding:"omitempty,max=500"`
	Strategy     *model.GroupStrategy `json:"strategy"`
}

// hPatchGroup:PATCH /groups/:groupId —— 分段更新。
func hPatchGroup(c *gin.Context) {
	// id := parseUintParam(c, "groupId")
	// group := db.First(&model.Group{}, id); if nil { 404 }
	// var req GroupUpdateReq; if !bind(c, &req) { return }
	// if req.Name != nil { group.Name = *req.Name }                      // 指针字段覆盖
	// if req.Strategy != nil { group.Strategy = *req.Strategy }          // 策略整体替换
	// ...同法覆盖 fileId/color/initial/announcement(改头像前校验文件存在)
	// db.Save(&group)                                                    // UpdatedAt 自动维护
	// hub.Publish(EventSettingsChanged, {scope: "group", id: group.ID})
	// respond(c, 200, groupWithMembers(*group))
}

// hDeleteGroup:DELETE /groups/:groupId —— 解散群(204)。
// 级联(单事务):groups 行删除 → conversations(群会话)→ messages/点名/附件关系/
// rounds→round_speakers/group_members 由外键级联清理。要归档的群先 /data/export 再删。
func hDeleteGroup(c *gin.Context) {
	// id := parseUintParam(c, "groupId")
	// group := db.First(&model.Group{}, id); if nil { 404 }
	// db.Delete(&group)                                                  // 级联见上方说明
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// GroupMembersReq:POST /groups/:groupId/members 请求体。
type GroupMembersReq struct {
	// MemberIds:新增角色(已在群内者忽略)。
	MemberIds []uint `json:"memberIds" binding:"required,min=1,max=20"`
}

// hAddGroupMembers:POST /groups/:groupId/members —— 添加成员(已在内者忽略)。
func hAddGroupMembers(c *gin.Context) {
	// id := parseUintParam(c, "groupId")
	// group := db.First(&model.Group{}, id); if nil { 404 }
	// var req GroupMembersReq; if !bind(c, &req) { return }
	// existing := memberIDSetOf(group)
	// rows := for id := range req.MemberIds { if companionExists(id) && !existing[id] {
	//     model.GroupMember{GroupID: group.ID, CompanionID: id, JoinedAt: now} } }
	// db.CreateInBatches(rows, 50)
	// respond(c, 200, groupWithMembers(*group))
}

// hRemoveGroupMember:DELETE /groups/:groupId/members/:companionId —— 移出成员。
func hRemoveGroupMember(c *gin.Context) {
	// id := parseUintParam(c, "groupId"); mid := parseUintParam(c, "companionId")
	// group := db.First(&model.Group{}, id); if nil { 404 }
	// if len(memberIDsOf(group)) <= 2 { respondErr(422, "群成员少于 2 人,请解散群"); return } // 契约保护
	// db.Delete(&model.GroupMember{GroupID: group.ID, CompanionID: mid})
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// hListGroupRounds:GET /groups/:groupId/rounds —— 轮次记录(rounds 行,触发时刻倒序分页)。
func hListGroupRounds(c *gin.Context) {
	// id := parseUintParam(c, "groupId")
	// if !groupExists(id) { respondErr(404, "群不存在"); return }
	// rows := db.Where("group_id = ?", id).Order("triggered_at DESC").Find(&[]model.Round{})  // 分页
	// respond(c, 200, Page[model.Round]{items: rows, nextCursor: ...})
}

// hTriggerGroupRound:POST /groups/:groupId/rounds —— 触发一轮(同步执行,201)。
func hTriggerGroupRound(c *gin.Context) {
	// id := parseUintParam(c, "groupId")
	// group := db.Preload("Conversation").First(&model.Group{}, id); if nil { 404 }
	// members := loadEnabledMembers(group)                               // 剔除已删角色
	// if len(members) < 2 { respondErr(422, "可用成员不足 2 人"); return }
	// if group.Conversation == nil { respondErr(409, "群会话不存在"); return }
	// userPrompt := bindJSON {prompt?}(可空;空=由调度选人起话题)
	// result := ai.RunGroupRound(ctx, GroupRoundArgs{group, members, userPrompt,
	//     ConversationID: group.Conversation.ID, now})                   // 会话 id = conversations.id(幂等已挡重放)
	// // 错误映射:COOLDOWN_ACTIVE→409 / PROACTIVE_DISABLED→409 / PROVIDER_ERROR→502
	// respond(c, 201, result)                                            // completed/cancelled(契约 201)
}
