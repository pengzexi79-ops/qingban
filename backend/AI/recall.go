package ai

// 记忆召回(向量优先,FTS5 关键词降级)。
// 链路位置:单聊发送(注入前)、群聊成员发言前、记忆数据台"语义检索"调试页。
// 原则(BACKEND_HANDOFF §4.3 + api接口.md §6.2):embedding 本地生成/向量本地存储;
// embedding 不可用 → 自动降级 FTS5,命中 method="keyword",前端明确展示降级状态。
// 伪代码草稿:逻辑以函数体内伪代码注释占位。

import (
	"time"

	"qingban/model"
)

// RecallQuery:召回入参(service 从 /memories/search body 或发送链路上下文组装)。
type RecallQuery struct {
	// Query:用户当前消息/检索原文(≤500)。
	Query string
	// CompanionID:限定角色(空=全角色含全局记忆)。
	CompanionID string
	// TopK:返回条数上限(默认 8,最大 50)。
	TopK int
	// Threshold:相似度阈值(默认 0.65;keyword 为折算分阈值)。
	Threshold float64
	// Types:限定记忆类型(preference/event/relationship/summary;空=全部)。
	Types []string
	// OnlyConfirmed:只召回过已确认记忆(聊天链路 true,排除候选)。
	OnlyConfirmed bool
	// SinceDays:时间窗天数(对应 memorySettings.timeRangeDays,默认 365)。
	SinceDays int
}

// RecallResult:召回结果(检索响应 + 链路内部注入源)。
type RecallResult struct {
	// TraceID:本次召回追踪号(响应与日志共用)。
	TraceID string
	// Method:vector/keyword(整次召回统一口径)。
	Method string
	// Hits:命中列表(含 score;注入前可按 importance 排序截断)。
	Hits []model.MemoryHit
	// Summary:注入前压缩摘要(空=不压缩直接注原条目)。
	Summary string
}

// RecallMemories:执行一次记忆召回(调用点:memories/search 与 chat/group 注入前)。
func RecallMemories(q RecallQuery) (*RecallResult, error) {
	// if q.TopK <= 0 { q.TopK = 8 }                              // 归一入参
	// if q.Threshold <= 0 { q.Threshold = 0.65 }
	// if q.SinceDays <= 0 { q.SinceDays = 365 }
	// traceID := "trace-" + uuid4()
	// if vecAvail := vecExtLoaded() && embeddingModelAvailable(); vecAvail {
	//     emb, err := EmbedTexts([q.Query], "")                   // ② 向量路径
	//     if err == nil {
	//         hits := db.Raw(`SELECT m.id, m.title, m.content, m.type, m.companion_id,
	//                                 cosine(v.vector, ?) AS score
	//                          FROM memory_vectors v JOIN memories m ON m.id = v.memory_id
	//                          WHERE v.companion_id 过滤(空则含全局: v.companion_id IS NULL OR = q.CompanionID)
	//                            AND m.date >= today - q.SinceDays
	//                            AND (q.Types 为空 OR m.type IN q.Types)
	//                            AND (q.OnlyConfirmed OR m.status='confirmed')
	//                          ORDER BY score DESC LIMIT q.TopK`, emb[0])
	//         hits = filter(hits, score >= q.Threshold)           // 阈值过滤
	//         return &RecallResult{traceID, "vector", hits, summary: ""}, nil
	//     }
	//     log("向量路径失败,降级关键词", err)
	// }
	// // ③ 降级 FTS5(中文可先 LIKE 兜底,分词器见 init 建表注释)
	// rows := ftsSearch("fts_memories", q.Query,                  // MATCH 查询词(剥离引用标记)
	//     filter: 同上 companion/type/status/时间窗)
	// hits := rows → {score: 1/(1+bm25) /*归一*/, method: "keyword"}
	// // ④ 后处理:去重同内容;importance 降序;截断 TopK
	// // ⑤ summary:总字符 > 1200 → 调用摘要模型压缩;失败裁剪正文前 N 字(降级不报错)
	// return &RecallResult{traceID, "keyword", hits, summary}, nil
	return nil, nil // TODO(实现):见函数注释 ①~⑤
}

// EmbedTexts:文本批量向量化(embedding 统一出口;reindex 与召回共用)。
// 参数 modelName:embedding 模型 id;空=从默认配置推断。全部服务不可达 → PROVIDER_ERROR。
func EmbedTexts(texts []string, modelName string) ([][]float32, error) {
	// profile := pickProfile(modelName)                            // ① 可用 ApiProfile(embedding 或 chat 模型兜底)
	// dim := currentVecDim(); batch ≤ 32 条/次
	// req := {model: modelName, input: texts}
	// if profile.Protocol == ProtoOllama { POST base/api/embed }   // ② 协议分派
	// else { POST base/embeddings }(openai 兼容)
	// dims := unique(len(vec))
	// if len(dims) > 1 || (dim 已有 && dim != 新维) {               // ③ 维度变化 → 清空 memory_vectors 重建,防错乱
	//     truncate("memory_vectors")
	// }
	// return vecs, nil                                             // ④
	return nil, nil // TODO(实现):见函数注释 ①~④
}

// timeNow:当前时间(包内统一,便于将来切 clock 注入做任务测试)。
func timeNow() time.Time { return time.Now().UTC() }
