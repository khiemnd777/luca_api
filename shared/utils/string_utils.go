package utils

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"

	"github.com/mozillazg/go-unidecode"
	"golang.org/x/text/unicode/norm"
)

func GenerateRandomString(length int) string {
	// e.g.: andy -> d4e5a7c8e3f9a1b2
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)[:length]
}

func NormalizeText(s string) string {
	return unidecode.Unidecode(s)
}

var wspRe = regexp.MustCompile(`\s+`)

func RemoveVietnameseDiacritics(s string) string {
	decomp := norm.NFD.String(s)
	sb := make([]rune, 0, len(decomp))
	for _, r := range decomp {
		switch r {
		case 'đ':
			r = 'd'
		case 'Đ':
			r = 'D'
		}
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		sb = append(sb, r)
	}
	return string(sb)
}

func NormalizeSearchKeyword(s string) string {
	s = wspRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	return RemoveVietnameseDiacritics(s)
}

func NormalizeSplit(s *string, sep string) []string {
	raw := ""
	if s != nil {
		raw = *s
	}

	result := make([]string, 0)

	for _, p := range strings.Split(raw, sep) {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}
