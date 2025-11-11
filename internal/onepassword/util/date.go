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

// SecondsToYYYYMMDD converts a second string to UTC and (YYYY-MM-DD) format.
func SecondsToYYYYMMDD(secondsStr string) (string, error) {
	seconds, err := strconv.ParseInt(secondsStr, 10, 64)
	if err != nil {
		return "", err
	}
	t := time.Unix(seconds, 0).UTC()
	return t.Format(time.DateOnly), nil
}

// YYYYMMDDToSeconds converts a YYYY-MM-DD date string to a Unix timestamp (seconds) string.
func YYYYMMDDToSeconds(dateStr string) (string, error) {
	t, err := time.ParseInLocation(time.DateOnly, dateStr, time.Local)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(t.Unix(), 10), nil
}
