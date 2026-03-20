package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFileDefaultsMinMatchesToOne(t *testing.T) {
	conf := writeConfigFile(t, `
providers:
  - name: Default Provider
    content: |
      example.com
    action: blacklist
    type: domain-list
`)
	t.Cleanup(func() { _ = os.Remove(conf) })

	cfg, err := LoadConfigFile(conf)
	require.NoError(t, err)
	require.Len(t, cfg.Providers, 1)
	assert.Equal(t, 1, cfg.Providers[0].MinMatches)
}

func TestLoadConfigFileRejectsInvalidMinMatches(t *testing.T) {
	for _, tc := range []struct {
		name       string
		minMatches int
		action     ProviderAction
	}{
		{name: "Zero Blacklist", minMatches: 0, action: ProviderActionBlacklist},
		{name: "Negative Blacklist", minMatches: -1, action: ProviderActionBlacklist},
		{name: "Zero Whitelist", minMatches: 0, action: ProviderActionWhitelist},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conf := writeConfigFile(t, fmt.Sprintf(`
providers:
  - name: Invalid Provider
    content: |
      example.com
    action: %s
    type: domain-list
    min_matches: %d
`, tc.action, tc.minMatches))
			t.Cleanup(func() { _ = os.Remove(conf) })

			_, err := LoadConfigFile(conf)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `provider "Invalid Provider" has invalid min_matches`)
		})
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	filename := filepath.Join(dir, "config.yaml")

	require.NoError(t, os.WriteFile(filename, []byte(content), 0o600))

	return filename
}
