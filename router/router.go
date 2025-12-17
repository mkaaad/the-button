package router

import (
	"button/service"
	"errors"
)

func HandleMessage(msg []byte, userID int64, send chan []byte, broadcast chan []byte) error {
	switch string(msg) {
	case "leaderboard":
		return service.GetLeaderboard(send)
	case "button":
		return service.PressButton(userID, broadcast)
	case "time":
		return service.GetTime(send)
	default:
		return errors.New("parse msg error")
	}
}
