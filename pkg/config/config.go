package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	korvike "github.com/Luzifer/korvike/functions"
	"github.com/Luzifer/named-blacklist/pkg/helpers"
)

const (
	defaultTemplate = `$TTL 1H

@ SOA LOCALHOST. dns-master.localhost. (1 1h 15m 30d 2h)
  NS  LOCALHOST.

; Blacklist entries
{{ range .blacklist -}}
{{ to_punycode .Domain }} CNAME . ; {{ .Comments }}
{{ end }}`
)

type (
	// File represents the format the configuration file is expected in
	File struct {
		Providers []ProviderDefinition `yaml:"providers"`

		Template         string             `yaml:"template"`
		CompiledTemplate *template.Template `yaml:"-"`
	}

	// ProviderAction defines the available actions to take with the provider
	ProviderAction string

	// ProviderDefinition describes a provider to use for gathering domains
	ProviderDefinition struct {
		Action  ProviderAction `yaml:"action"`
		Content string         `yaml:"content"`
		File    string         `yaml:"file"`
		Name    string         `yaml:"name"`
		Type    ProviderType   `yaml:"type"`
		URL     string         `yaml:"url"`
	}

	// ProviderType defines the type of provider to execute for this list
	ProviderType string
)

const (
	// ProviderActionBlacklist defines all domain results should be blocked
	ProviderActionBlacklist ProviderAction = "blacklist"
	// ProviderActionWhitelist defines all domain results should be unblocked
	ProviderActionWhitelist ProviderAction = "whitelist"
)

// LoadConfigFile reads the configuration and parses the template
func LoadConfigFile(filename string) (*File, error) {
	f, err := os.Open(filename) //#nosec:G304 // Intended to load given config file
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.WithError(err).Error("closing config file")
		}
	}()

	out := &File{Template: defaultTemplate}
	if err = yaml.NewDecoder(f).Decode(out); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	funcs := korvike.GetFunctionMap()
	funcs["to_punycode"] = helpers.DomainToPunycode
	funcs["join"] = strings.Join
	funcs["sort"] = func(in []string) []string {
		sort.Slice(in, func(i, j int) bool { return strings.ToLower(in[i]) < strings.ToLower(in[j]) })
		return in
	}

	if out.CompiledTemplate, err = template.
		New("configTemplate").
		Funcs(funcs).
		Parse(out.Template); err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return out, nil
}

// GetContent retrieves the content of the given list for parsing with
// a provider
func (p ProviderDefinition) GetContent(appVersion string) (io.ReadCloser, error) {
	switch {
	case p.Content != "":
		return io.NopCloser(strings.NewReader(p.Content)), nil

	case p.File != "":
		f, err := os.Open(p.File)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		return f, nil

	case p.URL != "":
		return p.fetchURLContent(appVersion)

	default:
		return nil, fmt.Errorf("neither file nor URL specified")
	}
}

func (p ProviderDefinition) fetchURLContent(version string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, p.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("named-blacklist %s (https://github.com/Luzifer/named-blacklist)", version))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unexected status %d", resp.StatusCode)
	}

	return resp.Body, nil
}
