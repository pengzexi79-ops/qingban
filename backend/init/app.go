package initapp

// 启动编排(目录 init/,包名 initapp,规避关键字 init)。
// 形态:桌面本地后端进程(Wails 壳同机启动,监听 127.0.0.1),与服务器部署无关;
// 云端备份/多设备同步属日后独立阶段,本编排不承载任何上云职责。
// 骨架目录职责(init/enter.go):引入 core 并进行初始化,返回初始化后的 core 变量。
// 本文件为伪代码草稿:每一步的"预备编码位置"以编号注释说明将执行的逻辑;
// 计划依赖(阶段实现时加入 go.mod):gin、gorm.io/gorm、gorm.io/driver/sqlite、
// github.com/sirupsen/logrus(已入 go.mod)、github.com/google/uuid(如需)。

// App:初始化完成后的运行对象(main 持有并阻塞运行)。
// 字段作用:
type App struct {
	// engine:Gin 路由引擎(注册完 server/router.go 全部路由)。
	engine any // 类型计划:*gin.Engine
	// srv:HTTP 服务(实际监听地址 = core.ExternalAPI)。
	srv any // 类型计划:*http.Server
}

// NewApp:完整的初始化编排(唯一入口,main.go 调用)。
// 返回的 App 已具备全部运行时,可直接 Run。
// 阶段 P0 验收:"后端进程能起、前端能连、能收事件"由本函数的前 8 步保证。
func NewApp() (*App, error) {
	// ========== 1. 加载配置 ==========
	// 逻辑:cfg, err := core.LoadConfig();失败→返回错误
	// 赋值:core.Cfg = cfg;core.DataDir = cfg.DataDir
	// 准备目录:MkdirAll(DataDir/files)、MkdirAll(DataDir/logs)

	// ========== 2. 初始化日志(logrus) ==========
	// 逻辑:按 cfg.LogLevel 建 logrus.Logger(console 必出;LogToFile=true 时追加
	//   {DataDir}/logs/app.log 输出,建议带 caller 与文本/JSON 格式按形态选择);core.Log 赋值
	// 日志:启动横幅(BuildVersion/Commit/BuildTime/DataDir)

	// ========== 3. 装载/生成密钥盒(API Key 加密) ==========
	// 逻辑:box, err := utils.LoadOrCreateBox(DataDir);err→fatal
	// 说明:box 不设全局,由 api_profiles 服务经依赖注入闭包使用(避免全局明文句柄扩散)

	// ========== 4. 打开 SQLite 并迁移建表 ==========
	// 逻辑:gorm.Open(sqlite.Open(DBPath)?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on)
	// core.DB 赋值;执行 AutoMigrate 全部实体(依赖顺序,见 docs/model_design.md):
	//   model_configs/files/users/companions/groups/group_members/conversations/messages/
	//   message_files/message_mentions/rounds/round_speakers/memories/usage_records/
	//   config_kvs/member_cursors/chat_short_memories
	// 注意:_foreign_keys=on 必须开启,否则外键级联约束不生效。
	// 补充(代码位置,非 GORM 自动迁移,需原生 SQL):
	//   ① DROP TABLE IF EXISTS api_profiles;       ← 旧实体已删除,不留数据
	//   ② 建 FTS5 虚拟表(消息/记忆全文检索):
	//      CREATE VIRTUAL TABLE IF NOT EXISTS fts_messages USING fts5(content, tokenize='unicode61')
	//      (阶段一先支持英文分词,中文可后续换 jieba 分词器或 LIKE 兜底)
	//      + 消息插入/删除的触发器同步维护(或用应用层双写)
	//   ③ sqlite-vec 向量表(可选扩展):
	//      memory_vectors(memory_id INTEGER PRIMARY KEY, companion_id INTEGER, vector BLOB, model TEXT, updated_at)
	//      扩展不存在时自动降级关键词检索(见 AI/recall.go)
	//   ④ 补索引:usage_records(created_at)、memories 按 date 排序等(见各模型注释)
	// 启动自检:db.Ping;失败→degraded 状态继续(health 返回 dbOk=false)

	// ========== 5. 引导状态与种子数据 ==========
	// 逻辑:读 kv k:bootstrap:done
	//   未初始化:创建默认用户(昵称"我",默认设置:autoMessages=true 等)与
	//     默认 Ollama API 配置(provider=ollama, baseUrl=http://localhost:11434,
	//     protocol=openai-compatible, 无密钥)→ 写 kv 默认配置 id;但先不写 done 标记
	//     (等前端 POST /bootstrap/init 选 mode 后再标记,保证 firstRun 语义正确)
	//   已初始化:跳过

	// ========== 6. 本地令牌 ==========
	// 逻辑:读取 {DataDir}/token.txt;不存在→生成 32 字节随机 hex 写盘(0600)
	// 赋值:core.LocalToken

	// ========== 7. 运行时对象 ==========
	// 逻辑:core.Hub = core.NewSSEHub();core.Idem = core.NewIdemStore()
	// 追加(批次调度引擎,见 AI/dispatch.go 与 docs/batch_dispatch_design.md):
	//   repo := server.NewDispatchRepo(core.DB)             // DispatchRepo 的 GORM 实现(server 装配 TODO)
	//   ai.Dispatch = ai.NewDispatcher(repo, 装配后的 TurnRunner, server 组装的 DispatchHooks)
	//   说明:TurnRunner 在 Eino 装配(agent.go buildRuntime)后注入;此前 run=nil 降级只消费

	// ========== 8. 路由注册 ==========
	// 逻辑:engine := gin.New()(自带 Recovery);加载 server.RegisterRoutes(engine)
	//   说明:中间件(日志/令牌/幂等)在 RegisterRoutes 内按端点分组挂载
	// 返回 (*App{engine, srv}, nil)
	return &App{}, nil // TODO(实现):见以上步骤 1~8
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
	// TODO(实现):见函数注释 ①~④
	return nil
}
