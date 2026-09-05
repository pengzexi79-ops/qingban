package initapp

// 阶段 ⑧:路由引擎装配。
// 中间件分层/端点映射全部在 server/router.go(RegisterRoutes),本文件只负责"建引擎+注册"。

import (
	"github.com/gin-gonic/gin"

	"qingban/server"
)

// buildRouter:NewApp 第 8 步。返回 gin 引擎(注册完全部路由),交由 App.Run 监听。
func buildRouter() (*gin.Engine, error) {
	engine := gin.New() // 自带 Recovery 中间件
	server.RegisterRoutes(engine)
	return engine, nil
}
