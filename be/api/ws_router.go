package api

import (
	"button/service"
	"errors"
)

const (
	GET_TIME       = "1"
	GET_LEADEROARD = "2"
	PRESS_BUTTON   = "3"
)

func handleMessage(msg []byte, sessionID string, send chan []byte, broadcast chan []byte) error {
	switch string(msg) {
	case GET_LEADEROARD:
		return service.GetLeaderboard(send)
	case PRESS_BUTTON:
		username, ok := service.IsLogin(sessionID, send)
		if !ok {
			return nil
		}
		if !service.IsWithinTime(send) {
			return nil
		}
		if service.IsLocked(username, send) {
			return nil
		}
		return service.PressButton(username, broadcast)
	case GET_TIME:
		return service.GetTime(send)
	default:
		return errors.New("parse msg error")
	}
}
