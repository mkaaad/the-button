package main

import (
	"button/api"
	"button/dao"

	"github.com/gin-gonic/gin"
)

func main() {
	dao.InitRedis()
	go api.BroadCastMessage()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/ws", api.WebSocketHandler)
	r.Run("127.0.0.1:8080")
}
