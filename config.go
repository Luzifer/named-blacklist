package main

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
)

const (
	defaultTemplate = `$TTL 1H

@ SOA LOCALHOST. dns-master.localhost. (1 1h 15m 30d 2h)
  NS  LOCALHOST.

; Blacklist entries
{{ range .blacklist -}}
{{ to_punycode .Domain }} CNAME . ; {{ .Comment }}
{{ end }}`
)

type (
	configfile struct {
		Providers []providerDefinition `yaml:"providers"`

		Template string `yaml:"template"`
		tpl      *template.Template
	}

	providerAction string

	providerDefinition struct {
		Action  providerAction `yaml:"action"`
		Content string         `yaml:"content"`
		File    string         `yaml:"file"`
		Name    string         `yaml:"name"`
		Type    providerType   `yaml:"type"`
		URL     string         `yaml:"url"`
	}

	providerType string
)

const (
	providerActionBlacklist providerAction = "blacklist"
	providerActionWhitelist providerAction = "whitelist"
)

func loadConfigFile(filename string) (*configfile, error) {
	f, err := os.Open(filename) //#nosec:G304 // Intended to load given config file
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.WithError(err).Error("closing config file")
		}
	}()

	out := &configfile{Template: defaultTemplate}
	if err = yaml.NewDecoder(f).Decode(out); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	funcs := korvike.GetFunctionMap()
	funcs["to_punycode"] = domainToPunycode
	funcs["join"] = strings.Join
	funcs["sort"] = func(in []string) []string {
		sort.Slice(in, func(i, j int) bool { return strings.ToLower(in[i]) < strings.ToLower(in[j]) })
		return in
	}

	if out.tpl, err = template.
		New("configTemplate").
		Funcs(funcs).
		Parse(out.Template); err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return out, nil
}

func (p providerDefinition) GetContent() (io.ReadCloser, error) {
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
		return p.fetchURLContent()

	default:
		return nil, fmt.Errorf("neither file nor URL specified")
	}
}

func (p providerDefinition) fetchURLContent() (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, p.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

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
