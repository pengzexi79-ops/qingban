package initapp

// 阶段 ①:加载进程配置(env/默认),派生路径并建数据目录。
// 全局落点:core.Cfg、core.DataDir。
// 伪代码草稿:核心逻辑在 core.LoadConfig(按注释①~④补齐);本函数做编排与目录准备,
// 并在核心实现前以 core.DefaultConfig 兜底保证编排可空跑。

import (
	"os"
	"path/filepath"

	"qingban/core"
)

// loadConfig:NewApp 第 1 步。
func loadConfig() (*core.Config, error) {
	cfg, err := core.LoadConfig() // ① 默认值 + env 覆盖(含 QINBAN_DATA_DIR/LOG_LEVEL 等)
	if err != nil {
		return nil, err
	}
	if cfg == nil { // 过渡兜底:core.LoadConfig 未实现前的空跑(实现落地后删除本分支)
		cfg = core.DefaultConfig()
	}
	core.Cfg = cfg
	core.DataDir = cfg.DataDir
	if cfg.DataDir == "" {
		return cfg, nil // ② 数据目录未解析出来前不建目录
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil { // ③ 建数据根目录
		return nil, err
	}
	// ④ 子目录:files/ 与 logs/(见 core/globals.go DataDir 结构)
	for _, sub := range []string{"files", "logs"} {
		if err := os.MkdirAll(filepath.Join(cfg.DataDir, sub), 0o755); err != nil {
			return nil, err
		}
	}
	cfg.DBPath = filepath.Join(cfg.DataDir, "qingban.db") // ⑤ 派生 DB 路径
	return cfg, nil
}
