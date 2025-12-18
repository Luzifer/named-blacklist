package fqdn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidFQDN(t *testing.T) {
	for name, expResult := range map[string]bool{
		"www.foo_bar.example.com": true,  // IDNA validation but accepted by DNS and browsers
		"www.foo.bar.example.com": true,  // Standard domain
		"bücher.example.com":      true,  // IDNA validation should be fine
		"||abp.example.com^":      false, // unparsed ABP rule
		"-foo.example.com":        true,  // Invalid host, browsers will try
		"foo-.example.com":        true,  // Invalid host, browsers will try
	} {
		assert.Equal(t, expResult, IsValidEntry(name), name)
	}
}
