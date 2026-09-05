package core

// 桌面壳能力抽象(Wails 壳接入实现并注入)。
// 依赖方向(提前定好,避免反向依赖):
//   core   —— 只声明本接口 + 默认空实现(未接入壳时 no-op);
//   server/AI —— 只面向 core.DesktopService / core.NotifyUser 调用,不 import wails;
//   wails  —— 实现 core.DesktopService 并提供给 Wails 壳的绑定方法(见 backend/wails/service.go)。
// 注入点:Wails 壳的 main 里构造 wails.Service 后赋值 core.DesktopService;
// 纯 HTTP/浏览器形态(壳未接入)保持 nil → 通知等能力静默跳过。

// DesktopService:桌面壳提供的系统能力。
type DesktopService interface {
	// Notify:系统通知(主动消息到达/备份完成等;浏览器形态走页面内事件,不依赖本方法)。
	Notify(title, body string) error
	// OpenFolder:在系统文件管理器中打开目录(如"打开数据目录/附件目录")。
	OpenFolder(path string) error
	// SetLaunchAtLogin:设置开机自启(托盘/设置页开关)。
	SetLaunchAtLogin(enabled bool) error
}

// Desktop:当前已注入的桌面壳实现(nil=未接入,见 NotifyUser 语义)。
// 约定:仅 init/壳 main 可赋值;server/AI 只读。
var Desktop DesktopService

// NotifyUser:发系统通知的统一入口。壳未接入(Desktop==nil)时静默成功——
// 主动消息等调用方无需关心形态差异,页面内事件(SSE)始终照常推送。
func NotifyUser(title, body string) error {
	if Desktop == nil {
		return nil
	}
	return Desktop.Notify(title, body)
}

// OpenFolderFor:打开指定目录;壳未接入时静默成功。
func OpenFolderFor(path string) error {
	if Desktop == nil {
		return nil
	}
	return Desktop.OpenFolder(path)
}
