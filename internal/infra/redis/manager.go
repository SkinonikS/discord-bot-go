package redis

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Manager interface {
	Connection(name string) (*redis.Client, error)
}

type managerImpl struct {
	mu          sync.RWMutex
	config      *Config
	connections map[string]*redis.Client
}

type Params struct {
	fx.In

	Lc     fx.Lifecycle
	Log    *zap.Logger
	Config *Config
}

func NewManager(p Params) Manager {
	manager := &managerImpl{
		config:      p.Config,
		connections: make(map[string]*redis.Client, len(p.Config.Connections)),
	}

	p.Lc.Append(fx.StartStopHook(func(ctx context.Context) error {
		go func() {
			for name := range p.Config.Connections {
				connection, err := manager.Connection(name)
				if err != nil {
					p.Log.Info("failed to establish connection", zap.String("name", name), zap.Error(err))
					return
				}

				if err := connection.Ping(ctx).Err(); err != nil {
					p.Log.Error("redis ping failed", zap.String("name", name), zap.Error(err))
					return
				}

				p.Log.Info("connection established", zap.String("name", name))
			}
		}()

		return nil
	}, func(_ context.Context) error {
		for _, conn := range manager.connections {
			_ = conn.Close()
		}

		p.Log.Info("closed all connections")
		return nil
	}))

	return manager
}

func (m *managerImpl) Connection(name string) (*redis.Client, error) {
	m.mu.RLock()
	if client, ok := m.connections[name]; ok {
		m.mu.RUnlock()
		return client, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	client, err := m.resolve(name)
	if err != nil {
		return nil, err
	}

	m.connections[name] = client
	return client, nil
}

func (m *managerImpl) resolve(name string) (*redis.Client, error) {
	config, ok := m.config.Connections[name]
	if !ok {
		return nil, fmt.Errorf("failed to resolve redis config: %s", name)
	}

	return redis.NewClient(&redis.Options{
		Addr:       net.JoinHostPort(m.config.Host, strconv.Itoa(m.config.Port)),
		Username:   m.config.Username,
		Password:   m.config.Password,
		DB:         config.DB,
		ClientName: name,
	}), nil
}
