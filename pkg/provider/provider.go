package provider

import (
	"fmt"
	"sync"

	"github.com/Luzifer/named-blacklist/pkg/config"
)

var (
	providerRegistry     = map[config.ProviderType]Provider{}
	providerRegistryLock sync.Mutex
)

type (
	// Entry represents an entry of the black-/whitelist including
	// comments where it was found
	Entry struct {
		Domain   string
		Comments []string
	}

	// Provider represents a source of domain Entries
	Provider interface {
		GetDomainList(appVersion string, pd config.ProviderDefinition) ([]Entry, error)
	}
)

// GetDomainList executes the provider given through the passed definition
func GetDomainList(appVersion string, p config.ProviderDefinition) (entries []Entry, err error) {
	pro, ok := providerRegistry[p.Type]
	if !ok {
		return nil, fmt.Errorf("unknown provider type %q", p.Type)
	}

	if entries, err = pro.GetDomainList(appVersion, p); err != nil {
		return nil, fmt.Errorf("getting domain-list: %w", err)
	}

	return entries, nil
}

func registerProvider(t config.ProviderType, p Provider) {
	providerRegistryLock.Lock()
	defer providerRegistryLock.Unlock()

	providerRegistry[t] = p
}
