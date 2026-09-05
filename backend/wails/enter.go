package wails

// 桌面壳接入层(提前抽象;真正接入壳时启用)。
// 依赖方向(避免反向依赖):
//   core   → 声明 core.DesktopService 接口 + 默认 no-op(见 core/desktop.go);
//   server/AI → 只调用 core.NotifyUser / core.DesktopService,不 import 本包;
//   本包(wails) → 实现 core.DesktopService,并提供给 Wails 壳的绑定方法(见 service.go)。
//
// 接入时(壳阶段):
//   - 由壳 main(参考 docs/tem/helloWails)构造 wails.NewService(实际API地址, 令牌, 壳实现),
//     赋值 core.DesktopService;HTTP 后端在同一进程 goroutine 运行(init.NewApp().Run() 由后端 init 提供);
//   - 前端通过绑定方法拿 API 地址/令牌/能力,业务数据仍走 HTTP + SSE。
//
// 浏览器开发形态:本包不参与编译(壳未接入),core.DesktopService=nil,通知等自动 no-op。
