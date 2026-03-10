package model

// EnvironmentVariable represents an environment variable from a 1Password Environment.
type EnvironmentVariable struct {
	Name   string
	Value  string
	Masked bool
}
