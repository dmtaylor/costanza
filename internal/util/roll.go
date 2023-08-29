package util

import (
	"strings"
	"unicode"
)

// PreprocessRoll handles preprocessing roll data from user input. Currently, this converts "d20" -> "1d20" for common
// shorthand
func PreprocessRoll(input string) string {
	if len(input) < 1 { // bail early if empty string to avoid unnecessary allocation
		return input
	}
	inputr := []rune(input)
	result := strings.Builder{}
	result.Grow(len(input))

	for i, r := range inputr {
		if r == 'd' && (i == 0 || !unicode.IsNumber(inputr[i-1])) {
			result.WriteRune('1')
		}
		result.WriteRune(r)
	}

	return result.String()
}
