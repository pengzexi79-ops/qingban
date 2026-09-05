package ai

// 批次对话调度引擎(单聊与群聊统一)—— 伪代码草稿,调度细节以注释占位。
// 语义依据:docs/batch_dispatch_design.md。核心规则:
//   - 群聊用户消息@谁 → 点名:立即只调度被点名成员;无@ → 攒批:先落库,
//     到"不定时"触发时机(静默窗/攒批硬顶/群冷却放行)再整批投喂;
//   - 投喂 = 读该成员"已读水位(游标)"之后全部未读,拼成一次模型调用;
//   - 消费即归档:调用结束(回复/静默/失败兜底)即推进水位,已读消息永不重投
//     (其内容只经短期记忆任务压缩留存,见 summary.go);
//   - 静默:模型允许"已读不回",照常消费并排定短期记忆总结(根治一条消息回一条的鬼畜);
//   - AI 的发言只进入其他成员的未读池,不催生新一轮(AI 互聊防滚雪球);
//   - 失败重试:同成员连续 maxConsecutiveFails 次失败后按"已读不回"兜底消费,防死循环。
//
// 依赖面:数据经 DispatchRepo(接入层 GORM 实现,单测内存假实现),模型调用经
// TurnRunner 注入(装配期接 RunTurn/agent.go 或 ChatWithCompanion/chat.go)。
// 本文件只落类型、常量与函数签名,实现顺序建议见文件尾注。

import (
	"context"
	"sync"
	"time"

	"qingban/core"
	"qingban/model"
)

// 引擎缺省参数(会话/角色级配置缺省见 model.DispatchSettings 常量)。
const (
	// engineTick:事件循环轮询间隔(静默窗到点粒度)。
	engineTick = 500 * time.Millisecond
	// maxConsecutiveFails:同成员连续失败上限;达上限按"已读不回"兜底消费。
	maxConsecutiveFails = 3
	// batchCharCap:投喂正文安全字数上限(超出部分不再截断,前段靠短期记忆兜底)。
	batchCharCap = 6000
	// turnTimeout:单成员一次模型调用超时(与 provider 端 60s 一致)。
	turnTimeout = 60 * time.Second
)

// ---- 仓储接口(接入层实现;GORM 落点见 server 装配 TODO) ----

// ConvInfo:会话静态信息(每次投喂前现读,成员/冷却/配置改动即时生效)。
type ConvInfo struct {
	// Kind:companion(单聊)/group(群聊)——由会话行归属(companion_id/group_id)推断。
	Kind string `json:"kind"`
	// MemberIDs:AI 成员 id(companions.id;单聊=[会话归属角色],群聊=全部 AI 成员)。
	MemberIDs []uint `json:"memberIds"`
	// CooldownSeconds:群聊两轮投喂最小间隔(groups.strategy.cooldownSeconds)。
	CooldownSeconds int `json:"cooldownSeconds"`
}

// DispatchRepo:调度引擎的持久化依赖面(单机 SQLite+GORM 基准)。
type DispatchRepo interface {
	// Conversation:会话信息(类型/成员/群冷却;conversationID=conversations.id)。
	Conversation(ctx context.Context, conversationID uint) (ConvInfo, error)
	// CompanionsByIDs:批量取角色(批次配置/静默开关/人设);未知 id 允许缺行。
	CompanionsByIDs(ctx context.Context, ids []uint) (map[uint]model.Companion, error)
	// MessagesAfter:某会话 id>afterID 的消息(id 升序=时间正序)。
	MessagesAfter(ctx context.Context, conversationID uint, afterID uint) ([]model.Message, error)
	// MessageMentionIDs:某条消息的点名成员(群聊 user 消息;空=无点名)。
	MessageMentionIDs(ctx context.Context, messageID uint) ([]uint, error)
	// Cursor:某接收方已读水位(未找到 found=false,水位视为 0;readerID 用户=model.ReaderUserID)。
	Cursor(ctx context.Context, conversationID, readerID uint) (model.MemberCursor, bool, error)
	// SaveCursor:写入/推进已读水位。
	SaveCursor(ctx context.Context, c model.MemberCursor) error
	// Memory:某成员在某会话的短期记忆(无则 found=false)。
	Memory(ctx context.Context, companionID, conversationID uint) (model.ShortMemory, bool, error)
	// SaveMemory:覆盖写入新一期短期记忆。
	SaveMemory(ctx context.Context, m model.ShortMemory) error
	// SaveAssistantMessage:落库一条 AI 回复(引擎填归属与正文,接入层补 ID 后广播)。
	SaveAssistantMessage(ctx context.Context, m *model.Message) error
}

