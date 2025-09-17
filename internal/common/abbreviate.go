package common

import "strings"

// Abbreviate generates an abbreviation from the given text
func Abbreviate(text string) string {
	words := strings.Fields(text)
	abbr := strings.ToUpper(text[:minNum(3, len(text))])

	if len(words) == 2 {
		if len(words[0]) > 0 && len(words[1]) > 0 {
			abbr = strings.ToUpper(string(words[0][0]) + words[1][:minNum(2, len(words[1]))])
		}
	}

	if len(words) > 2 {
		if len(words[0]) > 0 && len(words[1]) > 0 && len(words[2]) > 0 {
			abbr = strings.ToUpper(string(words[0][0]) + string(words[1][0]) + string(words[2][0]))
		}
	}

	return abbr
}

// min returns the smaller of two integers
func minNum(a, b int) int {
	if a < b {
		return a
	}
	return b
}
