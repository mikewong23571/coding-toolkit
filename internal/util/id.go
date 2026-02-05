package util

import (
	"crypto/sha1"
	"fmt"
)

func GenID(input string) string {
	h := sha1.Sum([]byte(input))
	return fmt.Sprintf("%x", h)[:6]
}
