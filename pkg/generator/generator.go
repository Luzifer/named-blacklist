package generator

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"sync"

	"github.com/Luzifer/named-blacklist/pkg/config"
	"github.com/Luzifer/named-blacklist/pkg/provider"
	"github.com/sirupsen/logrus"
)

// GenerateBlacklist takes a collection of providers and compiles their
// content into a single list of blacklisted domains
func GenerateBlacklist(appVersion string, providers []config.ProviderDefinition) (blacklist []provider.Entry, err error) {
	var (
		errs      []error
		whitelist []provider.Entry
		write     = new(sync.Mutex)
		wg        sync.WaitGroup
	)

	wg.Add(len(providers))
	for _, p := range providers {
		go func(p config.ProviderDefinition) {
			defer wg.Done()

			logger := logrus.WithField("provider", p.Name)
			logger.Info("starting domain list extraction")

			entries, err := provider.GetDomainList(appVersion, p)
			if err != nil {
				errs = append(errs, fmt.Errorf("getting domain list for %q: %w", p.Name, err))
				return
			}

			write.Lock()
			defer write.Unlock()

			switch p.Action {
			case config.ProviderActionBlacklist:
				blacklist = append(blacklist, entries...)

			case config.ProviderActionWhitelist:
				whitelist = append(whitelist, entries...)

			default:
				errs = append(errs, fmt.Errorf("invalid action for name %q: %s", p.Name, p.Action))
				return
			}

			logger.WithField("no_entries", len(entries)).Info("extraction complete")
		}(p)
	}

	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("collecting entries: %w", errors.Join(errs...))
	}

	logrus.Info("Removing duplicates...")
	blacklist = removeDuplicateEntries(blacklist)
	whitelist = removeDuplicateEntries(whitelist)
	logrus.Info("Done")

	blacklist = slices.DeleteFunc(blacklist, func(be provider.Entry) bool {
		return slices.ContainsFunc(whitelist, func(we provider.Entry) bool { return we.Domain == be.Domain })
	})

	sort.Slice(blacklist, func(i, j int) bool { return blacklist[i].Domain < blacklist[j].Domain })

	return blacklist, nil
}

func removeDuplicateEntries(list []provider.Entry) (unique []provider.Entry) {
	keys := make(map[string]int)

	for _, e := range list {
		i, contains := keys[e.Domain]
		if contains {
			unique[i].Comments = sort.StringSlice(append(unique[i].Comments, e.Comments...))
			continue
		}

		// store index for domain
		keys[e.Domain] = len(unique)
		unique = append(unique, e)
	}

	return unique
}
