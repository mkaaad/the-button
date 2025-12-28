package model

type User struct {
	Username    string `json:"username" gorm:"unique;not null" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	VerifyCode  string `json:"verify_code" gorm:"-" binding:"required"`
}