// TurnRunner:一次"给某成员的模型调用"。返回 text=回复正文,silent=true=成员已读不回。
// 装配期接入 RunTurn(agent.go,经 Eino)或 ChatWithCompanion(chat.go)。
type TurnRunner func(ctx context.Context, comp model.Companion, kind string, msgs []ChatMessage) (text string, silent bool, err error)

// DispatchHooks:引擎旁路回调(接入层注入;nil 自动跳过)。
type DispatchHooks struct {
	// OnRoundStart:一轮投喂开始(round_start 语义),参数 convID+目标成员。
	OnRoundStart func(conversationID uint, readerIDs []uint)
	// OnReply:AI 回复落库后回调(接入层广播 new_message)。
	OnReply func(conversationID uint, m model.Message)
	// OnSilent:成员已读不回(诊断/审计,无前端事件)。
	OnSilent func(conversationID, readerID uint)
	// OnConsumed:成员消费推进到某水位(含静默/失败兜底;诊断/审计)。
	OnConsumed func(conversationID, readerID uint, throughID uint)
	// LogError:非致命错误上报(接入层接 core.Log)。
	LogError func(msg string, args ...any)
}

// DispatchSettingsOf:取角色批次配置(零值回落 model.DispatchSettings 缺省常量)。
func DispatchSettingsOf(c model.Companion) model.DispatchSettings {
	return c.Dispatch.ApplyDefaults() // 缺省值见 model/dispatch.go Default* 常量
}

// ---- 纯决策/构建函数(签名先行,算法见注释;单测目标,实现顺序见文件尾注) ----

// dueBySilence:距最后一条用户消息 ≥ 静默窗 → 可投喂(设计 §3.2 触发条件 1)。
func dueBySilence(idle, debounce time.Duration) bool {
	// return debounce > 0 && idle >= debounce
	return false // TODO(实现)
}

// dueByBacklog:未读条数/字数达到攒批硬顶 → 不等静默窗立即投喂(设计 §3.2 触发条件 2)。
func dueByBacklog(count, chars, maxCount, maxChars int) bool {
	// if maxCount > 0 && count >= maxCount { return true }   // 条数硬顶
	// if maxChars > 0 && chars >= maxChars { return true }   // 字数硬顶
	// return false
	return false // TODO(实现)
}

// MentionReaderIDs:把消息点名展开为"本批要投喂的成员"(设计 §3.1)。
// 语义:
//
//	① mentionAll=true(@全员/手动触发轮次)→ 返回全部成员,everyone=true;
//	② 否则把 mentions(被点名成员 id)与成员集合求交:交集外忽略(已退群/删号),去重;
//	③ 返回 targets 为空 = 点名的都不在群(调用方按忽略处理,消息仍作攒批内容)。
func MentionReaderIDs(memberIDs []uint, mentionAll bool, mentions []uint) (targets []uint, everyone bool) {
	// if mentionAll { return memberIDs, true }                        // ①
	// members := set(memberIDs); seen := map[uint]bool{}
	// for _, id := range mentions {                                   // ②
	//     if members[id] && !seen[id] { targets = append(targets, id); seen[id] = true }
	// }
	// return targets, false                                           // ③
	return nil, false // TODO(实现):见函数注释 ①~③
}

