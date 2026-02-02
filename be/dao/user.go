package dao

import "button/model"

func CreatUser(u *model.User) error {
	return Ldb.Create(u).Error
}
func FindUserByPhoneNumber(phoneNumber string) (u model.User, err error) {
	err = Ldb.Limit(1).Where("phone_number = ?", phoneNumber).Find(&u).Error
	return
}
func FindUserByUsername(username string) (u model.User, err error) {
	err = Ldb.Limit(1).Where("username = ?", username).Find(&u).Error
	return
}
