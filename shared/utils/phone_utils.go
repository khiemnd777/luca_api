package utils

import (
	"regexp"
	"strings"
)

func NormalizePhone(input *string) string {
	if input == nil {
		return ""
	}

	re := regexp.MustCompile(`\D`)
	digits := re.ReplaceAllString(*input, "")

	if strings.HasPrefix(digits, "84") {
		return "+" + digits
	} else if strings.HasPrefix(digits, "0") {
		return "+84" + digits[1:]
	}
	return "+" + digits
}

func IsPhone(input string) bool {
	var phoneRegex = regexp.MustCompile(`^(?:\+84|84|0)\d{9,10}$`)
	digits := strings.ReplaceAll(input, " ", "")
	digits = strings.ReplaceAll(digits, ".", "")
	digits = strings.ReplaceAll(digits, "-", "")
	return phoneRegex.MatchString(digits)
}

func NormalizeEnsuredPhone(phone *string) (*string, bool) {
	var normalizedPhone *string
	isPhone := false
	if phone != nil && IsPhone(*phone) {
		np := NormalizePhone(phone)
		if np != "" {
			normalizedPhone = &np
			isPhone = true
		}
	}
	return normalizedPhone, isPhone
}
