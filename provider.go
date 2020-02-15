package main

import (
	"sync"

	"github.com/pkg/errors"
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

func getDomainList(p providerDefinition) ([]entry, error) {
	pro, ok := providerRegistry[p.Type]
	if !ok {
		return nil, errors.Errorf("Unknown provider type %q", p.Type)
	}

	return pro.GetDomainList(p)
}
