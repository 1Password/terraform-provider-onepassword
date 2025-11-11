package util

import (
	"strconv"
	"time"
)

// IsValidDateFormat checks if provided string is in valid 1Password DATE format (YYYY-MM-DD)
func IsValidDateFormat(dateString string) bool {
	_, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return false
	}
	return true
}

// SecondsToYYYYMMDD converts a seconds string to a (YYYY-MM-DD) formatted secondsStr string
// The date is formatted in UTC to avoid timezone-related day shifts
func SecondsToYYYYMMDD(secondsStr string) (string, error) {
	seconds, err := strconv.ParseInt(secondsStr, 10, 64)
	if err != nil {
		return "", err
	}
	t := time.Unix(seconds, 0).UTC()
	return t.Format(time.DateOnly), nil
}
