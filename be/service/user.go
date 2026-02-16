package service

import (
	"button/dao"
	"button/errorx"
	"button/model"
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterOrLogin(c *gin.Context, username string, phoneNumber string) (string, string, error) {
	u, err := dao.FindUserByUsername(username)
	if err != nil {
		err = &errorx.DatabaseErr{}
		return "", "", err
	}
	// Username exists and is bound to another phone number.
	if u.Username != "" && u.PhoneNumber != phoneNumber {
		err = &errorx.UsernameExistErr{}
		return "", "", err
	}

	u, err = dao.FindUserByPhoneNumber(phoneNumber)
	if err != nil {
		err = &errorx.DatabaseErr{}
		return "", "", err
	}

	// Phone number already exists but binds to another username.
	if u.Username != "" && u.Username != username {
		err = &errorx.UsernameExistErr{}
		return "", "", err
	}

	if u.Username == "" {
		newUser := model.User{
			PhoneNumber: phoneNumber,
			Username:    username,
		}
		err = dao.CreatUser(&newUser)
		if err != nil {
			err = &errorx.DatabaseErr{}
			return "", "", err
		}

	}
	sessionID := uuid.New().String()
	err = dao.Rdb.Set(context.Background(), sessionID, username, 24*time.Hour).Err()
	if err != nil {
		return "", "", &errorx.DatabaseErr{}
	}
	return sessionID, username, nil
}
func IsLogin(sessionID string, send chan []byte) (string, bool) {
	username, err := dao.Rdb.Get(context.Background(), sessionID).Result()
	if err != nil {
		msg := message{
			Type: "unauthorized",
		}
		data, _ := json.Marshal(msg)
		send <- data
		return "", false

	}
	return username, true
}
