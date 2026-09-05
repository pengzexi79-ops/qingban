package server

// P1/P3 AI 通讯录端点(角色 CRUD)+ 搜索 + proactive 501 占位。
// 建表语义(v3):创建角色 → 落库后在同一事务里建一行单聊会话
// (Conversation{CompanionID: &companion.ID},companion_id 唯一);
// 删除角色 → 只删 companions 行,级联链(model_design §3.1 删除规则):
//   conversation(单聊会话行)→ messages/message_mentions/message_files/member_cursors/
//   chat_short_memories 随会话级联(见 model/conversation.go、message.go、cursor.go、short_memory.go);
//   memories 直接级联(角色删其记忆随删);group_members 关系行级联;
//   该角色发言过的"群消息"保留正文,SenderCompanionID 由外键 SET NULL;
//   files 仅解除引用(头像文件不物理删,交孤儿清理);
//   记忆向量/记忆 FTS 行非外键管理,需手动按 companion_id 清理(见 hDeleteCompanion)。
//   注:round.go 已不存在——轮次为运行期易失态(model_design §3.3),无落库表。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。
// v2 注记:角色出参键 avatarFileId/avatarFileId/apiProfileId 等由视图映射(实体 json 为
// file_id/api_config_id 蛇形键),勿直接序列化 model.Companion。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// CompanionQuery:GET /companions 查询参数。
type CompanionQuery struct {
	// Q:关键词(匹配 name/category/tagline)。
	Q string `form:"q"`
	// Category:分类精确筛选。
	Category string `form:"category"`
	// HasUnread:仅返回有未读的角色。
	HasUnread bool `form:"hasUnread"`
	// Sort:排序(默认置顶+最近活跃;可选 name/memoryCount,扩展位)。
	Sort string `form:"sort"`
}

// hListCompanions:GET /companions —— 角色列表(含 memoryCount/unread/pinned/online;开放数组不分页)。
func hListCompanions(c *gin.Context) {
	// var q CompanionQuery; c.ShouldBindQuery(&q)
	// tx := db.Model(&model.Companion{}).
	//     Preload("Conversation")                                        // 会话行一次带出(列表排序/红点/置顶依据)
	// if q.Q != "" { tx = tx.Where("name LIKE ? OR category LIKE ? OR tagline LIKE ?", ...) }
	// if q.Category != "" { tx = tx.Where("category = ?", q.Category) }
	// list := tx.Order("created_at DESC").Find(&[]model.Companion{})
	// sort: 置顶(会话行 pinned)优先 → 会话行 last_active_at 倒序(会话行为空会话按行零值垫底)
	// memCount := db.Model(&model.Memory{}).Select("companion_id,COUNT(*)").Where("companion_id IN ?", ids).Group("companion_id")
	// for i := range list {                                               // 视图组装(派生字段不入库):
	//     if conv := list[i].Conversation; conv != nil {
	//         list[i].Unread = conv.Unread; list[i].Pinned = conv.Pinned
	//     }
	//     list[i].MemoryCount = memCount[list[i].ID]; list[i].Online = true // 本地占位(第二阶段接可用性)
	// }
	// if q.HasUnread { list = slices.DeleteFunc(list, func(x) bool { return x.Unread == 0 }) }
	// respond(c, 200, list)                                               // []Companion 不分页
	// v2 注记:unread/pinned 在会话行,companions 表无此列(勿加列);排序无法仅靠角色表完成。
}

// CompanionCreateReq:POST /companions 请求体(快速创建:name 必填,其余默认)。
// 注:id/头像/配置引用均为数字(id 直出;FileID 指向 files.id,指针=未设置)。
type CompanionCreateReq struct {
	// Name:角色名(≤32,必填)。
	Name string `json:"name" binding:"required,max=32"`
	// Initial:文字头像字(缺省取 name 首字符)。
	Initial *string `json:"initial" binding:"omitempty,max=1"`
	// Color:主题色 #RRGGBB(缺省从调色板按 name 哈希取值)。
	Color *string `json:"color"`
	// FileID:头像文件(可选;须存在且类型为图片)。
	FileID *uint `json:"fileId"`
	// Category:分类(缺省"自定义")。
	Category *string `json:"category"`
	// Tagline:一句话介绍(≤40)。
	Tagline *string `json:"tagline" binding:"omitempty,max=40"`
	// APIConfigID:该角色独自引用的 API 配置(缺省=空,运行时回落默认配置)。
	APIConfigID *uint `json:"apiConfigId"`
	// Persona:人设(仅提供 relationship/identity 即可,其余子键默认空)。
	Persona *model.Persona `json:"persona"`
	// MemorySettings/ChatStyle/Dispatch/Proactive/Capabilities:可选覆盖;缺省用后端默认。
	MemorySettings *model.MemorySettings   `json:"memorySettings"`
	ChatStyle      *model.ChatStyle        `json:"chatStyle"`
	Dispatch       *model.DispatchSettings `json:"dispatchSettings"`
	Proactive      *model.ProactiveConfig  `json:"proactive"`
	Capabilities   *model.Capabilities     `json:"capabilities"`
}

