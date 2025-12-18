package generator

import (
	"strings"
	"testing"

	"github.com/Luzifer/named-blacklist/pkg/config"
	"github.com/Luzifer/named-blacklist/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSimpleInline(t *testing.T) {
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