// convDebounce:用户"输入完成"判定窗口。
// 说明:统一取 core.IdleWindow(默认 2s,core 编译期可注入);角色级 DispatchSettings
// DebounceSeconds 已不参与该判定(字段保留作历史兼容,见 model/dispatch.go 注记)。
func convDebounce(_ map[uint]model.Companion, _ []uint) time.Duration {
	return core.IdleWindow
}

// BuildBatchLines:未读批渲染为"一次性拼接文本"(设计 §2.2)。
// 语义:
//
//	① 群聊每行 "[发送者名]: 内容"(用户消息发送者为空 → "我";@点名引用保留在正文原文,
//	   @全员消息头部注记 [所有人]);发送者名由装配层按 SenderCompanionID 联查填充,
//	   本函数接收已解析为"显示名"的轻量输入(见 composeBatchLines 的入参说明);
//	② 单聊为裸文本多行(不需要前缀);
//	③ 超过 maxChars 保留末尾 maxChars 字符(前段已被短期记忆覆盖,见 batchCharCap)。
//
// 超限截断单位按字符(rune)计,不做字节截断。
func BuildBatchLines(kind string, batch []model.Message, maxChars int) string {
	// b := strings.Builder{}
	// for _, m := range batch {                                // ① 逐条拼接
	//     text := m.Content
	//     if kind == "group" {
	//         name := displayNameOf(m.SenderCompanionID); if name == "" { name = "我" }
	//         if m.MentionAll { text = "[所有人] " + text }
	//         b.WriteString("[" + name + "]: " + text + "\n")
	//     } else { b.WriteString(text + "\n") }                // ② 单聊裸文本
	// }
	// s := b.String()
	// if maxChars > 0 && RuneCount(s) > maxChars { s = 取末尾 maxChars 字符 }  // ③
	// return TrimRight(s, "\n")
	return "" // TODO(实现):见函数注释 ①~③
}

// ShortMemoryBlock:短期记忆 → 注入文本(记忆区头部,含期数与覆盖区间便于模型判断新旧)。
// 说明:空记忆返回空串(调用方跳过记忆段,composeTurnMessages 只发正文一段)。
func ShortMemoryBlock(m *model.ShortMemory) string {
	// if m == nil || TrimSpace(m.Content) == "" { return "" }
	// return "【会话短期记忆 第" + Itoa(m.Generation) + "期(压缩至消息 #" + m.CoveredToMessageID + ")】\n" + m.Content
	return "" // TODO(实现)
}

// composeTurnMessages:一次投喂的模型消息序列(设计 §2.2 三段式中的"动态段")。
// 说明:人设/风格/边界等 system 指令由 TurnRunner 装配层按 comp 前置(见 chat.go
// buildSystemPrompt,前缀稳定是厂商 KV 命中的关键);本函数只产出
// "会话记忆区 + 本批正文"两个 user 段(记忆区为空时仅正文一段)。
func composeTurnMessages(kind, memText, body string) []ChatMessage {
	// out := []ChatMessage{}
	// if TrimSpace(memText) != "" { out = append(out, {Role: user, Content: memText}) }
	// out = append(out, {Role: user, Content: body})
	// return out
	return nil // TODO(实现)
}

// ---- 引擎状态(字段语义先行;编排细节见方法注释) ----

// DispatchState:会话的进程内动态(待投/计时/点名队列/连败计数)。
// 说明:消息本体与水位在 DB,进程重启后由 member_cursors 完全重建,本结构可丢。
type DispatchState struct {
	conversationID uint
	kind           string
	// pending:有待投批(用户新消息到达置位;整批投喂开始时清除)。
	pending bool
	// targetQueue:点名目标队列(点名撞群冷却时排队;放行后只投队列成员,不扩大为全员)。
	targetQueue []uint
	// lastUserMsgAt:最后一条用户消息到达时刻(静默窗起点;零值=无)。
	lastUserMsgAt time.Time
	// running:该会话正在投喂中(防重入;点名打断场景自然并入下一轮攒批)。
	running bool
	// lastFlushAt:最近一次投喂开始时刻(群冷却判断)。
	lastFlushAt time.Time
	// failStreak:每位成员连续失败次数(达上限→已读不回兜底)。
	failStreak map[uint]int
}

