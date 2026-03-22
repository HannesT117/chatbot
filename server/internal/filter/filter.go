package filter

import "strings"

// ContainsBlocked reports whether text contains any term from the blocklist (case-insensitive).
func ContainsBlocked(text string, terms []string) bool {
	lower := strings.ToLower(text)
	for _, term := range terms {
		if strings.Contains(lower, strings.ToLower(term)) {
			return true
		}
	}
	return false
}

// ContainsCanary reports whether text contains the exact canary token.
func ContainsCanary(text, canaryToken string) bool {
	return strings.Contains(text, canaryToken)
}
