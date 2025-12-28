package errorx

type UsernameExistErr struct{}

func (e *UsernameExistErr) Error() string {
	return "用户名已存在"
}

type DatabaseErr struct{}

func (e *DatabaseErr) Error() string {
	return "数据库错误"
}
