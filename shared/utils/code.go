package utils

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
)

func GenerateHashCode(s string, size ...int) string {
	// bdf: GenerateHasCode(s)
	n := 10
	if len(size) > 0 && size[0] > 0 {
		n = size[0]
	}
	hash := sha1.Sum([]byte(s))
	return fmt.Sprintf("%x", hash)[:n]
}

func GenerateReferenceCode(s string) string {
	refCode := GenerateHashCode(s, 7)
	return strings.ToUpper(refCode)
}

func GenerateQRCodeString(typ, name string) string {
	randomStr := GenerateRandomString(12)
	buf := fmt.Sprintf("%s:%s:%s", typ, randomStr, name)
	return base64.StdEncoding.EncodeToString([]byte(buf))
}

func GenerateQRCodeStringByID(typ string, id int) string {
	randomStr := GenerateRandomString(12)
	buf := fmt.Sprintf("%s:%s:%d", typ, randomStr, id)
	return base64.StdEncoding.EncodeToString([]byte(buf))
}

func GenerateQRCodeStringForProduct(typ, ref string) string {
	buf := fmt.Sprintf("%s:%s", typ, ref)
	encoded := base64.StdEncoding.EncodeToString([]byte(buf))
	return fmt.Sprintf("honvang://qr/%s", encoded)
}

func GenerateQRCodeStringForUser(ref string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(ref))
	return fmt.Sprintf("honvang://user/%s", encoded)
}
