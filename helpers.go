package main

import (
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

func domainToPunycode(name string, v ...string) (string, error) {
	return idna.ToASCII(name)
}
