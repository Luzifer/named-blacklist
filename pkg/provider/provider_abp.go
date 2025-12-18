package provider

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/Luzifer/named-blacklist/pkg/config"
	"github.com/Luzifer/named-blacklist/pkg/fqdn"
	"github.com/Luzifer/named-blacklist/pkg/helpers"
	"github.com/sirupsen/logrus"
)

func init() {
	registerProvider("adblock-plus", providerAdblockPlus{})
}

type providerAdblockPlus struct{}

func (providerAdblockPlus) GetDomainList(appVersion string, d config.ProviderDefinition) ([]Entry, error) {
	r, err := d.GetContent(appVersion)
	if err != nil {
		return nil, fmt.Errorf("getting source content: %w", err)
	}

	defer func() {
		if err := r.Close(); err != nil {
			logrus.WithError(err).Error("closing domain-list")
		}
	}()

	var (
		entries []Entry
		logger  = logrus.WithField("provider", d.Name)
		scanner = bufio.NewScanner(r)
	)

nextLine:
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if helpers.LineIsComment(line) {
			continue
		}

		switch {
		case strings.HasPrefix(line, "@@") && d.Action == config.ProviderActionBlacklist:
			// Whitelist-entry and blacklist-mode, skip that one
			logger.WithField("domain", line).Debug("skipping: wrong mode")
			continue nextLine

		case strings.HasPrefix(line, "||") && d.Action == config.ProviderActionWhitelist:
			// Blacklist-entry and whitelist-mode, skip that one
			logger.WithField("domain", line).Debug("skipping: wrong mode")
			continue nextLine

		case strings.HasPrefix(line, "|htt"):
			// We do not support that format
			logger.WithField("domain", line).Debug("skipping: unsupported format, schema")
			continue nextLine

		case !strings.HasSuffix(line, "^"):
			// Propably optioned rule, we don't support that
			logger.WithField("domain", line).Debug("skipping: unsupported format, options")
			continue nextLine
		}

		// Now sanitize the entry
		domain := strings.TrimSuffix(line, "^")
		domain = strings.TrimPrefix(domain, "@@")
		domain = strings.TrimPrefix(domain, "||")

		if !fqdn.IsValidEntry(domain) {
			logger.WithField("domain", domain).Debug("skipping: not a valid domain")
			continue nextLine
		}

		entries = append(entries, Entry{
			Domain:   domain,
			Comments: []string{d.Name},
		})
	}

	return entries, nil
}
