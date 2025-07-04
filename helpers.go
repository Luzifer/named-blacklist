package main

import (
	"fmt"
	"strings"

	"golang.org/x/net/idna"

	"github.com/Luzifer/go_helpers/v2/str"
)

var genericBlacklist = []string{
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

func lineIsComment(line string) bool {
	if len(strings.TrimSpace(line)) == 0 {
		return true
	}

	return line[0] == '#' || line[0] == ';' || line[0] == '!'
}

func isBlacklisted(domain string) bool {
	return str.StringInSlice(domain, genericBlacklist)
}

func domainToPunycode(name string) (s string, err error) {
	if s, err = idna.ToASCII(name); err != nil {
		return s, fmt.Errorf("converting domain to punycode: %w", err)
	}
	return s, nil
}
