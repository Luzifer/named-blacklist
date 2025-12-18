package provider

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/Luzifer/named-blacklist/pkg/config"
	"github.com/Luzifer/named-blacklist/pkg/helpers"
	"github.com/sirupsen/logrus"
)

func init() {
	registerProvider("hosts-file", providerHostFile{})
}

type providerHostFile struct{}

func (providerHostFile) GetDomainList(appVersion string, d config.ProviderDefinition) ([]Entry, error) {
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

	var (
		entries []Entry
		matcher = regexp.MustCompile(`^(?:[0-9.]+|[a-z0-9:]+)\s+([^\s]+)(?:\s+#(.+)|\s+#)?$`)
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if helpers.LineIsComment(line) {
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

		if helpers.IsBlacklisted(groups[1]) {
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

		entries = append(entries, Entry{
			Domain:   groups[1],
			Comments: []string{comment},
		})
	}

	return entries, nil
}
