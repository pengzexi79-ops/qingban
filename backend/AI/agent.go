package ai

// Agent 运行时管理(青伴 × Eino ADK)。
// 设计依据(Eino 官方文档 2026-05 版,quick_start ch02/ch03 + multi_agent_hosting):
//   - 框架层:Agent 接口 + adk.NewChatModelAgent + adk.NewTypedRunner[M](Run/Query);
//     历史存储是"业务层职责",Eino 不提供 Session——青伴历史在 SQLite messages,
//     记忆在 memories+召回,天然符合官方推荐的业务层模式。
//   - 因此本包职责 = "多 agent 的运行注册表":懒构建 / 配置哈希失效重建 /
//     并发防抖(singleflight)/ 驱逐 / 销毁 / 统计;不缓存会话消息。
// 映射:一个 Companion = 一个可缓存的 Agent runtime;一次发言 = 取 runtime → 组装历史 → runner.Run。
//
// 依赖(实现阶段 go get 并核对版本,当前 Eino v0.9 系):
//   github.com/cloudwego/eino                      // schema/msgops/adk(agentic-runtime)
//   github.com/cloudwego/eino-ext/components/model/openai  // OpenAI 兼容(Ollama 走 /v1)
//   github.com/cloudwego/eino-ext/components/model/ollama  // 本地 Ollama 原生(可选)
//   官方可运行样例:github.com/cloudwego/eino-examples/quickstart/chatwitheino/cmd/ch02|ch03
// 伪代码草稿:装配细节以函数体内伪代码注释占位;类型 M 实现时取 *schema.AgenticMessage。

import (
	"context"
	"sync"
	"time"

	"qingban/model"
)

// AgentRuntime:某个角色装配好的 Eino 运行时(实现时为泛型组合的轻量包装):
//
//	agent *adk.ChatModelAgent(或 TypedChatModelAgent[M]) + runner *adk.TypedRunner[M]
//
// 设计:agent 与 runner 在装配时一次构建、长期复用;每个 Run 调用只传消息列表。
type AgentRuntime struct {
	// AgentName:Eino agent 的名字(取 companionID,便于事件/回调区分角色)。
	AgentName string
	// Generation:构建代数(每次重建 +1;用来丢弃并发中已失效的旧运行结果)。
	Generation int64
	// 内部字段(实现时填充,注释说明用途):
	//   a  any // *adk.ChatModelAgent——仅作占位;避免草稿依赖 Eino 类型
	//   r  any // *adk.TypedRunner[M]
	//   cb any // *logCallback——回调见 Eino ch06(用量/追踪挂这里)
}

// BuildKey:判定"配置是否变了"的哈希输入。作用:同 key 说明可复用旧 runtime。
type BuildKey struct {
	// CompanionID:角色 id。
	CompanionID string
	// PersonaHash:人设+画像+风格+能力 的序列化哈希(改人设 → 重建)。
	PersonaHash string
	// ProfileHash:绑定的 API 配置哈希(model/baseURL/密钥轮换/温度 → 重建)。
	ProfileHash string
	// UpdatedAt:companions.updated_at(兜底,任何 PATCH 都触失效)。
	UpdatedAt time.Time
}

// BuildArgs:一次构建所需全部输入(Manager.build 的入参)。
type BuildArgs struct {
	// Companion:角色实体(人设/记忆设置/能力/聊天风格)。
	Companion model.Companion
	// Profile:绑定 API 配置(空=回落默认本地配置)。
	Profile *model.ApiProfile
	// SecretKey:解密后的 API 密钥(仅构建期取用,不落 runtime 之外)。
	SecretKey string
	// UserPersona:用户画像(Instruction 静态段的一部分)。
	UserPersona string
	// Tools:角色可用工具(第二阶段:记忆检索/联网等;本阶段为空)。
	Tools []any // 实现时 []*agent.Tool(见 Eino ch04 Tools)
}

