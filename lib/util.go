package constago

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// toCamelCase converts a string to camelCase
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	words := splitIntoWords(s)

	return arrayToCamelCase(words)
}

func arrayToCamelCase(words []string) string {

	if len(words) == 0 {
		return ""
	}

	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if words[i] != "" {
			result += cases.Title(language.Und, cases.NoLower).String(strings.ToLower(words[i]))
		}
	}
	return result
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	if s == "" {
		return s
	}

	words := splitIntoWords(s)

	return arrayToPascalCase(words)
}

func arrayToPascalCase(words []string) string {
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	for _, word := range words {
		if word != "" {
			result.WriteString(cases.Title(language.Und, cases.NoLower).String(strings.ToLower(word)))
		}
	}

	return result.String()
}

// splitIntoWords splits a string into words based on various separators
func splitIntoWords(s string) []string {
	if s == "" {
		return []string{}
	}

	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		// Check for various separators
		if isSeparator(r) {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		} else if unicode.IsUpper(r) {
			// Handle camelCase/PascalCase boundaries
			if currentWord.Len() > 0 && !unicode.IsUpper(rune(s[i-1])) {
				// Previous character was lowercase, this is uppercase - start new word
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
			currentWord.WriteRune(r)
		} else {
			currentWord.WriteRune(r)
		}
	}

	// Add the last word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// isSeparator checks if a rune is a word separator
func isSeparator(r rune) bool {
	return r == '_' || r == '-' || r == ' ' || r == '.' || r == '/'
}

func boolPtr(b bool) *bool {
	return &b
}

// isValidGoIdentifier checks if a string is a valid Go identifier
func isValidGoIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	// First character must be letter or underscore
	if !((s[0] >= 'a' && s[0] <= 'z') || (s[0] >= 'A' && s[0] <= 'Z') || s[0] == '_') {
		return false
	}

	// Remaining characters must be letters, digits, or underscores
	for i := 1; i < len(s); i++ {
		if !((s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= '0' && s[i] <= '9') || s[i] == '_') {
			return false
		}
	}

	return true
}

// isValidSource checks if a source pattern is valid
func isValidSource(pattern string) bool {
	// package:mypkg
	if strings.HasPrefix(pattern, "package:") {
		return isValidGoIdentifier(strings.TrimPrefix(pattern, "package:"))
	}

	if !strings.HasSuffix(pattern, ".go") {
		return false
	}

	// Normalize separators before validating
	p := filepath.ToSlash(filepath.Clean(pattern))

	// Validate glob syntax (supports ** and {a,b})
	if !doublestar.ValidatePattern(p) {
		return false
	}

	return true
}

func isStringBlank[T ~string](s T) bool {
	return len(strings.TrimSpace(string(s))) == 0
}
