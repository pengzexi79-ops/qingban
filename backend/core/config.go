package core

// 进程配置定义与加载。
// 配置来源优先级:flag(暂不引入)→ 环境变量(QINBAN_*)→ 默认值。
// 伪代码草稿:LoadConfig 逻辑以函数体内伪代码注释占位。

// Config:进程运行配置。作用:描述"后端进程如何启动",与业务数据(users 表)无关。
type Config struct {
	// DataDir:数据根目录。为空按平台默认:
	// Windows %LOCALAPPDATA%/qingban;macOS ~/Library/Application Support/qingban;Linux ~/.local/share/qingban
	DataDir string

	// Host:监听地址,默认 127.0.0.1(只允许本机;Wails 壳与浏览器同机)。
	Host string

	// Port:首选监听端口,默认 8080(契约约定);占用自动顺延,最终地址回写 ExternalAPI。
	Port int

	// DBPath:SQLite 路径,派生自 DataDir(默认 {DataDir}/qingban.db)。
	DBPath string

	// LogLevel:zap 日志级别(debug/info/warn/error),默认 info;联调可设 debug 观察 SSE。
	LogLevel string

	// LogToFile:是否写文件日志(默认 true,{DataDir}/logs/app.log)。
	LogToFile bool

	// AllowEmptyToken:调试开关。true 时 X-Local-Token 缺失也放行(浏览器同源调试);
	// 正式打包默认 false。见 server/middleware.go。
	AllowEmptyToken bool
}

// DefaultConfig:返回默认配置(调用点:init.NewApp() 启动第一步)。
func DefaultConfig() *Config {
	return &Config{
		Host:            "127.0.0.1",
		Port:            8080,
		LogLevel:        "info",
		LogToFile:       true,
		AllowEmptyToken: false,
		// DataDir/DBPath:init 阶段按平台解析后回填
	}
}

// LoadConfig:按"env/默认"顺序加载并派生路径(DataDir→DBPath→子目录),结果赋给全局 Cfg。
func LoadConfig() (*Config, error) {
	// cfg := DefaultConfig()                                        // ① 默认值
	// cfg.DataDir  = env("QINBAN_DATA_DIR") or platformDefault()    // ② 环境变量覆盖(含 PORT/LOG_LEVEL/ALLOW_EMPTY_TOKEN)
	// os.MkdirAll(cfg.DataDir, 0o755)                               // ③ 建数据目录(files/、logs/ 子目录同法)
	// cfg.DBPath = filepath.Join(cfg.DataDir, "qingban.db")         // ④ 派生 DB 路径
	// return cfg, nil
	return nil, nil // TODO(实现):见函数注释 ①~④
}

// ListenAddr:计算实际监听地址 "host:port"(HTTP Server 与启动日志使用)。
func (c *Config) ListenAddr() string {
	// return fmt.Sprintf("%s:%d", c.Host, c.Port)
	return ""
}
