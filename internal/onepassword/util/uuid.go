package util

import "regexp"

// IsValidUUID return 'true' if the provided string is valid 1Password UUID.
func IsValidUUID(u string) bool {
	r := regexp.MustCompile("^[a-z0-9]{26}$")
	return r.MatchString(u)
}