// hCreateCompanion:POST /companions —— 创建角色(默认值补齐后落库,201)。
func hCreateCompanion(c *gin.Context) {
	// var req CompanionCreateReq; if !bind(c, &req) { return }              // 422
	// if req.APIConfigID != nil && !apiConfigExists(*req.APIConfigID) { respondErr(422, "API 配置不存在"); return }
	// companion := defaultCompanion(req)                                    // 集中补默认(见下):
	//     // id 自增(gorm.Model,不赋 ID);initial=首字符(req.Initial 缺省时);color=哈希取色板;category="自定义"
	//     // persona.relationship 缺省"朋友";memorySettings=hybrid/rolling/365/0.65/12
	//     // chatStyle=markdown/streaming/typing/splitMessages=true,650,soft,allowSilent:true
	//     // dispatch=零值经 model.DispatchSettings.ApplyDefaults 回落(见 defaultCompanion);
	//     // proactive/capabilities=零值(第二阶段的默认执行按 core/backend 文档)
	// db.Create(&companion)                                                 // ① 角色(自增 id)
	// db.Create(&model.Conversation{CompanionID: &companion.ID})            // ② 单聊会话行(companion_id 唯一)
	// hub.Publish(EventThreadsChanged, {})                                  // 列表整体刷新
	// respond(c, 201, companion)
	// v2 注记:201 响应出参为角色视图(见文件头 v2 注记),含派生 unread/pinned(读新会话行)。
}

// hGetCompanion:GET /companions/:companionId —— 角色详情(编辑表单全量)。
func hGetCompanion(c *gin.Context) {
	// id := parseUintParam(c, "companionId")
	// companion := db.Preload("Conversation").First(&model.Companion{}, id)
	// if companion == nil { respondErr(404, CodeNotFound, "角色不存在"); return }
	// companion.MemoryCount = countMemoriesOf(companion.ID)                // 派生字段
	// if conv := companion.Conversation; conv != nil { companion.Unread = conv.Unread; companion.Pinned = conv.Pinned }
	// respond(c, 200, companion)
}

// CompanionUpdateReq:PATCH /companions/:companionId 请求体(分段提交,未传键保持)。
// 说明:字段清单同 CompanionCreateReq(全部指针化、去掉 required,子键缺省不校验);
// 嵌套对象(persona/memorySettings/chatStyle/dispatch/proactive/capabilities)前端整对象提交→整体替换。
type CompanionUpdateReq struct {
	// Name:角色名(≤32)。
	Name *string `json:"name" binding:"omitempty,max=32"`
	// Initial:文字头像字(≤1)。
	Initial *string `json:"initial" binding:"omitempty,max=1"`
	// Color:主题色 #RRGGBB。
	Color *string `json:"color"`
	// FileID:头像文件(可空=解除头像;指向 files.id 须存在)。
	FileID *uint `json:"fileId"`
	// Category:分类。
	Category *string `json:"category"`
	// Tagline:一句话介绍(≤40)。
	Tagline *string `json:"tagline" binding:"omitempty,max=40"`
	// APIConfigID:绑定的 API 配置(可空=解除绑定,回落默认配置;指向须存在)。
	APIConfigID *uint `json:"apiConfigId"`
	// Persona/MemorySettings/ChatStyle/Dispatch/Proactive/Capabilities:
	// 嵌套对象整提交→JSON 列整体替换;未传键保持。
	Persona        *model.Persona          `json:"persona"`
	MemorySettings *model.MemorySettings   `json:"memorySettings"`
	ChatStyle      *model.ChatStyle        `json:"chatStyle"`
	Dispatch       *model.DispatchSettings `json:"dispatchSettings"`
	Proactive      *model.ProactiveConfig  `json:"proactive"`
	Capabilities   *model.Capabilities     `json:"capabilities"`
}