// Dispatcher:批次调度引擎(进程内单例,init 装配期创建并 Start)。
// 并发模型(实现时):单事件循环 goroutine(engineLoop)串行驱动全部会话:
// 心跳 engineTick 轮询"到点"会话,NotifyMessage 经 wake 通道即时唤醒;
// 本地单机几十会话规模,无需更重并发模型。
type Dispatcher struct {
	// repo/run/now/hooks:持久化依赖、模型调用器、时钟注入(测试)、旁路回调。
	repo  DispatchRepo
	run   TurnRunner
	now   func() time.Time
	hooks DispatchHooks
	// mu/states:会话状态表(键=conversations.id)。
	mu     sync.Mutex
	states map[uint]*DispatchState
	// wake/stop/wg:事件循环信号与退出编排。
	wake chan struct{}
	stop chan struct{}
	wg   sync.WaitGroup
	// summaryDue:短期记忆任务时刻表 key=conv\x00companion(见 summary.go)。
	summaryDue map[string]time.Time
}

// NewDispatcher:构造引擎。调用点:init.NewApp 运行时装配完成后。
// 说明:run 为 nil 时引擎只消费不回复(降级/调试态,配合 LogError 可见)。
func NewDispatcher(repo DispatchRepo, run TurnRunner, hooks DispatchHooks) *Dispatcher {
	// if hooks.LogError == nil { hooks.LogError = func(...){} }   // 空回调兜底
	// return &Dispatcher{repo, run, now: time.Now, hooks, states: map[...]string,
	//                    wake: make(chan struct{},1), stop: make(chan struct{}),
	//                    summaryDue: map[string]time.Time{}}
	return &Dispatcher{} // TODO(实现)
}

// Start:启动引擎后台循环(engineLoop + summaryLoop,见 summary.go)。
func (d *Dispatcher) Start() {
	// d.wg.Add(2); go d.engineLoop(); go d.summaryLoop()
}

// Close:停止引擎并等待在途调用结束(进程优雅退出时调用)。
func (d *Dispatcher) Close() {
	// close(d.stop); d.wg.Wait()
}

// NotifyMessage:一条消息落库后通知引擎(server 发送链路落库后调用)。语义:
//
//	① AI 消息(role=assistant)→ 只自然进入他人未读池,不置 pending(不催生新一轮);
//	② 用户消息 → 置 pending、刷新 lastUserMsgAt,唤起事件循环(wakeLoop);
//	③ 群聊用户消息且有点名(MentionAll 或 message_mentions 非空)→ 异步走
//	   FlushConversation(点名立即投喂,不等静默窗);点名目标经 repo.MessageMentionIDs 读;
//	④ 单聊无点名语义(@ 按普通文本),统一走攒批。
func (d *Dispatcher) NotifyMessage(ctx context.Context, m model.Message) {
	// if m.Role != "user" { return }                          // ①
	// lock; st := stateLocked(m.ConversationID)                // ②
	// st.pending = true; st.lastUserMsgAt = now(); unlock; wakeLoop()
	// if !m.MentionAll { ids, err := d.repo.MessageMentionIDs(ctx, m.ID); if len(ids) == 0 { return } }
	// go func() {                                              // ③
	//     info, err := repo.Conversation(ctx, m.ConversationID)
	//     if err != nil || info.Kind != "group" { return }     // ④ 单聊忽略点名
	//     FlushConversation(ctx, m.ConversationID, m.MentionAll, ids)
	// }()
}

