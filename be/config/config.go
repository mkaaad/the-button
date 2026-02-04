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
	accessKeyID, ok1 := os.LookupEnv("ACCESS_KEY_ID")
	accessKeySecret, ok2 := os.LookupEnv("ACCESS_KEY_SECRET")
	redisAddr, ok3 := os.LookupEnv("REDIS_ADDR")
	redisPassword, ok4 := os.LookupEnv("REDIS_PASSWORD")
	redisDB, ok5 := os.LookupEnv("REDIS_DB")
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		panic("failed to get info form env")
	}
	AccessKeyID, AccessKeySecret, RedisAddr, RedisPassword = accessKeyID, accessKeySecret, redisAddr, redisPassword
	RedisDB, _ = strconv.Atoi(redisDB)
}
