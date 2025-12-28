package router

import (
	"button/api"
	"button/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoute() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/ws", middleware.VerifyCookie, api.WebSocketHandler)
	r.POST("/sms/code", api.SendVerifyCode)
	r.POST("/sms/verify", api.VerifyCode)
	r.Run(":8080")
}
