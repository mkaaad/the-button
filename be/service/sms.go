package service

import (
	"button/errorx"
	"os"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi "github.com/alibabacloud-go/dypnsapi-20170525/v2/client"
	"github.com/joho/godotenv"
)

var (
	clinet        *dypnsapi.Client
	signName            = "速通互联验证码"
	templateCode        = "100001"
	templateParam       = `{"code":"##code##","min":"5"}`
	CodeLenth     int64 = 6
)

func InitSMSClient() {
	godotenv.Load()
	keyID, ok1 := os.LookupEnv("ACCESS_KEY_ID")
	keySecret, ok2 := os.LookupEnv("ACCESS_KEY_SECRET")
	if !ok1 || !ok2 {
		panic("failed to get sms key from env")
	}
	config := &openapi.Config{
		AccessKeyId:     &keyID,
		AccessKeySecret: &keySecret,
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
