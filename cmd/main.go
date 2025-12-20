package main

import (
	"button/api"
	"button/dao"
	"button/router"
	"button/service"
)

func main() {
	dao.InitRedis()
	go api.BroadCastMessage()
	service.StoreTime()
	router.SetupRoute()
}
