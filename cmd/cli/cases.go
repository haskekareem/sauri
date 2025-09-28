package main

import (
	"regexp"
	"strings"
	"unicode"
)

// capitalizeFirst converts the first letter of a word to uppercase
func capitalizeFirst(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// splitCompoundWord inserts a dash before known suffixes like "handler"
func splitCompoundWord(s string) string {
	knownSuffixes := []string{"handler", "controller"}
	for _, suffix := range knownSuffixes {
		if strings.HasSuffix(s, suffix) && len(s) > len(suffix) {
			return strings.TrimSuffix(s, suffix) + "-" + suffix
		}
	}
	return s
}

// normalizeSeparators converts underscores to dashes and removes duplicate dashes
func normalizeSeparators(s string) string {
	s = strings.ReplaceAll(s, "_", "-")                   // Convert all underscores to dashes
	s = regexp.MustCompile("-+").ReplaceAllString(s, "-") // Collapse multiple dashes into one
	return s
}

// toCamelCase converts a normalized dash-separated string to CamelCase
func toCamelCase(s string) string {
	// Split string into parts using "-"
	parts := strings.Split(s, "-")
	for i := range parts {
		// Capitalize each part
		parts[i] = capitalizeFirst(parts[i])
	}
	// Join without separators
	return strings.Join(parts, "")
}

// convertInput handles all edge cases and returns clean CamelCase
func convertInput(raw string) string {
	raw = strings.ToLower(raw)     // Make it lowercase
	raw = splitCompoundWord(raw)   // insert dash for known suffixes
	raw = normalizeSeparators(raw) // change _ to - and remove double --
	return toCamelCase(raw)        // final CamelCase conversion
}
