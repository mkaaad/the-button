package main

import (
	"button/api"
	"button/dao"
	"log"
	"net/http"
)

func main() {
	dao.InitRedis()
	go api.BroadCastMessage()
	http.HandleFunc("/ws", api.WebSocketHandler)
	log.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
