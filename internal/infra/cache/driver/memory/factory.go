package memory

import (
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/infra/cache/driver"
	"github.com/SkinonikS/discord-bot-go/internal/util/cfgutil"
)

type Config struct {
	MaxEntries int
}

type factoryImpl struct{}

func NewFactory() driver.Factory {
	return &factoryImpl{}
}

func (*factoryImpl) Create(_ string, cfg any) (driver.Driver, error) {
	typedCfg, err := cfgutil.CastConfig[*Config](cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to cast config: %w", err)
	}

	return &driverImpl{
		cache:      make(map[string]cacheItem),
		maxEntries: typedCfg.MaxEntries,
	}, nil
}

func (*factoryImpl) Name() string {
	return "memory"
}
