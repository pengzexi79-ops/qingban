package initapp

// 阶段 ⑥:本地访问令牌。
// 全局落点:core.LocalToken。首次随机生成并持久化到 {DataDir}/token.txt(0600);
// 浏览器调试模式可留空(见 cfg.AllowEmptyToken 与 server/middleware.go)。

import "qingban/core"

// initLocalToken:NewApp 第 6 步。
func initLocalToken(cfg *core.Config) error {
	_ = cfg
	// path := filepath.Join(cfg.DataDir, "token.txt")            // ① 读取
	// if b, err := os.ReadFile(path); err == nil && len(b) > 0 {
	//     core.LocalToken = strings.TrimSpace(string(b)); return nil
	// }
	// tok := hex.EncodeToString(randBytes(32))                   // ② 不存在 → 生成 32 字节随机 hex
	// os.WriteFile(path, []byte(tok), 0o600)
	// core.LocalToken = tok
	return nil // TODO(实现):见函数注释 ①~②
}
