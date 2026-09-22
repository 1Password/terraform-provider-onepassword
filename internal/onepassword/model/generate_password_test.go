package model

import (
	"regexp"
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		recipe   *GeneratorRecipe
		validate func(t *testing.T, password string)
	}{
		"generates a random password": {
			recipe: &GeneratorRecipe{
				Kind:          RecipeKindRandom,
				Length:        20,
				CharacterSets: []CharacterSet{CharacterSetDigits, CharacterSetSymbols},
			},
			validate: func(t *testing.T, password string) {
				if len(password) != 20 {
					t.Errorf("length: got %d, want 20", len(password))
				}
				if !regexp.MustCompile(`[0-9]`).MatchString(password) {
					t.Errorf("digits: got %q, want at least one digit", password)
				}
				if !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
					t.Errorf("symbols: got %q, want at least one symbol", password)
				}
			},
		},
		"generates a memorable password": {
			recipe: &GeneratorRecipe{
				Kind:       RecipeKindMemorable,
				WordCount:  3,
				Separator:  "hyphens",
				WordList:   "full_words",
				Capitalize: false,
			},
			validate: func(t *testing.T, password string) {
				// Three full words separated by hyphens: word-word-word
				if !regexp.MustCompile(`^[a-z]+(-[a-z]+){2}$`).MatchString(password) {
					t.Errorf("memorable format: got %q, want three lowercase hyphen-separated words", password)
				}
			},
		},
		"generates a memorable password with spaces and capitalization": {
			recipe: &GeneratorRecipe{
				Kind:       RecipeKindMemorable,
				WordCount:  3,
				Separator:  "spaces",
				WordList:   "syllables",
				Capitalize: true,
			},
			validate: func(t *testing.T, password string) {
				if !regexp.MustCompile(`^[A-Za-z]+( [A-Za-z]+){2}$`).MatchString(password) {
					t.Errorf("memorable format: got %q, want three space-separated words", password)
				}
			},
		},
		"generates a pin": {
			recipe: &GeneratorRecipe{
				Kind:   RecipeKindPin,
				Length: 6,
			},
			validate: func(t *testing.T, password string) {
				if !regexp.MustCompile(`^[0-9]{6}$`).MatchString(password) {
					t.Errorf("pin format: got %q, want 6 digits", password)
				}
			},
		},
		"unset kind defaults to random": {
			recipe: &GeneratorRecipe{
				Length:        16,
				CharacterSets: []CharacterSet{CharacterSetDigits},
			},
			validate: func(t *testing.T, password string) {
				if len(password) != 16 {
					t.Errorf("length: got %d, want 16", len(password))
				}
			},
		},
		"generates a random password excluding characters": {
			recipe: &GeneratorRecipe{
				Kind:              RecipeKindRandom,
				Length:            32,
				CharacterSets:     []CharacterSet{CharacterSetDigits, CharacterSetSymbols},
				ExcludeCharacters: "l1IO0",
			},
			validate: func(t *testing.T, password string) {
				if len(password) != 32 {
					t.Errorf("length: got %d, want 32", len(password))
				}
				if strings.ContainsAny(password, "l1IO0") {
					t.Errorf("excluded characters present: got %q, want none of %q", password, "l1IO0")
				}
				if !regexp.MustCompile(`[0-9]`).MatchString(password) {
					t.Errorf("digits: got %q, want at least one digit despite exclusions", password)
				}
			},
		},
		"exclusions apply without digits or symbols": {
			recipe: &GeneratorRecipe{
				Kind:              RecipeKindRandom,
				Length:            20,
				ExcludeCharacters: "aeiou",
			},
			validate: func(t *testing.T, password string) {
				if strings.ContainsAny(password, "aeiou") {
					t.Errorf("excluded characters present: got %q, want none of %q", password, "aeiou")
				}
			},
		},
	}

	for description, test := range tests {
		t.Run(description, func(t *testing.T) {
			t.Parallel()
			password, err := generatePassword(test.recipe)
			if err != nil {
				t.Fatalf("generatePassword() error = %v", err)
			}
			if password == "" {
				t.Fatal("generatePassword() returned an empty password")
			}
			test.validate(t, password)
		})
	}
}

func TestGeneratePasswordUnsatisfiableExclusions(t *testing.T) {
	t.Parallel()

	_, err := generatePassword(&GeneratorRecipe{
		Kind:              RecipeKindRandom,
		Length:            20,
		CharacterSets:     []CharacterSet{CharacterSetDigits},
		ExcludeCharacters: "0123456789",
	})
	if err == nil {
		t.Fatal("generatePassword() want error for exclusions conflicting with required digits, got nil")
	}
	if !strings.Contains(err.Error(), "excluding") {
		t.Errorf("generatePassword() error = %v, want it to mention the exclusion set", err)
	}
}
