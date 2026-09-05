package ai

// 记忆候选提取与入库策略。
// 语义(BACKEND_HANDOFF §3/§4.3.3):发送消息 done 事件携带 memoryCandidates;
// 默认等用户确认;memorySettings.mode=automatic 的角色自动入库(source=自动提取)。
// 伪代码草稿:逻辑以函数体内伪代码注释占位。

import "qingban/model"

// candidatesBudget:候选提取输入预算(最近对话文本过长时先截尾)。
const candidatesBudget = 8000

// ExtractCandidates:从"用户消息+AI 回复+最近几轮"中提取值得长期保存的条目。
// 调用点:ChatWithCompanion 收尾(同步完成或异步 goroutine + hub 推 memory_candidates)。
// 实现方式:调用与 chat 同 profile 的模型,附"记忆提取提示词",要求严格 JSON 输出:
//
//	[{"type":"preference|event|relationship|summary","title":"≤28","content":"…","importance":0~1}]
func ExtractCandidates(turns []ChatTurn, sourceMessageID string, threshold float64) []model.MemoryDraft {
	// text := concatTurns(turns)                                 // ① 输入裁剪(保留完整轮次,≤candidatesBudget)
	// if containsNoMemory(text) { return nil }                   //    用户表达"别记住"类指令 → 整体跳过
	// result, err := chatOnce(extractPrompt(text))               // ② 模型调用(chatModel)
	// if err != nil { log("候选提取失败", err); return nil }      //    静默:记忆是增强能力,不能拖垮聊天
	// drafts := parseJSON(result)                                 // ③ 清洗:type 映射枚举/限长/importance 夹取 [0,1]
	// drafts = dedupe(drafts)                                     // ④ 防轰炸:与已确认记忆相似 >0.9 丢弃;单次 ≤3 条
	// for i := range drafts { drafts[i].SourceMessageID = sourceMessageID }   // ⑤ 来源消息回填
	// return drafts
	return nil // TODO(实现):见函数注释 ①~⑤
}

// PersistCandidates:按角色记忆模式决定候选去向:
//   - mode=automatic:直接落库(status=confirmed, source="自动提取", embeddingStatus=pending→入队 reindex)
//   - 其余(mode=hybrid/curated):仅随响应/SSE 下发,用户确认走 PATCH /memories/{id}{status:confirmed}
//
// 调用点:server/messages.go 收尾、群聊轮次(群聊提取本阶段可关闭)。
// 返回:落库成功的 Memory 与仍为草稿的候选。
func PersistCandidates(drafts []model.MemoryDraft, companion model.Companion) (saved []model.Memory, pending []model.MemoryDraft) {
	// if companion.MemorySettings.Mode != "automatic" { return nil, drafts }   // 非自动模式全部转前端确认
	// for d := range drafts {
	//     m := Memory{ID: memory-, CompanionId: &companion.ID, Type: d.Type,
	//                Title: d.Title, Content: d.Content, Date: today, Source: "自动提取",
	//                Importance: d.Importance, Status: confirmed,
	//                EmbeddingStatus: pending, SourceMessageId: &d.SourceMessageID}
	//     db.Insert(&m); enqueueIndex(m.ID)                        // 自动入库 + 异步向量化
	//     saved = append(saved, m)
	// }
	// return saved, nil
	return nil, nil // TODO(实现)
}
