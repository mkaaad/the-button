package main

import (
	"button/api"
	"button/dao"

	"github.com/gin-gonic/gin"
)

func main() {
	dao.InitRedis()
	go api.BroadCastMessage()
	r := gin.Default()
	r.GET("/", api.WebSocketHandler)
	r.Run()
}
