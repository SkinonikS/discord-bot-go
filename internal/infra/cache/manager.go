package cache

import (
	"fmt"
	"sync"

	"github.com/SkinonikS/discord-bot-go/internal/infra/cache/driver"
	"go.uber.org/fx"
)

type Manager interface{}

type managerImpl struct {
	mu        sync.RWMutex
	config    *Config
	resolved  map[string]driver.Driver
	factories map[string]driver.Factory
}

type ManagerParams struct {
	fx.In

	Factories []driver.Factory `group:"cache_driver_factories"`
	Config    *Config
}

func NewManager(p ManagerParams) (Manager, error) {
	manager := &managerImpl{
		factories: make(map[string]driver.Factory, len(p.Factories)),
		config:    p.Config,
	}

	for _, factory := range p.Factories {
		if _, ok := manager.factories[factory.Name()]; ok {
			return nil, fmt.Errorf("duplicate cache driver factory name: %s", factory.Name())
		}
		manager.factories[factory.Name()] = factory
	}

	return manager, nil
}

func (m *managerImpl) Store(name string) (driver.Driver, error) {
	m.mu.RLock()
	if store, ok := m.resolved[name]; ok {
		m.mu.RUnlock()
		return store, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	store, err := m.resolveStore(name)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve store: %w", err)
	}

	m.resolved[name] = store
	return store, nil
}

func (m *managerImpl) resolveStore(name string) (driver.Driver, error) {
	storeConfig, ok := m.config.Stores[name]
	if !ok {
		return nil, fmt.Errorf("store %s not found", name)
	}

	storeFactory, ok := m.factories[storeConfig.Driver]
	if !ok {
		return nil, fmt.Errorf("store driver %s not found", storeConfig.Driver)
	}

	return storeFactory.Create(name, storeConfig)
}
