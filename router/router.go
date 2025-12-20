package router

import (
	"button/api"

	"github.com/gin-gonic/gin"
)

func SetupRoute() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/ws", api.WebSocketHandler)
	r.Run(":8080")
}
