package common

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

func ToE164(phone string) string {
	re := regexp.MustCompile(`[^+0-9]`)
	cleaned := re.ReplaceAllString(phone, "")
	if strings.HasPrefix(cleaned, "+1") {
		return cleaned
	}

	if len(cleaned) == 10 {
		return "+1" + cleaned
	}

	if len(cleaned) == 11 && strings.HasPrefix(cleaned, "1") {
		return "+" + cleaned
	}

	return cleaned
}

func To11Digit(phone string) string {
	re := regexp.MustCompile(`[^+0-9]`)
	cleaned := re.ReplaceAllString(phone, "")
	cleaned = strings.TrimPrefix(cleaned, "+")
	if len(cleaned) == 11 && strings.HasPrefix(cleaned, "1") {
		return cleaned
	}

	return "1" + cleaned
}

// FormatPhoneNumber formats a phone number based on the given pattern.
func FormatPhoneNumber(phoneNumber, pattern string) (string, error) {
	// Extract only numeric digits from the input
	cleanedNumber := ""
	for _, char := range phoneNumber {
		if unicode.IsDigit(char) {
			cleanedNumber += string(char)
		}
	}

	// Count the number of 'x' in the pattern (this determines how many digits are needed)
	requiredDigits := strings.Count(pattern, "x")

	// Ensure we have enough digits, otherwise return an error
	if len(cleanedNumber) < requiredDigits {
		return "", errors.New("invalid phone number: not enough digits")
	}

	// Take only the last `requiredDigits` numbers
	cleanedNumber = cleanedNumber[len(cleanedNumber)-requiredDigits:]

	// Format the number according to the pattern
	result := ""
	digitIndex := 0

	for i := 0; i < len(pattern); i++ {
		if pattern[i] == 'x' && digitIndex < len(cleanedNumber) {
			result += string(cleanedNumber[digitIndex])
			digitIndex++
		} else {
			result += string(pattern[i])
		}
	}

	return result, nil
}
