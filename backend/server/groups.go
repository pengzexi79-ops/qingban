package server

// P3 群聊端点:群 CRUD、成员管理、手动触发一轮(交棒 AI 包调度引擎)。
// 分工:回合触发(冷却/点名全员/逐成员调用/广播)整体在 AI 包 RunGroupRound(易失态,不落库),
// 本文件只做 HTTP 校验与入参组装。
// 建表语义(v3):创建群 → 同事务落 groups 行 + group_members 行(JoinedAt)+ 一行群会话
// (Conversation{GroupID: &group.ID});解散群 → 删 groups 行,群会话/消息/点名/附件关系
// 经外键级联清理(如需归档请先 /data/export 再删;运行态回合随进程消失)。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。
// v2 注记:Group 出参键 avatarFileId/createdAt 等由视图映射(实体 json 为 file_id);
// strategy.enabled 契约默认 true,而 model.GroupStrategy 零值为 false——落库前显式回落契约缺省,
// 勿依赖 Go 零值(手动"触发一轮"不受 enabled 限制,其仅约束主动起话题任务)。

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
	//     Initial: firstRune(req.Name), Color: req.Color 或 paletteHash(req.Name),
	//     FileID: req.FileID, Announcement: req.Announcement,
	//     Strategy: defaultStrategy(req.Strategy)}                        // 子键零值回落契约缺省
	// tx {
	//     db.Create(&group)
	//     rows := 每组员 = model.GroupMember{GroupID: group.ID, CompanionID: id, JoinedAt: now}
	//     db.CreateInBatches(rows, 50)
	//     db.Create(&model.Conversation{GroupID: &group.ID})              // 群会话行(group_id 唯一)
	// }
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 201, groupWithMembers(group))
	// v2 注记:strategy 必填(契约);子键缺省 random/30/2/member-order、enabled 默认 true
	// (模型零值为 false,见文件头 v2 注记;defaultStrategy 负责回落)。
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
// group_members 由外键级联清理。要归档的群先 /data/export 再删。
func hDeleteGroup(c *gin.Context) {
	// id := parseUintParam(c, "groupId")
	// group := db.First(&model.Group{}, id); if nil { 404 }
	// db.Delete(&group)                                                  // 级联见上方说明
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// GroupMembersReq:POST /groups/:groupId/members 请求体。
// 注:契约对添加接口仅约束 min=1(建群上限 20 只约束创建,见 GroupCreateReq;群规模上限不在此限)。
type GroupMembersReq struct {
	// MemberIds:新增角色(已在群内者忽略)。
	MemberIds []uint `json:"memberIds" binding:"required,min=1"`
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

// hTriggerGroupRound:POST /groups/:groupId/rounds —— 触发一轮(运行期动态回合,201)。
// 语义:轮次不落库(易失,见 core/cache.go key=round:<convID>);本端点只做 HTTP 校验,
// 交棒 AI.RunGroupRound——其内部完成:剔除已删角色(成员<2→错误)、冷却检查
// (COOLDOWN_ACTIVE→409)、写易失现场(round:<convID>)并交棒 Dispatcher.FlushConversation
// (点名全员立即整批投喂);回复过程经装配层 hooks 推送 round_start/round_message/new_message/
// round_end SSE;回合现场回复完成即删(无历史回放接口)。幂等中间件防重放。
func hTriggerGroupRound(c *gin.Context) {
	// id := parseUintParam(c, "groupId")
	// group := db.Preload("Conversation").First(&model.Group{}, id); if nil { 404 }
	// if group.Conversation == nil { respondErr(409, "群会话不存在"); return }
	// result, err := ai.RunGroupRound(c.Request.Context(), ai.GroupRoundArgs{
	//     Group: *group, Members: 群成员(Companion 全量,由调度引擎现读配置),
	//     ConversationID: group.Conversation.ID,          // conversations.id(数字;非 group.id)
	//     Now: time.Now()})
	// // 错误映射:COOLDOWN_ACTIVE→409 / 成员不足→422(VALIDATION_ERROR)/ PROVIDER_ERROR→502
	// if err != nil { respondErr(c, common.ToHTTPStatus(err)...); return }
	// respond(c, 201, result)                            // *ai.RoundResult{roundId(进程内数字), status:"running"}
	// v2 注记:契约 RoundResult.roundId 为数字(运行期进程内递增);消息经 round_message/new_message 事件,
	// 响应内 messages 恒为空(异步)——与旧"同步执行一轮、messages 返回产出"的语义不同。
}
