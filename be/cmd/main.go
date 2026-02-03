package main

import (
	"button/api"
	"button/config"
	"button/dao"
	"button/router"
	"button/service"
)

func main() {
	config.InitConfig()
	dao.InitRedis()
	dao.InitSQLite()
	service.InitSMSClient()
	go api.BroadCastMessage()
	service.StoreTime()
	router.SetupRoute()
}
