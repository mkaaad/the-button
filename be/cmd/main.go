package main

import (
	"button/api"
	"button/dao"
	"button/router"
	"button/service"
)

func main() {
	dao.InitRedis()
	dao.InitSQLite()
	service.InitSMSClient()
	go api.BroadCastMessage()
	service.StoreTime()
	router.SetupRoute()
}