// FlushConversation:立即投喂(手动触发轮次 POST /groups/{id}/rounds 与点名外部入口)。语义:
//
//	① mentionAll=true 或 mentions 为空 = 点名全员(手动触发轮次语义,绕过静默窗直接整批);
//	② 点名集合经 MentionReaderIDs 与成员求交:空交集 → 忽略(消息仍作攒批内容);
//	③ 群冷却未到:点名目标入 targetQueue 等待引擎循环放行(不丢批,放行后只投队列);
//	④ 冷却已过/无冷却:立即 flushOnce(targets)。
func (d *Dispatcher) FlushConversation(ctx context.Context, conversationID uint, mentionAll bool, mentions []uint) error {
	// info, err := repo.Conversation(ctx, conversationID); if err != nil { return err }
	// targets := info.MemberIDs                              // ①
	// if !mentionAll {
	//     t, everyone := MentionReaderIDs(info.MemberIDs, false, mentions)  // ②
	//     if len(t) == 0 && !everyone { return nil }
	//     targets = t
	// }
	// lock; st := stateLocked(conversationID); st.kind = info.Kind
	// if st.running { unlock; return nil }                   // 防重入:并入下一轮攒批
	// if 群 && 距 st.lastFlushAt < cooldown {
	//     if !mentionAll { st.targetQueue = appendUnique(st.targetQueue, targets...) } // ③
	//     unlock; wakeLoop(); return nil
	// }
	// unlock; flushOnce(st, info, targets)                   // ④
	return nil // TODO(实现):见函数注释 ①~④
}

// engineLoop:事件循环主协程。
// 语义:
//
//	① select:stop → 退出;wake / engineTick → scanAndFlush;
//	② scanAndFlush:遍历 states,跳过 running 与"无 pending 且无 targetQueue"的会话,
//	   逐个 maybeFlush(串行,保证同会话不并发写消息);
//	③ 投喂收尾由 flushOnce 挂 defer 统一 wakeLoop(投喂期间新到消息不丢)。
func (d *Dispatcher) engineLoop() {
	// ticker := time.NewTicker(engineTick); defer ticker.Stop()
	// for { select {
	// case <-d.stop:  return                                    // ①
	// case <-d.wake:  d.scanAndFlush()
	// case <-ticker.C: d.scanAndFlush()
	// } }
}

// scanAndFlush:收集"待投"会话并逐个 maybeFlush(冷却/静默未到由 maybeFlush 判定)。
func (d *Dispatcher) scanAndFlush() {
	// lock; due := []*DispatchState{}
	// for _, st := range states {
	//     if st.running { continue }
	//     if st.pending || len(st.targetQueue) > 0 { due = append(due, st) }
	// }
	// unlock
	// for _, st := range due { d.maybeFlush(st) }               // 串行
}

// maybeFlush:到点判定后投喂;未到点保持待投等下一个 tick。语义:
//
//	① 现读 ConvInfo + CompanionsByIDs(配置热更新即时生效);
//	② 群冷却未到 → return(不丢批,点名队列继续等);
//	③ 点名队列非空 → 无需静默窗,直接投(目标=队列,并清队列);
//	④ 攒批(队列空)→ 静默窗(dueBySilence × convDebounce)或攒批硬顶(dueByBacklog ×
//	   backlogExceeded)任一满足才投;
//	⑤ 目标=队列或全员;清 pending;flushOnce。
func (d *Dispatcher) maybeFlush(st *DispatchState) {
	// info := repo.Conversation(...); comps := repo.CompanionsByIDs(info.MemberIDs)  // ①
	// if 群 && 距 st.lastFlushAt < cooldown { return }                               // ②
	// queued := len(st.targetQueue) > 0
	// if !queued {                                                                   // ④
	//     if !dueBySilence(now()-st.lastUserMsgAt, convDebounce(comps, members)) {
	//         if !backlogExceeded(st, info, comps) { return }
	//     }
	// }
	// targets := info.MemberIDs; if queued { targets = st.targetQueue; st.targetQueue = nil }  // ③⑤
	// flushOnce(st, info, targets)
}

