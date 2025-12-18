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
	registerProvider("domain-list", providerdomainList{})
}

type providerdomainList struct{}

func (providerdomainList) GetDomainList(appVersion string, d config.ProviderDefinition) ([]Entry, error) {
	r, err := d.GetContent(appVersion)
	if err != nil {
		return nil, fmt.Errorf("getting source content: %w", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			logrus.WithError(err).Error("closing domain-list")
		}
	}()

	logger := logrus.WithField("provider", d.Name)

	var entries []Entry

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if helpers.LineIsComment(scanner.Text()) {
			continue
		}

		domain := strings.TrimSpace(strings.Split(scanner.Text(), "#")[0])

		if strings.Contains(domain, " ") {
			logger.WithField("line", scanner.Text()).Warn("invalid line found")
			continue
		}

		if helpers.IsBlacklisted(domain) {
			logger.WithField("domain", domain).Debug("skipping because of blacklist")
			continue
		}

		if !fqdn.IsValidEntry(domain) {
			logger.WithField("domain", domain).Debug("skipping because not a valid domain")
			continue
		}

		entries = append(entries, Entry{
			Domain:   domain,
			Comments: []string{d.Name},
		})
	}

	return entries, nil
}
