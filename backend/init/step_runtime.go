package initapp

// 阶段 ⑦:运行时对象(事件总线/幂等表/易失缓存)与调度引擎装配。
// 全局落点:core.Hub、core.Idem、core.Mem、ai.Dispatch(由 AI 包提供装配)。
// 易失缓存语义见 core/cache.go:只放"重启即失"的运行态(活动心跳/冷却/回合现场),
// 与持久层严格分工;core.IdleWindow(默认 2s)为编译期可注入变量,可在此覆盖。

import "qingban/core"

// initRuntime:NewApp 第 7 步。
func initRuntime() error {
	// core.Hub = core.NewSSEHub()          // ① SSE 事件总线(/events 数据源)
	// core.Idem = core.NewIdemStore()      // ② 幂等键登记表
	mem, err := core.NewMemCache(60) // ③ 进程内易失 KV(bigcache;心跳/冷却/回合现场)
	if err != nil {
		return err
	}
	core.Mem = mem
	// core.IdleWindow = 2 * time.Second    // ④ 输入完成判定窗(默认 2s;装配/测试可覆盖)
	// ---- 批次调度引擎装配(server 装配 TODO)----
	// repo := server.NewDispatchRepo(core.DB)      // DispatchRepo 的 GORM 实现
	// hooks := server.DispatchHooks{OnReply/OnSilent/OnConsumed/OnRoundStart/LogError}(见 server/messages.go)
	// ai.Dispatch = ai.NewDispatcher(repo, 装配后的 TurnRunner, hooks); ai.Dispatch.Start()
	//   说明:TurnRunner 在 Eino 装配(agent.go buildRuntime)后注入;此前 run=nil 降级只消费
	// ---- 主动消息(正态随机)任务 ----(见 docs/BACKEND_HANDOFF §主动消息)
	//   scheduler := proactive.New(core.DB, core.Mem, hooks...)  // 阶段实现:每日计划生成/触发/审计
	// ---- 桌面壳能力(core.DesktopService)----
	//   壳 main 构造 wails.NewService(...) 后赋值 core.Desktop;纯 HTTP 形态保持 nil,
	//   主动消息通知等统一走 core.NotifyUser(未接壳自动 no-op,SSE 事件不受影响)。
	return nil // TODO(实现):见上方装配注记
}
