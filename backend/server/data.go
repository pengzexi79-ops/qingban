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
// 快照语义(v3):主实体(users/companions/groups/memories/configs)+ 会话行 +
// 各关系表(message_files/message_mentions/group_members/rounds/round_speakers/
// member_cursors/chat_short_memories)整库可恢复;附件二进制随数据目录整体备份,清单只列 id。
type ExportData struct {
	// Version:本格式版本(qingban_api_v1;旧品牌兼容命名)。
	Version string `json:"version"`
	// ExportedAt:导出时间。
	ExportedAt string `json:"exportedAt"`
	// User:用户资料。
	User *model.User `json:"user,omitempty"`
	// ModelConfigs:模型配置(脱敏:APIKey 已置空)。
	ModelConfigs []model.APIConfig `json:"modelConfigs,omitempty"`
	// Companions:角色全量(不含派生字段)。
	Companions []model.Companion `json:"companions,omitempty"`
	// Groups:群全量(成员关系走 GroupMembers)。
	Groups []model.Group `json:"groups,omitempty"`
	// GroupMembers:群成员关系(含入群时间)。
	GroupMembers []model.GroupMember `json:"groupMembers,omitempty"`
	// Conversations:会话行(消息归属/未读/置顶快照)。
	Conversations []model.Conversation `json:"conversations,omitempty"`
	// Messages:消息全量(按 id 序;附件/点名经下面两张关系表恢复)。
	Messages []model.Message `json:"messages,omitempty"`
	// MessageFiles/MessageMentions:消息附件/点名关系。
	MessageFiles    []model.MessageFile    `json:"messageFiles,omitempty"`
	MessageMentions []model.MessageMention `json:"messageMentions,omitempty"`
	// Rounds/RoundSpeakers:轮次与发言明细。
	Rounds        []model.Round        `json:"rounds,omitempty"`
	RoundSpeakers []model.RoundSpeaker `json:"roundSpeakers,omitempty"`
	// MemberCursors/ShortMemories:已读水位与短期记忆。
	MemberCursors []model.MemberCursor `json:"memberCursors,omitempty"`
	ShortMemories []model.ShortMemory  `json:"shortMemories,omitempty"`
	// Memories:长期记忆全量。
	Memories []model.Memory `json:"memories,omitempty"`
	// FilesManifest:附件 id 清单(二进制随 dataDir 目录备份,JSON 只做引用一致性)。
	FilesManifest []uint `json:"filesManifest,omitempty"`
	// Settings:config_kvs 非敏感键的"明文视图"(key→值;敏感键不导出)。
	Settings map[string]string `json:"settings,omitempty"`
}

// hExportData:GET /data/export —— 全量导出 JSON 快照。
func hExportData(c *gin.Context) {
	// out := ExportData{Version: "qingban_api_v1", ExportedAt: nowRFC3339()}
	// out.User = firstUser()
	// out.Companions = db.Find(&[]model.Companion{})                    // 实体全量(关系走关系段)
	// out.Groups = db.Find(&[]model.Group{})
	// out.GroupMembers = db.Find(&[]model.GroupMember{})
	// out.Conversations = db.Find(&[]model.Conversation{})
	// out.Messages = db.Order("id").Find(&[]model.Message{})
	// out.MessageFiles = db.Find(&[]model.MessageFile{})
	// out.MessageMentions = db.Find(&[]model.MessageMention{})
	// out.Rounds = db.Find(&[]model.Round{}); out.RoundSpeakers = db.Find(&[]model.RoundSpeaker{})
	// out.MemberCursors = db.Find(&[]model.MemberCursor{}); out.ShortMemories = db.Find(&[]model.ShortMemory{})
	// out.Memories = db.Find(&[]model.Memory{})
	// cfgs := db.Find(&[]model.ModelConfig{})                            // ① 脱敏:
	// for i := range cfgs { cfgs[i].APIKey = "" }                        // 永不含 key 密文
	// out.ModelConfigs = cfgs
	// db.Model(&model.File{}).Pluck("id", &out.FilesManifest)            // 附件二进制随 dataDir 备份
	// out.Settings = kvDump(排除:密钥/敏感键)                              // config_kvs 明文视图
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
// 依赖序(v3 快照段,见 ExportData):
//
//	a) users:已有则取较新(nickname/settings 覆盖;幂等跳过)
//	b) modelConfigs:脱敏占位导入(无密钥;恢复默认配置引用)+ config_kvs 明文段
//	c) companions(自增重映射 old→new)→ groups → group_members(joined_at)
//	d) conversations:companion_id/group_id 按映射重指向;缺失的按归属补建
//	e) messages:重映射 conversation_id;先建行再重建 message_files/message_mentions(校验存在)
//	f) rounds→round_speakers(消息 id 重映射);memories(companion 悬挂→置空转全局)
//	g) member_cursors/chat_short_memories:会话/成员重映射后插入
//	h) 前端演示格式(v4/v3/v2)在版本归一器内逐版本映射到上述段(样例见 frontend/js/store.js)
func importPayloadCore(payload map[string]any, version string) ImportStats {
	// TODO(实现):见函数注释;演示数据 avatarImage(base64)可选落盘 files 或忽略。
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
	// tx { 按依赖序清空(先子后父或直接按表删,SQLite 外键级联生效下可只删主实体):
	//     model_configs/companions/groups(级联 group_members/group 会话及其消息链/
	//     rounds/memories 等)/conversations(遗留孤儿)/messages 兜底/memories 兜底/
	//     files/usage_records/member_cursors/chat_short_memories }
	// os.RemoveAll({DataDir}/files)                                       // 附件随业务清除(提示先导出)
	// // 保留:users 与 config_kvs(bootstrap done 标记)→ 不回引导页,资料与设置仍在
	// //       (语义分歧点:如需"彻底重置"走未来 DELETE /me,注释供评审)
	// hub.Publish(EventThreadsChanged, {})
	// respond(c, 204)
}
