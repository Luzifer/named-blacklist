package fqdn

import (
	"net"
	"strings"
	"unicode"

	"golang.org/x/net/idna"
)

// IsValidEntry checks the domain names against a validation subset in
// order to find entries not being domains
func IsValidEntry(input string) bool {
	if isValidForPlainASCIICheck(input) {
		return true
	}

	// Normalize using IDNA and check if it's valid afterwards
	_, err := idna.Lookup.ToASCII(input)
	return err == nil
}

func isValidForPlainASCIICheck(input string) bool {
	input = strings.TrimSpace(input)

	// Reject empty or obviously bogus
	if input == "" || len(input) > 253 {
		return false
	}

	// Must not be an IP address
	if net.ParseIP(input) != nil {
		return false
	}

	// Remove optional trailing dot
	input = strings.TrimSuffix(input, ".")

	labels := strings.Split(input, ".")
	if len(labels) < 2 { //nolint:mnd
		return false
	}

	for _, label := range labels {
		if len(label) < 1 || len(label) > 63 {
			return false
		}

		for _, r := range label {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
				continue
			}
			return false
		}
	}

	return true
}
