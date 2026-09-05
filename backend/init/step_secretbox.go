package initapp

// 阶段 ③:装载/生成密钥盒(APIKey 加密)。
// 说明:box 不设全局——由 api_configs 服务经依赖注入闭包使用(避免全局明文句柄扩散),
// 本阶段只做"装载成功"的前置校验;box 透传到路由/运行时装配的闭包由实现阶段接线。

import "qingban/core"

// initSecretBox:NewApp 第 3 步。
func initSecretBox(cfg *core.Config) error {
	_ = cfg
	// box, err := utils.LoadOrCreateBox(cfg.DataDir)    // ① 不存在则生成并写盘(0600)
	// if err != nil { return err }                      // ② 失败→fatal(密钥盒是安全地基)
	// // ③ 闭包注入:server/api_configs 服务(读/写时解密),此处仅透传说明
	return nil // TODO(实现):见函数注释 ①~③
}