// Manager:多 agent 运行注册表(进程内单例,init.NewApp 时创建并注入)。
// 并发模型:读多写少 → RWMutex + 每 key 单飞(构建中并发请求共享同一构建)。
type Manager struct {
	mu      sync.RWMutex                                                     // 保护 entries/version
	entries map[string]*AgentEntry                                           // companionID → 运行条目
	buildFn func(ctx context.Context, spec BuildArgs) (*AgentRuntime, error) // 可注入的构建器(测试替身/延迟依赖)
	// 统计字段(供 /me/stats、运维面板):
	hits       int64 // 缓存命中次数
	rebuilds   int64 // 重建次数
	buildFails int64 // 构建失败次数
}

// AgentEntry:注册表条目。
type AgentEntry struct {
	// Key:该条目对应的构建键(重建判断用)。
	Key BuildKey
	// RT:装配好的运行时。
	RT *AgentRuntime
	// LastUsedAt:最近一次被取用时刻(未来可做空闲驱逐 LRU)。
	LastUsedAt time.Time
}

// NewManager:创建注册表(调用点:init.NewApp 第 7 步之后;buildFn 默认实现见 buildRuntime)。
func NewManager(buildFn func(ctx context.Context, spec BuildArgs) (*AgentRuntime, error)) *Manager {
	// return &Manager{entries: make(map[string]*AgentEntry), buildFn: buildFn}
	return &Manager{}
}

// Get:取(或按需构建)某角色的运行时。调用点:server 单聊链路与 group_round 逐成员发言。
// 语义:
//
//	① key := buildKeyOf(spec)
//	② 读锁查 entries[key.CompanionID]:
//	   - 命中且 entry.Key == key → 更新 LastUsedAt,直接复用(缓存命中)
//	③ 未命中或哈希变了 → 升级写锁,double-check 后走 ④(防止并发重复构建)
//	④ 构建:buildFn(ctx, spec)(内部对同一 key 只构建一次,其余请求等待同一结果)
//	⑤ 新条目入表;失败 → 返回错误,调用方走 fallback(不缓存失败)
//
// 返回条目时同时返回 generation(调用方写日志/追踪用)。
func (m *Manager) Get(ctx context.Context, spec BuildArgs) (*AgentEntry, error) {
	// TODO(实现):见函数注释 ①~⑤
	return nil, nil
}

// Evict:驱逐某角色 runtime(调用点:companion PATCH/DELETE 后、profile 变更波及)。
// 语义:删除条目;若旧 runtime 支持 Close(如流资源),在锁外执行清理。
func (m *Manager) Evict(companionID string) {
	// m.mu.Lock(); entry := m.entries[companionID]; delete(m.entries, companionID); m.mu.Unlock()
	// if entry != nil { closeRuntime(entry.RT) }   // 锁外销毁,避免持有锁做 IO
}

// EvictByProfile:某 API 配置被改/删时,把所有绑定它的角色一并驱逐(调用点:api_profiles PATCH/DELETE)。
func (m *Manager) EvictByProfile(profileID string) {
	// m.mu.Lock()
	// ids := []string{}
	// for id, e := range m.entries { if e.Key.ProfileHash 由该 profile 计算 { ids = append(ids, id) } }
	// for id := range ids { delete(m.entries, id) }
	// m.mu.Unlock()
	// // 实现注意:需在 BuildKey.ProfileHash 之外冗余记录 profileID→[]companionID 反查索引,
	// // 否则需全表比对(本地角色量小,全表可接受;注释先行说明)
}

// Clear:清空全部运行时(调用点:/data 清空、注销;或空闲回收策略)。
func (m *Manager) Clear() {
	// m.mu.Lock(); entries := m.entries; m.entries = map[string]*AgentEntry{}; m.mu.Unlock()
	// for _, e := range entries { closeRuntime(e.RT) }
}

// Stats:注册表统计(缓存命中/重建/失败;供诊断页与日志)。
func (m *Manager) Stats() map[string]int64 {
	// return {"hits": m.hits, "rebuilds": m.rebuilds, "buildFails": m.buildFails, "live": len(m.entries)}
	return nil
}

