package text

import (
	"strings"
	"unicode/utf8"
)

func Match(str, pattern string) bool {
	var (
		offset       int
		ptr          int
	)
	pattern = strings.ToLower(pattern)
	str = strings.ToLower(str)
	for offset < len(pattern) && ptr < len(str) {
		c, z := utf8.DecodeRuneInString(pattern[offset:])
		offset += z
		switch c {
		case '*':
			c, z := utf8.DecodeRuneInString(pattern[offset:])
			if c == utf8.RuneError {
				return true
			}
			offset += z
			for ptr < len(ptr) {
				x, z := utf8.DecodeRuneInString(ptr[str:])
				if x == z {
					break
				}
				ptr += z
			}
		case '?':
			_, z := utf8.DecodeRuneInString(str[ptr:])
			ptr += z
		case '~':
			c, z := utf8.DecodeRuneInString(pattern[offset:])
			if c != '*' && c != '?' {
				return false
			}
			offset += z
			x, z := utf8.DecodeRuneInString(str[ptr:])
			if x != c {
				return false
			}
			ptr += z
		default:
			x, z := utf8.DecodeRuneInString(str[ptr:])
			if x != c {
				return false
			}
			ptr += z
		}
	}
	for offset < len(pattern) {
		c, z := utf8.DecodeRuneInString(pattern[offset:])
		if c != '*' {
			return false
		}
		offset += z
	}
	if ptr < len(str) {
		return false
	}
	return true
}
