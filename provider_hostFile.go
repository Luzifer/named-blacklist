package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	registerProvider("hosts-file", providerHostFile{})
}

type providerHostFile struct{}

func (providerHostFile) GetDomainList(d providerDefinition) ([]entry, error) {
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

	var (
		entries []entry
		matcher = regexp.MustCompile(`^(?:[0-9.]+|[a-z0-9:]+)\s+([^\s]+)(?:\s+#(.+)|\s+#)?$`)
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if lineIsComment(line) {
			continue
		}

		if !matcher.MatchString(line) {
			logger.WithField("line", line).Warn("Invalid line found (format)")
			continue
		}

		groups := matcher.FindStringSubmatch(line)
		if len(groups) < 2 { //nolint:mnd
			logger.WithField("line", line).Warn("Invalid line found (groups)")
			continue
		}

		if isBlacklisted(groups[1]) {
			logger.WithField("domain", groups[1]).Debug("Skipping because of blacklist")
			continue
		}

		comment := fmt.Sprintf("%q", d.Name)
		if len(groups) == 3 && strings.Trim(groups[2], "#") != "" {
			comment = fmt.Sprintf("%s, Comment: %q",
				comment,
				strings.TrimSpace(strings.Trim(groups[2], "#")),
			)
		}

		entries = append(entries, entry{
			Domain:   groups[1],
			Comments: []string{comment},
		})
	}

	return entries, nil
}
