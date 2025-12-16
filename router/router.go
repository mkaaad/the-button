package router

import (
	"button/service"
	"errors"

	"github.com/gorilla/websocket"
)

func HandleMessage(conn *websocket.Conn, msg []byte, userID int64, message chan []byte) error {
	switch string(msg) {
	case "leaderboard":
		return service.GetLeaderboard(conn)
	case "button":
		return service.PressButton(userID, message)
	case "time":
		return service.GetTime(conn)
	default:
		return errors.New("parse msg error")
	}
}
