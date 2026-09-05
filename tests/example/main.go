package example

import "github.com/gin-gonic/gin"

func main() {
	//例如这是Gin的测试
	//最小化的可用基本功能实例
	r := gin.Default()
	r.POST("/jobs")
	_ = r.Run("127.0.0.1:8080")
}
func exampleHandle(c gin.Context) {
	c.JSON(200, "success")
}
