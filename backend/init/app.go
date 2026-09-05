package initapp

// 启动编排入口(目录 init/,包名 initapp,规避关键字 init)。
// 职责:把各"初始化阶段函数"(同目录 step_*.go)按依赖顺序组装成 NewApp,
// 并为进程提供 Run(监听/优雅退出)。每个阶段的伪代码留在对应文件内,
// 本文件只表达"顺序与装配",不再堆细节。
// 骨架目录职责(init/enter.go):引入 core 并初始化,返回运行对象。

// App:初始化完成后的运行对象(main 持有并阻塞运行)。
// 字段说明:
//   - engine:gin 路由引擎(RegisterRoutes 注册完 server/router.go 全部路由);
//   - srv:HTTP 服务(实际监听地址 = core.ExternalAPI),由 Run 阶段创建。
type App struct {
	engine any // 类型计划:*gin.Engine
	srv    any // 类型计划:*http.Server
}

// NewApp:完整的初始化编排(唯一入口,main.go 调用)。
// 阶段顺序与文件/函数对应:
//
//	① 配置:loadConfig          (step_config.go)
//	② 日志:initLogging         (step_log.go)
//	③ 密钥盒:initSecretBox     (step_secretbox.go)
//	④ 数据库:openDB→migrateDB  (step_db.go)
//	⑤ 引导种子:preSeedIfFirstRun(step_bootstrap.go)
//	⑥ 本地令牌:initLocalToken  (step_token.go)
//	⑦ 运行时对象:initRuntime   (step_runtime.go,含易失缓存与调度引擎装配)
//	⑧ 路由:buildRouter        (step_router.go)
//
// P0 验收:"后端进程能起、前端能连、能收事件"由前 8 步保证(缺实现时逐段点亮)。
func NewApp() (*App, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	if err := initLogging(cfg); err != nil {
		return nil, err
	}
	if err := initSecretBox(cfg); err != nil {
		return nil, err
	}
	if err := openDB(cfg); err != nil {
		return nil, err
	}
	if err := migrateDB(); err != nil {
		return nil, err
	}
	if err := preSeedIfFirstRun(); err != nil {
		return nil, err
	}
	if err := initLocalToken(cfg); err != nil {
		return nil, err
	}
	if err := initRuntime(); err != nil {
		return nil, err
	}
	engine, err := buildRouter()
	if err != nil {
		return nil, err
	}
	// 装配完成:engine/srv 交由 Run 使用;主动消息/心跳等阶段任务按需在 initRuntime 内追加。
	return &App{engine: engine}, nil
}

// Run:启动 HTTP 监听并阻塞,直至收到退出信号。
// 逻辑:
//
//	① 端口顺延:从 cfg.Port 开始 net.Listen 探测,占用则 +1(上限 +20,日志提示实际端口)
//	② core.ExternalAPI = fmt.Sprintf("http://127.0.0.1:%d", 实际端口) ← Wails 注入前端的地址
//	③ srv := &http.Server{Handler: engine};go srv.Serve(listener)
//	④ 等待退出信号(SIGINT/SIGTERM);收到后执行:
//	   - core.Hub.Close()(先断 SSE,让前端转轮询/重连)
//	   - srv.Shutdown(ctx 5s)(等消息落库与模型调用收尾)
//	   - 关闭 SQLite(GORM sqlDB.Close())
//	   - 日志"优雅退出完成"
func (a *App) Run() error {
	// engine := a.engine.(*gin.Engine)          // 断言真实引擎(见 step_router.go)
	// ...见上方 ①~④
	return nil // TODO(实现):见函数注释 ①~④
}
