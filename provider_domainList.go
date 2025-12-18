package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	registerProvider("domain-list", providerdomainList{})
}

type providerdomainList struct{}

func (providerdomainList) GetDomainList(d providerDefinition) ([]entry, error) {
	r, err := d.GetContent()
	if err != nil {
		return nil, fmt.Errorf("getting source content: %w", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			logrus.WithError(err).Error("closing domain-list")
		}
	}()

	logger := logrus.WithField("provider", d.Name)

	var entries []entry

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if lineIsComment(scanner.Text()) {
			continue
		}

		domain := strings.TrimSpace(strings.Split(scanner.Text(), "#")[0])

		if strings.Contains(domain, " ") {
			logger.WithField("line", scanner.Text()).Warn("invalid line found")
			continue
		}

		if isBlacklisted(domain) {
			logger.WithField("domain", domain).Debug("skipping because of blacklist")
			continue
		}

		entries = append(entries, entry{
			Domain:   domain,
			Comments: []string{d.Name},
		})
	}

	return entries, nil
}