// backlogExceeded:任一成员未读批达到攒批硬顶(条数/字数)——群聊逐成员按自身水位算。
// 语义:for each member { 读水位; MessagesAfter(水位后); 过滤本人发送; 统计 count/chars;
//
//	cfg := DispatchSettingsOf(comp); dueByBacklog(...) 任一满足即 true }。
func (d *Dispatcher) backlogExceeded(ctx context.Context, st *DispatchState, info ConvInfo, comps map[uint]model.Companion) (bool, error) {
	// for _, rid := range info.MemberIDs {
	//     cur,_,_ := repo.Cursor(st.conversationID, rid)
	//     msgs, err := repo.MessagesAfter(ctx, st.conversationID, cur.LastReadMessageID); if err != nil { return false, err }
	//     n, chars := 0, 0
	//     for _, m := range msgs { if m.SenderCompanionID != nil && *m.SenderCompanionID == rid { continue }; n++; chars += RuneCount(m.Content) }
	//     if dueByBacklog(n, chars, cfg(comps[rid]).MaxBatchCount, cfg.MaxBatchChars) { return true, nil }
	// }
	// return false, nil
	return false, nil // TODO(实现)
}

// flushOnce:对 targets 逐一"读未读 → 调用 → 推进水位",单个成员失败不阻断其余。语义:
//
//	① 置 running(防重入)与 lastFlushAt;挂 defer 收尾(running=false + wakeLoop);
//	② OnRoundStart(round_start);
//	③ for _, rid := range targets → replyAndConsume(rid):
//	   读水位 → 读未读批(过滤本人)→ BuildBatchLines → 记忆区 → composeTurnMessages → run();
//	④ targets 内成员缺失(已删)跳过并 LogError。
func (d *Dispatcher) flushOnce(st *DispatchState, info ConvInfo, targets []uint) {
	// lock; if st.running { unlock; return }; st.running=true; st.lastFlushAt=now(); unlock  // ①
	// defer func(){ lock; st.running=false; unlock; wakeLoop() }()
	// ctx, cancel := context.WithTimeout(context.Background(), turnTimeout); defer cancel()
	// comps := repo.CompanionsByIDs(ctx, info.MemberIDs)         // 目标含点名子集
	// OnRoundStart(st.conversationID, targets)                   // ②
	// for _, rid := range targets {                              // ③
	//     comp, ok := comps[rid]; if !ok { LogError(...); continue }   // ④
	//     if err := replyAndConsume(ctx, st, info, comp); err != nil { LogError(...) }
	// }
}