// buildRuntime:默认构建器(Eino 装配点,伪代码;实现时按 ch02 样例核对 API):
//
//	① cm, err := newChatModel(spec)                    // 按 Profile 装配模型组件:
//	   //   openai 兼容:model/openai.NewChatModel(ctx, &openai.ChatModelConfig{
//	   //       BaseURL: spec.Profile.BaseURL, Model: spec.Profile.ChatModel,
//	   //       APIKey: spec.SecretKey, ...})            // Ollama 走同端点(/v1),或 ollama.NewChatModel
//	   //   失败 → common.PROVIDER_ERROR(不缓存)
//	② instruction := staticPrompt(spec)                // 静态 Instruction(见下),构建期固定
//	③ agent := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
//	      Name: spec.Companion.ID, Description: companion.Tagline, Instruction: instruction, Model: cm})
//	   // 如需工具:config.Tools = spec.Tools(ch04)
//	④ runner := adk.NewTypedRunner[M](adk.TypedRunnerConfig[M]{Agent: agent, EnableStreaming: true})
//	⑤ 挂回调(用量/追踪:runner/agent 回调见 ch06,接入 core.Log 与 usage 落库点)
//	⑥ 返回 &AgentRuntime{AgentName: ID, a: agent, r: runner}
func buildRuntime(ctx context.Context, spec BuildArgs) (*AgentRuntime, error) {
	// TODO(实现):见函数注释 ①~⑥
	return nil, nil
}

// staticPrompt:Instruction 静态段(构建期固定)。
// 内容 = 人设 persona 六段 + 用户画像 + 风格指令(chatStyle)+ 边界;
// 说明:"动态记忆/摘要"不进 Instruction(否则每次都要重建)——动态段由业务层
// 在每次 Run 前注入消息列表(见 RunTurn),与官方 ch03 模式一致。
func staticPrompt(spec BuildArgs) string {
	// b := strings.Builder{}
	// b.WriteString("你是" + companion.Name + "。\n")
	// b.WriteString(personaSection(companion.Persona))        // 六段,空段跳过
	// if spec.UserPersona != "" { b.WriteString("\n[用户画像] " + spec.UserPersona) }
	// b.WriteString(styleSection(companion.ChatStyle))        // markdown/表达约束等
	// b.WriteString(safetySection(companion.Persona.Boundaries, companion.Persona.ForbiddenTopics))
	// return b.String()
	return ""
}

// RunTurn:一次完整发言(业务层入口,替代原 chat.go 内直接对 provider 的调用)。
// 调用点:server/messages.go 的同步/SSE 分支、group_round 逐成员调用(重构后的目标形态)。
// 入参 history 已含:DB 取回的最近 N 轮 + 记忆命中(顶部以"系统提示格式"插入的
// 记忆区消息);实现时对照 eino msgops/schema 是否支持 system 型 AgenticMessage:
//
//	支持 → 首条发 system(记忆区+摘要);不支持 → 退化为把记忆区拼进首条 user 消息前缀。
func RunTurn(ctx context.Context, entry *AgentEntry, history []any) (text string, events []any, err error) {
	// // runner := entry.RT.r             // *adk.TypedRunner[M]
	// // events := runner.Run(ctx, msgops.NormalizeMessagesForModelInput(history))
	// // for {                             // 消费 AsyncIterator[*AgentEvent]:
	// //     ev, ok := events.Next(); if !ok { break }
	// //     if ev.Err != nil { return 失败 }              // 映射 PROVIDER_ERROR
	// //     if ev.Output != nil && ev.Output.MessageOutput != nil {
	// //         增量文本 → 累积 text;同步转发(SSE delta / hub typing)
	// //     }
	// //     if ev.Action != nil { /* 工具/中断/移交:第二阶段 */ }
	// // }
	// // usage/token 统计从回调侧(第⑥步)收集,随 ev 一并返回供 server 落 usage_records
	return "", nil, nil // TODO(实现)
}

// Snapshot:列出当前全部活动角色 id(运维/诊断)。
func (m *Manager) Snapshot() []string {
	// m.mu.RLock(); defer m.mu.RUnlock()
	// ids := make([]string, 0, len(m.entries)); for id := range m.entries { ids = append(ids, id) }
	// return ids
	return nil
}

// 注:本文件草稿不含 Eino 真实 import(依赖未加入 go.mod)。实现顺序建议:
//   1) go get eino + eino-ext(model/openai) → 对照 eino-examples ch02 跑通单 agent
//   2) 用 tests 子包(如 tests/eino/)起最小可运行样例(仿 example: 本机 Ollama 一问一答)
//   3) 再逐块替换本文件 TODO。
