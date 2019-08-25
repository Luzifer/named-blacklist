package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const defaultTemplate = `$TTL 1H

@ SOA LOCALHOST. dns-master.localhost. (1 1h 15m 30d 2h)
  NS  LOCALHOST.

; Blacklist entries
{{ range .blacklist -}}
{{ to_punycode .Domain }} CNAME . ; {{ .Comment }}
{{ end }}`

type providerType string

type providerAction string

const (
	providerActionBlacklist providerAction = "blacklist"
	providerActionWhitelist providerAction = "whitelist"
)

type configfile struct {
	Providers []providerDefinition `yaml:"providers"`

	Template string `yaml:"template"`
	tpl      *template.Template
}

type providerDefinition struct {
	Action  providerAction `yaml:"action"`
	Content string         `yaml:"content"`
	File    string         `yaml:"file"`
	Name    string         `yaml:"name"`
	Type    providerType   `yaml:"type"`
	URL     string         `yaml:"url"`
}

func (p providerDefinition) GetContent() (io.ReadCloser, error) {
	switch {

	case p.Content != "":
		return ioutil.NopCloser(strings.NewReader(p.Content)), nil

	case p.File != "":
		return os.Open(p.File)

	case p.URL != "":
		resp, err := http.Get(p.URL) //nolint:bodyclose // This does not need to be closed here and would break stuff
		return resp.Body, err

	default:
		return nil, errors.New("Neither file nor URL specified")

	}
}

func loadConfigFile(filename string) (*configfile, error) {
	if _, err := os.Stat(filename); err != nil {
		return nil, errors.Wrap(err, "Unable to access given file")
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open given file")
	}
	defer f.Close()

	out := &configfile{Template: defaultTemplate}
	if err = yaml.NewDecoder(f).Decode(out); err != nil {
		return nil, errors.Wrap(err, "Unable to parse given file")
	}

	if out.tpl, err = template.New("configTemplate").Funcs(template.FuncMap{
		"to_punycode": domainToPunycode,
	}).Parse(out.Template); err != nil {
		return nil, errors.Wrap(err, "Unable to parse given template")
	}

	return out, nil
}
