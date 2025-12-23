package generator

import (
	"crypto/rand"
	"encoding/base64"
	"url-shortener/internal/domain"
)

const (
	ShortCodeLength = 3 // length of the generated short code
)

type RandomShortCodeGenerator struct{}

func NewRandomShortCodeGenerator() domain.ShortCodeGenerator {
	return &RandomShortCodeGenerator{}
}

func (g *RandomShortCodeGenerator) Generate() string {
	bytes := make([]byte, ShortCodeLength)
	rand.Read(bytes)
	code := base64.URLEncoding.EncodeToString(bytes)[:ShortCodeLength]
	code = sanitizeCode(code)
	return code
}

func sanitizeCode(code string) string {
	result := make([]rune, 0, len(code))
	for _, r := range code {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result = append(result, r)
		} else {
			result = append(result, 'x')
		}
	}
	return string(result)
}
