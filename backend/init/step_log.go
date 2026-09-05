package initapp

// 阶段 ②:装配进程日志(logrus)。
// 全局落点:core.Log。console 必出;cfg.LogToFile 时追加 {DataDir}/logs/app.log。

import "qingban/core"

// initLogging:NewApp 第 2 步。
func initLogging(cfg *core.Config) error {
	_ = cfg
	// l := logrus.New()                                            // ① 按 cfg.LogLevel 建 logger(debug/info/warn/error)
	// l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})   // ② 格式:文本 + 时间戳
	// if cfg.LogToFile {                                            // ③ 可选文件输出(建议 caller 字段)
	//     f, _ := os.OpenFile(filepath.Join(cfg.DataDir, "logs", "app.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	//     l.SetOutput(io.MultiWriter(os.Stdout, f))
	// }
	// core.Log = l
	// l.Info("启动横幅:", BuildVersion/Commit/BuildTime/DataDir)
	return nil // TODO(实现):见函数注释 ①~③
}
