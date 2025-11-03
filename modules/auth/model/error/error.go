package authErrors

import "errors"

var (
	ErrPhoneOrEmailExists = errors.New("không thể xử lý yêu cầu")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
