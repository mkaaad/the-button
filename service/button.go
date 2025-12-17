package service

import (
	"button/connx"
	"button/dao"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var (
	startTime = time.Now().UnixMilli()
	mu        sync.RWMutex
)

const rankKey = "button_leaderboard"

type message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
type Rank struct {
	UserID int64 `json:"user_id"`
	Rank   int   `json:"rank"`
	Time   int64 `json:"time"`
}
type Leaderboard struct {
	Entries []Rank `json:"entries"`
}
type ButtonPress struct {
	UserID    int64 `json:"user_id"`
	Timestamp int64 `json:"timestamp"`
}

func GetLeaderboard(conn *connx.SafeConn) error {
	rank, err := dao.Rdb.ZRevRangeWithScores(context.Background(), rankKey, 0, -1).Result()
	if err != nil {
		return err
	}
	leaderboard := Leaderboard{
		Entries: make([]Rank, 0, len(rank)),
	}
	for i, entry := range rank {
		userID, ok := entry.Member.(int64)
		if !ok {
			continue
		}
		leaderboard.Entries = append(leaderboard.Entries, Rank{
			UserID: userID,
			Rank:   i + 1,
			Time:   int64(entry.Score),
		})
	}
	msg := message{
		Type: "leaderboard",
		Data: leaderboard,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)

}
func PressButton(userID int64, messageChan chan []byte) error {
	// Calculate time since last press
	now := time.Now().UnixMilli()
	mu.Lock()
	score := now - startTime
	startTime = now
	mu.Unlock()
	// Update leaderboard
	luaScript := redis.NewScript(`
	local rankKey = KEYS[1]
	local userID = ARGV[1]
	local score = tonumber(ARGV[2])
	local currentScore = redis.call("ZSCORE", rankKey, userID)
	if currentScore == false or score < tonumber(currentScore) then
		redis.call("ZADD", rankKey, score, userID)
	end
	return true
	`)
	_, err := luaScript.Run(context.Background(), dao.Rdb, []string{rankKey}, userID, score).Result()
	if err != nil {
		return err
	}
	// Send button press message
	b := ButtonPress{
		UserID:    userID,
		Timestamp: now,
	}
	msg := message{
		Type: "button_press",
		Data: b,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	messageChan <- data
	return nil
}

func GetTime(conn *connx.SafeConn) error {
	mu.RLock()
	msg := message{
		Type: "time",
		Data: startTime,
	}
	mu.RUnlock()
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}
