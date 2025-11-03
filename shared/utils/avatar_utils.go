package utils

import (
	"fmt"
	"net/url"
)

func GetDummyAvatarURL(username string) string {
	safeName := url.QueryEscape(username)
	return fmt.Sprintf("https://api.dicebear.com/9.x/initials/svg?seed=%s", safeName)
}
