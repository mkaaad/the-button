package router

import (
	"button/api"
	"button/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoute() {
	r := gin.Default()
	r.Use(middleware.Cors())
	r.GET("/ws", api.WebSocketHandler)
	r.POST("/sms/code", api.SendVerifyCode)
	r.POST("/sms/verify", api.VerifyCode)
	r.GET("/sms/captcha", api.GetCaptcha)
	r.Run(":8080")
}
