package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	AccessKeyID     string
	AccessKeySecret string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	StartTime       time.Time
	EndTime         time.Time
)

func parseGameTime(raw string, loc *time.Location) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, raw, loc)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported datetime format: %s", raw)
}

func InitConfig() {
	// 加载环境变量
	godotenv.Load()
	accessKeyID, ok1 := os.LookupEnv("ACCESS_KEY_ID")
	accessKeySecret, ok2 := os.LookupEnv("ACCESS_KEY_SECRET")
	redisAddr, ok3 := os.LookupEnv("REDIS_ADDR")
	redisPassword, ok4 := os.LookupEnv("REDIS_PASSWORD")
	redisDB, ok5 := os.LookupEnv("REDIS_DB")
	startTimeRaw, ok6 := os.LookupEnv("GAME_START_TIME")
	endTimeRaw, ok7 := os.LookupEnv("GAME_END_TIME")
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !ok6 || !ok7 {
		panic("failed to get required config from env")
	}

	AccessKeyID, AccessKeySecret, RedisAddr, RedisPassword = accessKeyID, accessKeySecret, redisAddr, redisPassword
	db, err := strconv.Atoi(redisDB)
	if err != nil {
		panic(fmt.Sprintf("invalid REDIS_DB: %v", err))
	}
	RedisDB = db

	timezone := os.Getenv("GAME_TIMEZONE")
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		panic(fmt.Sprintf("failed to load GAME_TIMEZONE %q: %v", timezone, err))
	}

	startAt, err := parseGameTime(startTimeRaw, loc)
	if err != nil {
		panic(fmt.Sprintf("invalid GAME_START_TIME: %v", err))
	}
	endAt, err := parseGameTime(endTimeRaw, loc)
	if err != nil {
		panic(fmt.Sprintf("invalid GAME_END_TIME: %v", err))
	}
	if !startAt.Before(endAt) {
		panic("GAME_START_TIME must be earlier than GAME_END_TIME")
	}

	StartTime = startAt
	EndTime = endAt
}
