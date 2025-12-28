package api

import (
	"button/config"
	"button/model"
	"button/service"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

var mobileRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

type verify struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

func SendVerifyCode(c *gin.Context) {
	var v verify
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "格式错误",
		})
		return
	}
	if !mobileRegexp.MatchString(v.PhoneNumber) {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "手机号格式错误",
		})
		return
	}
	err := service.SendVerifyCode(v.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"info": "已发送短信验证码",
	})
}
func VerifyCode(c *gin.Context) {
	var u model.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "格式错误",
		})
		return
	}
	if len(u.VerifyCode) != int(config.CodeLenth) {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "验证码格式错误",
		})
		return
	}
	err := service.VerifyCode(u.PhoneNumber, u.VerifyCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": err.Error(),
		})
		return
	}
	err = service.RegisterOrLogin(c, u.Username, u.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"info": "登录成功",
	})
}
