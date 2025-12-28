package errorx

import "fmt"

type SMSRespCodeNullErr struct{}

func (e *SMSRespCodeNullErr) Error() string {
	return "与SMS服务器通信失败"
}

type SMSSendErr struct {
	Code string
}

func (e *SMSSendErr) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("发送验证码失败：%s", e.Code)
	}
	return "验证失败"
}

type SMSVerifyErr struct {
	Code string
}

func (e *SMSVerifyErr) Error() string {
	return fmt.Sprintf("验证验证码失败：%s", e.Code)
}

type SMSFrequenctErr struct{}

func (e *SMSFrequenctErr) Error() string {
	return "发送验证码太过频繁"
}
