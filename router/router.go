package router

import (
	"button/connx"
	"button/service"
	"errors"
)

func HandleMessage(conn *connx.SafeConn, msg []byte, userID int64, message chan []byte) error {
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
