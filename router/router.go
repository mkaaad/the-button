package router

import (
	"button/service"
	"errors"
)

const (
	GET_TIME       = "1"
	GET_LEADEROARD = "2"
	PRESS_BUTTON   = "3"
)

func HandleMessage(msg []byte, userName string, send chan []byte, broadcast chan []byte) error {
	switch string(msg) {
	case GET_LEADEROARD:
		return service.GetLeaderboard(send)
	case PRESS_BUTTON:
		return service.PressButton(userName, broadcast)
	case GET_TIME:
		return service.GetTime(send)
	default:
		return errors.New("parse msg error")
	}
}
