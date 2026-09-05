package initapp

// 阶段 ⑤:引导状态与种子数据(首次运行预置,不写"完成"标记)。
// 完成标记由 server 层 POST /bootstrap/init 在用户选择模式后写入(保证 firstRun 语义)。
// 伪代码草稿:逻辑以注释占位(导入内核见 server/data.go importPayloadCore)。

// preSeedIfFirstRun:NewApp 第 5 步。本地空间尚未初始化时预置:
//   - 单行用户(昵称"我",默认设置:autoMessages 等按默认);
//   - 默认 API 配置(本地 Ollama:BaseURI http://localhost:11434/v1,无密钥),并把其 id 写入
//     config_kvs 的 KVDefaultAPIConfigID(角色未绑定配置时回落);仍不写 KVBootstrapDone。
func preSeedIfFirstRun() error {
	// if kvGet(model.KVBootstrapDone) == "1" { return nil }          // 已初始化 → 跳过
	// tx {
	//     db.Create(&model.User{Nickname: "我", Settings: defaultSettings})   // ① 单行用户(id 自增=1)
	//     seed := model.APIConfig{Name: "local-ollama", DisplayName: "本地模型(Ollama)",
	//         BaseURI: "http://localhost:11434/v1", APIType: "ollama",         // ② 种子配置
	//         TextCompletion: true, Temperature: 0.7}
	//     db.Create(&seed)
	//     kvSet(model.KVDefaultAPIConfigID, strconv(seed.ID))
	// }
	// // ③ 不写 done 标记:等前端 POST /bootstrap/init 选模式后再写(empty / import-demo)
	return nil // TODO(实现):见函数注释;导入模式内核在 server/data.go importPayloadCore
}
