package service

import (
	"button/config"
	"button/errorx"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi "github.com/alibabacloud-go/dypnsapi-20170525/v2/client"
)

var (
	clinet        *dypnsapi.Client
	signName            = "速通互联验证码"
	templateCode        = "100001"
	templateParam       = `{"code":"##code##","min":"5"}`
	CodeLenth     int64 = 6
)

func InitSMSClient() {
	config := &openapi.Config{
		AccessKeyId:     &config.AccessKeyID,
		AccessKeySecret: &config.AccessKeySecret,
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
		SignName:      &signName,
		TemplateCode:  &templateCode,
		TemplateParam: &templateParam,
		CodeLength:    &CodeLenth,
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
