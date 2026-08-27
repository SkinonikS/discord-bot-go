package memory

import (
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/infra/cache/driver"
)

type cacheItem struct {
	value          []byte
	expiresAt      time.Time
	lastAccessedAt time.Time
}

type driverImpl struct {
	cache      map[string]cacheItem
	maxEntries int
}

func (d *driverImpl) Put(put driver.Put) error {
	_, exists := d.cache[put.Key]

	if d.maxEntries > 0 && !exists && len(d.cache) >= d.maxEntries {
		d.deleteLeastRecentlyUsed()
	}

	d.cache[put.Key] = cacheItem{
		value:          put.Value,
		expiresAt:      put.ExpiresAt,
		lastAccessedAt: time.Now(),
	}

	return nil
}

func (d *driverImpl) Find(key string) (driver.Result, error) {
	value, ok := d.cache[key]
	if !ok {
		return driver.Result{}, nil
	}

	if !value.expiresAt.IsZero() && !time.Now().Before(value.expiresAt) {
		delete(d.cache, key)

		return driver.Result{}, nil
	}

	value.lastAccessedAt = time.Now()
	d.cache[key] = value

	returnValue := make([]byte, len(value.value))
	copy(returnValue, value.value)

	return driver.Result{
		IsExists:  true,
		Value:     returnValue,
		ExpiresAt: value.expiresAt,
	}, nil
}

func (d *driverImpl) deleteLeastRecentlyUsed() {
	var oldestKey string
	var oldestAccess time.Time

	for key, item := range d.cache {
		if oldestKey == "" || item.lastAccessedAt.Before(oldestAccess) {
			oldestKey = key
			oldestAccess = item.lastAccessedAt
		}
	}

	if oldestKey != "" {
		delete(d.cache, oldestKey)
	}
}
