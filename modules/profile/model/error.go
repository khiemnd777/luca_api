package profileError

import "errors"

var (
	ErrPhoneExists = errors.New("không thể xử lý yêu cầu")
	ErrEmailExists = errors.New("không thể xử lý yêu cầu")
)
