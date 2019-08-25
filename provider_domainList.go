package main

import (
	"bufio"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	registerProvider("domain-list", providerdomainList{})
}

type providerdomainList struct{}

func (p providerdomainList) GetDomainList(d providerDefinition) ([]entry, error) {
	r, err := d.GetContent()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get source content")
	}
	defer r.Close()

	logger := log.WithField("provider", d.Name)

	var entries []entry

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if lineIsComment(scanner.Text()) {
			continue
		}

		domain := strings.TrimSpace(scanner.Text())

		if isBlacklisted(domain) {
			logger.WithField("domain", domain).Debug("Skipping because of blacklist")
			continue
		}

		entries = append(entries, entry{
			Domain:  domain,
			Comment: d.Name,
		})
	}

	return entries, nil
}
