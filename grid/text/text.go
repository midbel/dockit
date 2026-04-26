package text

import (
	"strings"
	"unicode/utf8"
)

type MatchMode int

const (
	ExactOrWildcard  MatchMode = iota // match_type = 0
	ApproxAscending                   // match_type = 1
	ApproxDescending                  // match_type = -1
)

func Match(str, pattern string, mode MatchMode) bool {
	var (
		offset       int
		ptr          int
		lastStarPos  = -1
		lastValidPtr = -1
	)
	pattern = strings.ToLower(pattern)
	str = strings.ToLower(str)
	for offset < len(pattern) && ptr < len(str) {
		c, z := utf8.DecodeRuneInString(pattern[offset:])
		offset += z
		switch c {
		case '*':
			lastStarPos = offset - z
			lastValidPtr = ptr
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
			if x == c {
				ptr += z
				continue
			}
			if lastStarPos >= 0 {
				_, z := utf8.DecodeRuneInString(str[lastValidPtr:])
				lastValidPtr += z
				ptr = lastValidPtr
				offset = lastStarPos
				continue
			}
			return false
		default:
			x, z := utf8.DecodeRuneInString(str[ptr:])
			if x == c {
				ptr += z
				continue
			}
			if lastStarPos >= 0 {
				_, z := utf8.DecodeRuneInString(str[lastValidPtr:])
				lastValidPtr += z
				ptr = lastValidPtr
				offset = lastStarPos
				continue
			}
			return false
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