// hPatchCompanion:PATCH /companions/:companionId —— 分段更新角色配置。
func hPatchCompanion(c *gin.Context) {
	// id := parseUintParam(c, "companionId")
	// companion := db.First(&model.Companion{}, id); if nil { 404 }
	// var req CompanionUpdateReq; if !bind(c, &req) { return }
	// if req.Name != nil { companion.Name = *req.Name }                    // 顶层标量指针覆盖
	// if req.Persona != nil { companion.Persona = *req.Persona }           // 嵌套子对象整体替换(JSON 列重写)
	// if req.MemorySettings != nil { companion.MemorySettings = *req.MemorySettings }
	// if req.ChatStyle != nil { companion.ChatStyle = *req.ChatStyle }
	// if req.Dispatch != nil { companion.Dispatch = *req.Dispatch }
	// if req.Proactive != nil { companion.Proactive = *req.Proactive }
	// if req.Capabilities != nil { companion.Capabilities = *req.Capabilities }
	// if req.APIConfigID != nil { if !apiConfigExists(*req.APIConfigID) { 422 }; companion.APIConfigID = req.APIConfigID }
	// if req.FileID != nil { if !fileExists(*req.FileID) { 422 }; companion.FileID = req.FileID }
	// // 注:FileID 为 *uint(可空=未设置);用 nil 指解除需专门表示(契约未定义"解除头像",
	// //     暂以传空串/忽略处理,v2 再定);AgentManager.Evict(companion.ID) 使旧 runtime 失效。
	// db.Save(&companion)                                                   // UpdatedAt 由 gorm.Model 自动维护
	// hub.Publish(EventSettingsChanged, {scope: "companion", id: companion.ID})
	// respond(c, 200, companion)
}

// hDeleteCompanion:DELETE /companions/:companionId —— 删除角色(204)。
// 级联(单事务):删 companions 行后,conversation(单聊会话)→messages→(message_mentions/
// message_files)、memories、group_members、member_cursors、chat_short_memories 均由外键级联;
// 非外键部分手动清理:
//   - 记忆向量行 + 记忆 FTS 行:按 companion_id 全清(delMemVectors/FTS);
//   - 运行期易失现场:core.Mem 内 round:<该角色会话id> 一并删(如有)。
//
// 该角色发言过的"群消息"保留正文,SenderCompanionID 置空(SET NULL);usage_records 保留(账单历史)。
func hDeleteCompanion(c *gin.Context) {
	// id := parseUintParam(c, "companionId")
	// companion := db.First(&model.Companion{}, id); if nil { 404 }
	// tx {
	//     db.Delete(&companion)                                            // ① 角色(级联见上方说明)
	//     delMemIndexesOf(companion.ID)                                    // ② 记忆向量+记忆 FTS 行(非外键,手动清)
	//     for g in groupsLeftWith(少于2成员) { dissolveGroup(g) }            // ③ 成员<2 的群连带解散
	// }
	// core.Mem.DeletePrefix("round:")                                     // ④ 会话级易失现场(含单聊会话 id)
	// // 头像/引用文件不物理删(解除引用,孤儿清理另行触发)
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}

// hSearchCompanions:GET /search/companions?q=(必填)—— 通讯录搜索。
func hSearchCompanions(c *gin.Context) {
	// q := c.Query("q"); if q == "" { respondErr(422, "缺少关键词"); return }
	// respond(c, 200, companionListWith(q))                                // 复用 hListCompanions 内核
}

// hListCompanionMemories:GET /companions/:companionId/memories —— 该角色记忆(兼容早期契约)。
// 说明:参数与 /memories 相同,companionId 固定过滤;委托 memories 查询内核。
func hListCompanionMemories(c *gin.Context) {
	// id := parseUintParam(c, "companionId")
	// if !companionExists(id) { 404 }
	// page := listMemoriesCore(companionID: id, q, type, before, limit)    // 见 memories.go
	// respond(c, 200, page)
}

// hTriggerProactive501:POST /companions/:companionId/proactive —— 501 占位。
// 说明:主动消息定时任务属第二阶段;路由先注册,前端可安全调用。
func hTriggerProactive501(c *gin.Context) {
	// respondErr(c, 501, CodeNotImplemented, "主动消息功能将在第二阶段提供")
}

// defaultCompanion:由 Create 请求补齐默认值(纯函数,单测目标;不赋 ID,主键自增)。
// 缺省原则:人设/记忆/聊天风格等按文件头与模型注释默认值补齐;
// Dispatch 等零值配置统一经 model.DispatchSettings.ApplyDefaults() 回落(常量见 model/dispatch.go)。
func defaultCompanion(req CompanionCreateReq) model.Companion {
	// c := model.Companion{Name: req.Name,
	//     Initial: firstRune(req.Name), Color: paletteHash(req.Name), Category: "自定义",
	//     FileID: req.FileID, APIConfigID: req.APIConfigID,             // FileID/APIConfigID 指针直填
	//     Persona: personaOrDefault(req.Persona),                       // relationship 缺省"朋友"
	//     MemorySettings: memorySettingsOrDefault(req.MemorySettings),
	//     ChatStyle: chatStyleOrDefault(req.ChatStyle),                 // allowSilent:true 等见 Companion 注释
	//     Dispatch: req.Dispatch.ApplyDefaults(),                       // 零值回落缺省(草稿阶段由 model 常量补齐)
	//     Proactive: proactiveOrDefault(req.Proactive), Capabilities: capabilitiesOrDefault(req.Capabilities)}
	// return c
	return model.Companion{}
}
