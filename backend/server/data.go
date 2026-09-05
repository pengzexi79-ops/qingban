package server

// P1 数据迁移端点:导出 / 导入(兼容 API 格式与前端演示格式)/ 清空。
// 语义:导出=整库 JSON 快照(apiProfiles 脱敏,附件只记 fileId);
// 导入=按 id 去重合并;清空=confirm=true 全删。
// 伪代码草稿:逻辑以函数体内伪代码注释占位(实现时按需恢复 import)。

import (
	"github.com/gin-gonic/gin"

	"qingban/model"
)

// ExportData:GET /data/export 响应体(openapi ExportData)。
type ExportData struct {
	// Version:本格式版本(qingban_api_v1;旧品牌兼容命名)。
	Version string `json:"version"`
	// ExportedAt:导出时间。
	ExportedAt string `json:"exportedAt"`
	// User:用户资料。
	User *model.User `json:"user,omitempty"`
	// APIProfiles:脱敏列表(secretConfigured 占位,永不含 key)。
	APIProfiles []model.ApiProfile `json:"apiProfiles,omitempty"`
	// Companions:角色全量(不含派生字段)。
	Companions []model.Companion `json:"companions,omitempty"`
	// Groups:群(members 走 memberIds)。
	Groups []model.Group `json:"groups,omitempty"`
	// Messages:conversationId → 消息数组(整库恢复用)。
	Messages map[string][]model.Message `json:"messages,omitempty"`
	// Memories:记忆全量。
	Memories []model.Memory `json:"memories,omitempty"`
	// FilesManifest:附件清单(fileId 列表;二进制随整目录备份,本 JSON 只做引用一致性)。
	FilesManifest []string `json:"filesManifest,omitempty"`
	// Settings:kv 非敏感键(如默认 profile id)。
	Settings map[string]string `json:"settings,omitempty"`
}

// hExportData:GET /data/export —— 全量导出 JSON 快照。
func hExportData(c *gin.Context) {
	// out := ExportData{Version: "qingban_api_v1", ExportedAt: nowRFC3339()}
	// out.User = firstUser()
	// out.Companions = db.Find(Companion{})                             // 实体全量
	// out.Groups = db.Find(Group{})                                     // MemberIds 由 group_members 联查填充
	// out.Messages = groupByConversation(db.Find(Message{}))             // conversation_id → []Message
	// out.Memories = db.Find(Memory{})
	// profiles := db.Find(ApiProfile{})                                 // ① 脱敏:
	// for p := range profiles { p.APIKeyEnc = ""; p.SecretConfigured = hadKey(p) }  // 永不含 key 密文
	// out.APIProfiles = profiles
	// out.FilesManifest = db.Pluck(File{}, "id")                        // 仅列 id(附件二进制随 dataDir 备份)
	// out.Settings = kvDump(排除: token/密钥类键)
	// respond(c, 200, out)
}

// ImportBody:POST /data/import 请求体。
type ImportBody struct {
	// Payload:导出 JSON 原文(兼容:本格式 v1 / 前端演示 qinban_frontend_v4/v3/v2)。
	Payload map[string]any `json:"payload"`
	// Merge:true=允许与已有数据合并(目标空间有数据且未提供 → 409)。
	Merge *bool `json:"merge,omitempty"`
}

// hImportData:POST /data/import —— 导入(按 id 去重;返回迁移统计)。
func hImportData(c *gin.Context) {
	// var body ImportBody; if !bind(c, &body) { return }
	// payload := body.Payload; if payload == nil { respondErr(422, "缺少 payload"); return }
	// version := detectVersion(payload)                                  // ① 探测顶层 version/兼容键
	// hasData := (count(companions)+count(messages)) > 0
	// if hasData && !merge { respondErr(409, "目标空间已有数据,需确认合并或先清空"); return }
	// stats := importPayloadCore(payload, version)                       // ③ 见下方依赖序(单事务)
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 200, stats)
}

// importPayloadCore:导入执行内核(单事务,幂等去重)。
// 依赖序:
//
//	a) users:已有则取较新(nickname/settings 覆盖;幂等跳过)
//	b) apiProfiles:仅脱敏占位导入(无密钥;保留 default 绑定)
//	c) companions → conversations 缓存行(缺失补建)→ groups+group_members
//	d) messages:conversation 不存在 → 跳过(孤儿);补 unread/摘要不做,统一按 0
//	e) memories:companionId 悬挂 → 置空(转全局)
//	f) 全部 INSERT OR IGNORE(按 id 去重;重复计入 skipped)
func importPayloadCore(payload map[string]any, version string) ImportStats {
	// TODO(实现):见函数注释;前端演示键差异(v4/v3/v2)在版本归一器内逐版本映射,
	//  样例结构见 frontend/js/store.js;avatarImage(base64)可选落盘 files 或保留字段。
	return ImportStats{}
}

// ImportStats:导入迁移统计(openapi ImportResult)。
type ImportStats struct {
	// Version:识别到的数据版本。
	Version string `json:"version"`
	// Counts:各实体成功迁移条数。
	Counts ImportCounts `json:"counts"`
	// Skipped:按 id 去重跳过的已有记录数。
	Skipped int64 `json:"skipped"`
}

// ImportCounts:迁移计数分项。
type ImportCounts struct {
	Companions    int64 `json:"companions"`
	Groups        int64 `json:"groups"`
	Conversations int64 `json:"conversations"`
	Messages      int64 `json:"messages"`
	Memories      int64 `json:"memories"`
	Files         int64 `json:"files"`
}

// hClearData:DELETE /data?confirm=true —— 清空本地业务数据(不可恢复)。
func hClearData(c *gin.Context) {
	// if c.Query("confirm") != "true" { respondErr(400, "需 confirm=true"); return }
	// tx { 清空业务表:companions/groups/group_members/rounds/conversations/messages/
	//      memories/files 行/usage_records }
	// os.RemoveAll({DataDir}/files)                                       // 附件随业务清除(提示先导出)
	// // 保留:users 与 kvs(bootstrap done 标记)→ 不回引导页,资料与设置仍在
	// //       (语义分歧点:如需"彻底重置"走未来 DELETE /me,注释供评审)
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}
