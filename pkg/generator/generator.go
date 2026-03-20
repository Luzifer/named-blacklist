package generator

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/Luzifer/named-blacklist/pkg/config"
	"github.com/Luzifer/named-blacklist/pkg/provider"
	"github.com/sirupsen/logrus"
)

type (
	blacklistAggregate struct {
		comments          []string
		matchingProviders int
		requiredMatches   int
	}

	providerResult struct {
		index    int
		provider config.ProviderDefinition
		entries  []provider.Entry
	}
)

// GenerateBlacklist takes a collection of providers and compiles their
// content into a single list of blacklisted domains
func GenerateBlacklist(appVersion string, providers []config.ProviderDefinition) (blacklist []provider.Entry, err error) {
	var (
		errs    []error
		results = make([]providerResult, len(providers))
		write   = new(sync.Mutex)
		wg      sync.WaitGroup
	)

	for _, p := range providers {
		switch p.Action {
		case config.ProviderActionBlacklist, config.ProviderActionWhitelist:
		default:
			errs = append(errs, fmt.Errorf("invalid action for name %q: %s", p.Name, p.Action))
		}

		if p.MinMatches < 0 {
			errs = append(errs, fmt.Errorf("invalid min_matches for name %q: %d", p.Name, p.MinMatches))
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("collecting entries: %w", errors.Join(errs...))
	}

	wg.Add(len(providers))
	for i, p := range providers {
		go func(i int, p config.ProviderDefinition) {
			defer wg.Done()

			logger := logrus.WithField("provider", p.Name)
			logger.Info("starting domain list extraction")

			entries, err := provider.GetDomainList(appVersion, p)
			if err != nil {
				write.Lock()
				errs = append(errs, fmt.Errorf("getting domain list for %q: %w", p.Name, err))
				write.Unlock()
				return
			}

			write.Lock()
			results[i] = providerResult{
				index:    i,
				provider: p,
				entries:  removeDuplicateEntries(entries),
			}
			write.Unlock()

			logger.WithField("no_entries", len(entries)).Info("extraction complete")
		}(i, p)
	}

	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("collecting entries: %w", errors.Join(errs...))
	}

	blacklist = compileBlacklist(results)
	sort.Slice(blacklist, func(i, j int) bool { return blacklist[i].Domain < blacklist[j].Domain })

	return blacklist, nil
}

func compileBlacklist(results []providerResult) (blacklist []provider.Entry) {
	logrus.Info("compiling final blacklist...")

	blacklistEntries := make(map[string]*blacklistAggregate)
	whitelistDomains := make(map[string]struct{})

	for _, result := range results {
		switch result.provider.Action {
		case config.ProviderActionBlacklist:
			for _, entry := range result.entries {
				aggregate, ok := blacklistEntries[entry.Domain]
				if !ok {
					aggregate = &blacklistAggregate{
						requiredMatches: effectiveMinMatches(result.provider),
					}
					blacklistEntries[entry.Domain] = aggregate
				}

				aggregate.matchingProviders++
				aggregate.requiredMatches = min(aggregate.requiredMatches, effectiveMinMatches(result.provider))
				aggregate.comments = mergeCommentsUnique(aggregate.comments, entry.Comments)
			}

		case config.ProviderActionWhitelist:
			for _, entry := range result.entries {
				whitelistDomains[entry.Domain] = struct{}{}
			}

		default:
			logrus.WithFields(logrus.Fields{
				"provider": result.provider.Name,
				"action":   result.provider.Action,
			}).Warn("skipping provider with invalid action")
		}
	}

	for domain, aggregate := range blacklistEntries {
		if aggregate.matchingProviders < aggregate.requiredMatches {
			continue
		}

		if _, ok := whitelistDomains[domain]; ok {
			continue
		}

		blacklist = append(blacklist, provider.Entry{
			Domain:   domain,
			Comments: aggregate.comments,
		})
	}

	logrus.Info("done")

	return blacklist
}

func effectiveMinMatches(p config.ProviderDefinition) int {
	if p.MinMatches == 0 {
		return 1
	}

	return p.MinMatches
}

func mergeCommentsUnique(existing, incoming []string) []string {
	seen := make(map[string]struct{}, len(existing))

	for _, comment := range existing {
		seen[comment] = struct{}{}
	}

	for _, comment := range incoming {
		if _, ok := seen[comment]; ok {
			continue
		}

		existing = append(existing, comment)
		seen[comment] = struct{}{}
	}

	return existing
}

func removeDuplicateEntries(list []provider.Entry) (unique []provider.Entry) {
	keys := make(map[string]int)

	for _, e := range list {
		i, contains := keys[e.Domain]
		if contains {
			unique[i].Comments = mergeCommentsUnique(unique[i].Comments, e.Comments)
			continue
		}

		// store index for domain
		keys[e.Domain] = len(unique)
		unique = append(unique, e)
	}

	return unique
}
