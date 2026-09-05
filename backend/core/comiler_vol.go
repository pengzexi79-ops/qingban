package core

// 本文件存放"编译期注入的变量"(对应骨架注释:comiler_vol = 编译期注入变量,
// 例如对外服务地址——桌面形态下即本地后端的监听地址)。
// 生产构建时通过 ldflags -X 注入;本地 go run / 开发模式使用默认值。
// 例:go build -ldflags "-X qingban/core.Version=0.2.0-phase1 -X qingban/core.Commit=abc1234"

// 编译注入变量:版本号。
// 作用:/health 的 apiVersion 与启动日志展示;保持与 openapi.phase1.yaml 的
// version 字段(0.2.0-phase1)及前端展示一致。
var Version = "0.2.0-phase1"

// 编译注入变量:git commit(短哈希)。
// 作用:仅日志与排障,标识当前二进制对应的源码版本,便于用户上报 bug 时定位。
var Commit = "dev"

// 编译注入变量:构建时间(UTC RFC3339)。
// 作用:仅日志展示(启动横幅"built at ..."),不做业务判断。
var BuildTime = ""

// 编译注入变量:实际对外暴露的 API 地址(形式 http://127.0.0.1:PORT)。
// 作用:Wails 壳注入前端 window.__QINBAN_API__ 的数据源;网页调试模式留空走同源。
// 说明:端口顺延后由 init 在运行时回写此变量,使壳与日志拿到最终地址。
var ExternalAPI = ""

//服务器地址
var SeverHost = ""

//前端地址
var AppHost = ""

var Mode = ""
