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
// The timestamp is interpreted as UTC to ensure consistent results regardless of the local timezone.
func SecondsToYYYYMMDD(secondsStr string) (string, error) {
	seconds, err := strconv.ParseInt(secondsStr, 10, 64)
	if err != nil {
		return "", err
	}
	t := time.Unix(seconds, 0).UTC()
	return t.Format(time.DateOnly), nil
}

// YYYYMMDDToSeconds converts a (YYYY-MM-DD) formatted date string to a Unix timestamp (seconds) in UTC.
//
// This is a workaround for a timezone issue in Connect: when Connect receives a date string (YYYY-MM-DD),
// it parses it using time.ParseInLocation with time.Local, which causes date shifts depending on Connect's timezone.
// By converting to a timestamp at 12:01:00 UTC before sending, we bypass Connect's timezone-dependent parsing
// and ensure the date is stored consistently regardless of where Connect is deployed.
func YYYYMMDDToSeconds(dateStr string) (string, error) {
	t, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		return "", err
	}
	// Use 12:01:00 UTC to match web client's modern date convention
	// This ensures the web client recognizes it as a "modern" date and uses UTC interpretation
	t = time.Date(t.Year(), t.Month(), t.Day(), 12, 1, 0, 0, time.UTC)
	return strconv.FormatInt(t.Unix(), 10), nil
}
