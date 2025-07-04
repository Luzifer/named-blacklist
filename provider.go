package main

import (
	"fmt"
	"sync"
)

var (
	providerRegistry     = map[providerType]provider{}
	providerRegistryLock sync.Mutex
)

type entry struct {
	Domain   string
	Comments []string
}

type provider interface {
	GetDomainList(providerDefinition) ([]entry, error)
}

func registerProvider(t providerType, p provider) {
	providerRegistryLock.Lock()
	defer providerRegistryLock.Unlock()

	providerRegistry[t] = p
}

func getDomainList(p providerDefinition) (entries []entry, err error) {
	pro, ok := providerRegistry[p.Type]
	if !ok {
		return nil, fmt.Errorf("unknown provider type %q", p.Type)
	}

	if entries, err = pro.GetDomainList(p); err != nil {
		return nil, fmt.Errorf("getting domain-list: %w", err)
	}

	return entries, nil
}
