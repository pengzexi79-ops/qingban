package wails

// 桌面壳接入层(提前抽象,壳阶段实现/注入)。
// 职责定位:Wails 壳只做"桥 + 系统能力",业务仍走 HTTP + SSE(backend/init + server 不变):
//   1) 把实际 API 地址与本机令牌交给前端(注入 window.__QINBAN_API__ / X-Local-Token);
//   2) 实现 core.DesktopService(通知/打开目录/开机自启),赋给 core.DesktopService 供 server/AI 无感调用;
//   3) 壳内操作(退出)经 Service 暴露给前端 window.go.wails.Service.* 绑定。
// 绑定约束(Wails v2):方法必须导出、参数/返回值 JSON 可序列化(不用 time.Time/gorm 类型);
// 本文件不 import Wails SDK——真正接入时由壳工程(参考 docs/tem/helloWails)的 main 构造
// Service 并 wails.Run,HTTP 后端在同进程 goroutine 运行(init.NewApp().Run())。

import "qingban/core"

// Service:壳绑定服务。字段由壳 main 构造时注入,构造后只读。
type Service struct {
	// apiBaseURL:实际后端地址(端口顺延后),形如 http://127.0.0.1:8080。
	apiBaseURL string
	// localToken:本地访问令牌(前端所有请求带 X-Local-Token)。
	localToken string
	// desktop:系统能力实现(壳内实现;未接壳形态可为 nil)。
	desktop core.DesktopService
	// quitFn:退出应用回调(壳 main 经 SetQuitHandler 注入)。
	quitFn func()
}

// NewService:构造绑定服务(壳 main 调用;desktop 传壳实现)。
func NewService(apiBaseURL, localToken string, desktop core.DesktopService) *Service {
	return &Service{apiBaseURL: apiBaseURL, localToken: localToken, desktop: desktop}
}

// ---- 前端接入信息(绑定方法,JSON 安全)----

// APIBaseURL:后端 API 基址(前端据此注入 window.__QINBAN_API__)。
func (s *Service) APIBaseURL() string { return s.apiBaseURL }

// LocalToken:本地访问令牌(前端请求头用;AllowEmptyToken 调试态可为空)。
func (s *Service) LocalToken() string { return s.localToken }

// AppInfo:应用信息(版本/数据目录等,设置页展示用)。
func (s *Service) AppInfo() map[string]string {
	return map[string]string{
		"apiBaseURL": s.apiBaseURL,
		"version":    "dev", // 壳编译期注入(参考 core/comiler_vol.go 约定)
	}
}

// ---- core.DesktopService 实现(转发壳实现;未接壳 no-op)----

// Notify:系统通知。
func (s *Service) Notify(title, body string) error {
	if s.desktop == nil {
		return nil
	}
	return s.desktop.Notify(title, body)
}

// OpenFolder:打开系统文件管理器到指定目录。
func (s *Service) OpenFolder(path string) error {
	if s.desktop == nil {
		return nil
	}
	return s.desktop.OpenFolder(path)
}

// SetLaunchAtLogin:设置开机自启。
func (s *Service) SetLaunchAtLogin(enabled bool) error {
	if s.desktop == nil {
		return nil
	}
	return s.desktop.SetLaunchAtLogin(enabled)
}

// ---- 壳内能力(不进 core.DesktopService;由壳 main 注入回调,Service 不依赖壳实现)----

// SetQuitHandler:注册"退出应用"回调(托盘"退出"/设置页 Quit() 时触发)。
func (s *Service) SetQuitHandler(fn func()) {
	s.quitFn = fn
}

// Quit:请求退出应用壳(前端绑定方法)。
func (s *Service) Quit() {
	if s.quitFn != nil {
		s.quitFn()
	}
}
