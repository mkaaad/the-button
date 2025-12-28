package service

import (
	"button/config"
	"button/errorx"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi "github.com/alibabacloud-go/dypnsapi-20170525/v2/client"
)

var (
	clinet *dypnsapi.Client
)

func InitSMSClient() {
	config := &openapi.Config{
		AccessKeyId:     &config.ACCESS_KEY_ID,
		AccessKeySecret: &config.ACCESS_KEY_SECRET,
	}
	c, err := dypnsapi.NewClient(config)
	if err != nil {
		panic(err)
	}
	clinet = c

}
func SendVerifyCode(phoneNumber string) error {
	req := &dypnsapi.SendSmsVerifyCodeRequest{
		PhoneNumber:   &phoneNumber,
		SignName:      &config.SignName,
		TemplateCode:  &config.TemplateCode,
		TemplateParam: &config.TemplateParam,
		CodeLength:    &config.CodeLenth,
	}
	resp, err := clinet.SendSmsVerifyCode(req)
	if err != nil {
		return &errorx.SMSSendErr{}
	}
	if resp.Body.Code == nil {
		return &errorx.SMSRespCodeNullErr{}
	}
	code := *resp.Body.Code
	if code == "biz.FREQUENCY" {
		return &errorx.SMSFrequenctErr{}
	}
	if code != "OK" {
		return &errorx.SMSSendErr{Code: code}
	}
	return nil
}
func VerifyCode(phoneNumber, verifyCode string) error {
	varifyReq := &dypnsapi.CheckSmsVerifyCodeRequest{
		PhoneNumber: &phoneNumber,
		VerifyCode:  &verifyCode,
	}
	resp, err := clinet.CheckSmsVerifyCode(varifyReq)
	if err != nil {
		return &errorx.SMSSendErr{}
	}
	if resp.Body.Code == nil {
		return &errorx.SMSRespCodeNullErr{}
	}
	code := *resp.Body.Code
	if code != "OK" {
		return &errorx.SMSVerifyErr{}
	}
	return nil
}
