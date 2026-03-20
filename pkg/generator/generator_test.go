package generator

import (
	"strings"
	"testing"

	"github.com/Luzifer/named-blacklist/pkg/config"
	"github.com/Luzifer/named-blacklist/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateBlacklistDefaultBehavior(t *testing.T) {
	b, err := GenerateBlacklist("testing", []config.ProviderDefinition{
		{
			Action: config.ProviderActionBlacklist,
			Content: strings.Join([]string{
				"a.example.com",
				"b.example.com",
				"c.example.com",
			}, "\n"),
			Name: "Local Blacklist",
			Type: "domain-list",
		},
		{
			Action: config.ProviderActionBlacklist,
			Content: strings.Join([]string{
				"a.example.com",
			}, "\n"),
			Name: "Second Local Blacklist",
			Type: "domain-list",
		},
		{
			Action: config.ProviderActionWhitelist,
			Content: strings.Join([]string{
				"b.example.com",
			}, "\n"),
			Name: "Local Whitelist",
			Type: "domain-list",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, []provider.Entry{
		{Domain: "a.example.com", Comments: []string{"Local Blacklist", "Second Local Blacklist"}},
		{Domain: "c.example.com", Comments: []string{"Local Blacklist"}},
	}, b)
}

func TestGenerateBlacklistMinMatches(t *testing.T) {
	b, err := GenerateBlacklist("testing", []config.ProviderDefinition{
		{
			Action: config.ProviderActionBlacklist,
			Content: strings.Join([]string{
				"once.example.com",
				"pair.example.com",
				"triple.example.com",
				"duplicate.example.com",
				"duplicate.example.com",
				"whitelisted.example.com",
			}, "\n"),
			MinMatches: 1,
			Name:       "Trusted Feed",
			Type:       "domain-list",
		},
		{
			Action: config.ProviderActionBlacklist,
			Content: strings.Join([]string{
				"needs-confirmation.example.com",
				"pair.example.com",
				"triple.example.com",
				"duplicate.example.com",
				"whitelisted.example.com",
			}, "\n"),
			MinMatches: 2,
			Name:       "Noisy Feed",
			Type:       "domain-list",
		},
		{
			Action: config.ProviderActionBlacklist,
			Content: strings.Join([]string{
				"pair.example.com",
				"triple.example.com",
			}, "\n"),
			MinMatches: 3,
			Name:       "Strict Feed",
			Type:       "domain-list",
		},
		{
			Action: config.ProviderActionWhitelist,
			Content: strings.Join([]string{
				"whitelisted.example.com",
			}, "\n"),
			MinMatches: 9,
			Name:       "Whitelist",
			Type:       "domain-list",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, []provider.Entry{
		{Domain: "duplicate.example.com", Comments: []string{"Trusted Feed", "Noisy Feed"}},
		{Domain: "once.example.com", Comments: []string{"Trusted Feed"}},
		{Domain: "pair.example.com", Comments: []string{"Trusted Feed", "Noisy Feed", "Strict Feed"}},
		{Domain: "triple.example.com", Comments: []string{"Trusted Feed", "Noisy Feed", "Strict Feed"}},
	}, b)
}
