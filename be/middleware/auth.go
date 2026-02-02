package middleware

import (
	"button/dao"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func VerifyCookie(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"info": "未登录",
		})
		c.Abort()
		return
	}
	username, err := dao.Rdb.Get(context.Background(), sessionID).Result()
	if err == redis.Nil {
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		c.JSON(http.StatusUnauthorized, gin.H{
			"info": "登录过期",
		})
		c.Abort()
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"info": "redis服务错误",
		})
		c.Abort()
		return
	}
	c.Set("username", username)
	c.Next()
}
