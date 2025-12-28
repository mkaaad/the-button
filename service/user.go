package service

import (
	"button/dao"
	"button/errorx"
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterOrLogin(c *gin.Context, userName string, phoneNumber string) error {
	u, err := dao.FindUserByUsername(userName)
	if err != nil {
		err = &errorx.DatabaseErr{}
		return err
	}
	if u.PhoneNumber != "" {
		err = &errorx.UsernameExistErr{}
		return err
	}
	u, err = dao.FindUserByPhoneNumber(phoneNumber)
	if err != nil {
		err = &errorx.DatabaseErr{}
		return err
	}
	if u.Username != "" {
		return setSessionIDToCookie(c, u.Username)
	}
	u.PhoneNumber = phoneNumber
	u.Username = userName
	err = dao.CreatUser(&u)
	if err != nil {
		err = &errorx.DatabaseErr{}
		return err
	}
	return setSessionIDToCookie(c, u.Username)
}
func setSessionIDToCookie(c *gin.Context, username string) error {
	sessionID := uuid.New().String()
	err := dao.Rdb.Set(context.Background(), sessionID, username, 24*time.Hour).Err()
	if err != nil {
		return &errorx.DatabaseErr{}
	}
	c.SetCookie("session_id", sessionID, 60*60*24, "/", "", false, true)
	return nil
}
