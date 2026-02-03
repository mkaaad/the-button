package api

import (
	"button/model"
	"button/service"
	"bytes"
	"encoding/base64"
	"net/http"
	"regexp"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

var mobileRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

type verify struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	CaptchaID   string `form:"captcha_id" json:"captcha_id" binding:"required"` // 验证码ID
	Code        string `form:"code" json:"code" binding:"required,digit"`
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
	ok := captcha.VerifyString(v.CaptchaID, v.Code)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"info": "验证码错误",
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
	if len(u.VerifyCode) != int(service.CodeLenth) {
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
func GetCaptcha(c *gin.Context) {
	captchaID := captcha.New()
	buf := new(bytes.Buffer)
	err := captcha.WriteImage(buf, captchaID, captcha.StdWidth, captcha.StdHeight)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"info": "生成验证码失败",
		})
		return
	}
	imgBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	c.JSON(http.StatusOK, gin.H{
		"info": "success",
		"data": gin.H{
			"captcha_id":   captchaID, // 关键：后续验证必须携带此ID
			"image_base64": imgBase64, // 前端直接作为img的src
		},
	})
}
