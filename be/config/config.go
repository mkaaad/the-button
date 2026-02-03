package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	AccessKeyID     string
	AccessKeySecret string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
)

func InitConfig() {
	// 加载环境变量
	godotenv.Load()
	AccessKeyID = os.Getenv("ACCESS_KEY_ID")
	AccessKeySecret = os.Getenv("ACCESS_KEY_SECRET")
	RedisAddr = os.Getenv("REDIS_ADDR")
	RedisPassword = os.Getenv("REDIS_PASSWORD")
	RedisDB, _ = strconv.Atoi(os.Getenv("REDIS_DB"))
}
