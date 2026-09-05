package core

// 全局变量与动态运行时(对应骨架注释:此处存放全局变量和动态运行时)。
// 架构说明:青伴当前形态 = 桌面本地应用(Wails 壳 + 本机后端 + 本地 SQLite),
// "本地单用户空间"是既定架构而非过渡态:后端进程只服务本机一个用户。
// 故采用包级全局单例存放运行时对象,由 init 包统一赋值;
// 云端多用户/备份属日后独立阶段,届时再引入账号体系与依赖注入,不预埋实现。
//
// 约定:本文件变量仅供 init 赋值、server/AI 只读,业务层不允许再赋值。
// 依赖说明:gorm.io/gorm 与 github.com/sirupsen/logrus 已入 go.mod。

import (
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Cfg:进程配置(端口、数据目录、日志级别等)。
// 由 init 包在启动第一步加载(默认值见 config.go),全局只读。
var Cfg *Config

// DB:本地 SQLite 主库连接(GORM)。
// 全部业务表的唯一读写入口;由 init.OpenDB() 赋值(开启 WAL + busy_timeout)。
// 类型已随 gorm 依赖落地;driver(gorm.io/driver/sqlite)在 init 实现阶段引入。
var DB *gorm.DB

// Log:进程内结构化日志(logrus)。
// 业务层统一用 core.Log.WithField("traceId",...).Info/Error 记录,避免散落 fmt;
// 由 init 包按 cfg.LogLevel 装配(console + 可选文件输出)。
var Log *logrus.Logger

// Hub:本地实时事件总线(SSE /events 的数据源,见 event.go)。
// 新消息/已读/群聊轮次/settings_changed 等均通过 Hub.Publish 广播给所有已连接前端。
var Hub *SSEHub

// Idem:幂等键登记表(见 idempotency.go)。
// 发送消息、触发轮次、导入等"发送类接口"通过 Idempotency-Key 头去重(同 key 只执行一次)。
var Idem *IdemStore

// LocalToken:本地空间访问令牌(X-Local-Token 请求头校验值)。
// 首次初始化时随机生成并持久化到数据目录 token 文件;浏览器调试模式可留空(见 server/middleware)。
// 作用:防止本机其他进程/页面误读写本地数据空间。
var LocalToken string

// DataDir:数据根目录。结构:
//
//	{DataDir}/qingban.db  SQLite 主库
//	{DataDir}/secret.key  密钥文件(API Key 加密用 AES-256-GCM,见 utils/secret.go)
//	{DataDir}/token.txt   本地访问令牌
//	{DataDir}/files/      附件根目录({DataDir}/files/{fileId},见 model/file.go)
//   {DataDir}/logs/       logrus 日志文件(可选,见 cfg.LogToFile)
var DataDir string

// StartTime:进程启动时刻(UTC)。
// 供 /health serverTime、uptime 与启动日志使用。
var StartTime = time.Now().UTC()

// timeNowUTC:当前 UTC 时间辅助(供包内初始化使用)。
// 所有持久化时间统一 RFC3339 UTC,前端按本地时区展示。
func timeNowUTC() time.Time {
	return time.Now().UTC()
}
