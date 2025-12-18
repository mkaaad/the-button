package service

import (
	"button/dao"
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	startTime atomic.Int64
)

const rankKey = "button_leaderboard"
const leaderboardLimit = 100

const countdownTime = 60000

type message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
type Rank struct {
	UserName string `json:"user_name"`
	Rank     int    `json:"rank"`
	Time     int64  `json:"time"`
}
type Leaderboard struct {
	Entries []Rank `json:"entries"`
}
type ButtonPress struct {
	UserName string `json:"user_name"`
}
type Time struct {
	Time int64 `json:"time"`
}

func StoreTime() {
	startTime.Store(time.Now().UnixMilli())
}
func GetLeaderboard(send chan []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rank, err := dao.Rdb.ZRangeWithScores(ctx, rankKey, 0, leaderboardLimit-1).Result()
	if err != nil {
		return err
	}
	leaderboard := Leaderboard{
		Entries: make([]Rank, 0, len(rank)),
	}
	for i, entry := range rank {
		userName, ok := entry.Member.(string)
		if !ok {
			continue
		}
		leaderboard.Entries = append(leaderboard.Entries, Rank{
			UserName: userName,
			Rank:     i + 1,
			Time:     int64(entry.Score),
		})
	}
	data, err := json.Marshal(leaderboard)
	msg := message{
		Type: "leaderboard",
		Data: data,
	}
	data, err = json.Marshal(msg)
	if err != nil {
		return err
	}
	send <- data
	return nil
}
func PressButton(userName string, broadcast chan []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Calculate time since last press
	now := time.Now().UnixMilli()
	prev := startTime.Swap(now)
	score := countdownTime - (now - prev)
	// Send button press message
	b := ButtonPress{
		UserName: userName,
	}
	data, err := json.Marshal(b)
	msg := message{
		Type: "button_press",
		Data: data,
	}
	data, err = json.Marshal(msg)
	if err != nil {
		return err
	}
	broadcast <- data
	// Update leaderboard
	_, err = dao.Rdb.ZAddArgs(ctx, rankKey, redis.ZAddArgs{
		LT: true,
		Members: []redis.Z{{
			Score:  float64(score),
			Member: userName,
		}},
	}).Result()
	if err != nil {
		return err
	}

	return nil
}

func GetTime(send chan []byte) error {
	t := &Time{
		Time: startTime.Load(),
	}
	data, err := json.Marshal(t)
	msg := message{
		Type: "time",
		Data: data,
	}
	data, err = json.Marshal(msg)
	if err != nil {
		return err
	}
	send <- data
	return nil
}