// replyAndConsume:读未读 → 组上下文 → 调用 → 视结果推进水位并排定短期记忆总结。
// 本函数是"消费即归档 + 静默已读不回"的落点。语义:
//
//	① 读水位 + MessagesAfter,过滤本人发送;空批 → 直接返回(点名/手动触发常见);
//	② 组上下文:BuildBatchLines(按 min(MaxBatchChars, batchCharCap) 保尾部)+
//	   ShortMemoryBlock + composeTurnMessages;
//	③ run():err → failStreak[rid]++:
//	   - 未达 maxConsecutiveFails:恢复 pending 等待下次触发重试(天然退避),返回 err;
//	   - 已达上限:清零计数,按"已读不回"兜底(silent=true)+ LogError;
//	④ 空输出归一:!silent 且 TrimSpace(text)=="" → 角色 allowSilent 则 silent=true;
//	   不允许静默 → 返回 errEmptyReply(走③失败分支);
//	⑤ !silent:SaveAssistantMessage(role=assistant, sender=rid) + OnReply(new_message);
//	   静默:OnSilent(诊断/审计,无前端事件);
//	⑥ 消费归档:SaveCursor(水位=through=本批最后一条 id)→ scheduleSummary(kvTTL×0.8)
//	   + OnConsumed。已读消息自此永不重投。
func (d *Dispatcher) replyAndConsume(ctx context.Context, st *DispatchState, info ConvInfo, comp model.Companion) error {
	// cur,_,_ := repo.Cursor(ctx, st.conversationID, comp.ID)            // ①
	// batch, through := 收集未读(msgs 过滤 SenderCompanionID==comp.ID,through=末条 id)
	// if len(batch) == 0 { return nil }
	// cfg := DispatchSettingsOf(comp)
	// body := BuildBatchLines(info.Kind, batch, min(cfg.MaxBatchChars, batchCharCap))  // ②
	// mem,_,_ := repo.Memory(ctx, comp.ID, st.conversationID)
	// msgsOut := composeTurnMessages(info.Kind, ShortMemoryBlock(&mem), body)
	// text, silent, err := d.run(ctx, comp, info.Kind, msgsOut)          // ③ run nil→按失败
	// ...见上方 ③④⑤⑥
	return nil // TODO(实现):见函数注释 ①~⑥
}

// errEmptyReply:角色不允许静默时空输出按失败处理(哨兵,供失败分支判定)。
var errEmptyReply = &errEmptyReplyType{}

type errEmptyReplyType struct{}

func (*errEmptyReplyType) Error() string { return "模型空输出且角色不允许静默" }

// scheduleSummary:消费后排定短期记忆总结(fireAt = now + kvTTL×0.8)。
// 说明:期间新消费会再次调用本函数顺延窗口(厂商 KV 缓存已被刷新);
// 到点执行见 summary.go 的 summaryLoop 与 RunSummaryJob。
func (d *Dispatcher) scheduleSummary(companionID, conversationID uint, kvTTLMinutes int) {
	// if kvTTLMinutes <= 0 { kvTTLMinutes = model.DefaultKvTTLMinutes }
	// lock; summaryDue[summaryKey(conversationID, companionID)] = now().Add(TTL * 8 / 10); unlock
}

// summaryKey:(会话,成员) 短期记忆任务键(分隔符 \x00 防拼接歧义)。
func summaryKey(conversationID, companionID uint) string {
	// return Itoa(conversationID) + "\x00" + Itoa(companionID)
	return "" // TODO(实现)
}

// stateLocked:取(或建)会话状态;需持锁调用。
func (d *Dispatcher) stateLocked(convID uint) *DispatchState {
	// st, ok := d.states[convID]
	// if !ok { st = &DispatchState{conversationID: convID, failStreak: map[uint]int{}}; d.states[convID] = st }
	// return st
	return nil // TODO(实现)
}

// wakeLoop:唤醒事件循环(非阻塞信号,通道满时合并多余唤醒)。
func (d *Dispatcher) wakeLoop() {
	// select { case d.wake <- struct{}{}: default: }
}

// appendUnique:点名队列追加去重(队列长度 ≤ 成员数,O(n²) 可接受)。
func appendUnique(dst []uint, items ...uint) []uint {
	// for _, it := range items { if !contains(dst, it) { dst = append(dst, it) } }
	// return dst
	return nil // TODO(实现)
}

// 注:本文件与 agent.go/chat.go 同为草稿。实现顺序建议:
//   1) 先落 DispatchRepo 的 GORM 实现与 member_cursors 读写(server 层装配 TODO);
//   2) 按函数注释 ①~⑥ 顺序实现 replyAndConsume(核心),再补引擎循环;
//   3) 纯函数(dueBySilence/dueByBacklog/MentionReaderIDs/BuildBatchLines 等)
//      在 tests 子目录起最小可运行样例(仿既有 utils/refs 测试形态)后逐个落地;
//   4) 全部就绪后在 init.NewApp 装配 NewDispatcher(...).Start()。
