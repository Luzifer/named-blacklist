package helpers

import (
	"fmt"
	"strings"

	"golang.org/x/net/idna"

	"github.com/Luzifer/go_helpers/v2/str"
)

// GenericBlacklist contains a list of entries not to include in any
// list as those wouldn't be useful in any black-/whitelist
var GenericBlacklist = []string{
	"broadcasthost",
	"ip6-allhosts",
	"ip6-allnodes",
	"ip6-allrouters",
	"ip6-localnet",
	"ip6-mcastprefix",
	"local",
	"localhost",
	"localhost.localdomain",
}

// LineIsComment contains logic to filter out non-useful lines
func LineIsComment(line string) bool {
	if len(strings.TrimSpace(line)) == 0 {
		return true
	}

	return line[0] == '#' || line[0] == ';' || line[0] == '!'
}

// IsBlacklisted checks an entry against the GenericBlacklist
func IsBlacklisted(domain string) bool {
	return str.StringInSlice(domain, GenericBlacklist)
}

// DomainToPunycode converts a domain name into its punycode equivalent
func DomainToPunycode(name string) (s string, err error) {
	if s, err = idna.ToASCII(name); err != nil {
		return s, fmt.Errorf("converting domain to punycode: %w", err)
	}
	return s, nil
}
