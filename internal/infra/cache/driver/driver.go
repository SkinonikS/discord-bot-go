package driver

import (
	"time"
)

type Put struct {
	Key       string
	Value     []byte
	ExpiresAt time.Time
}

type Result struct {
	IsExists  bool
	Value     []byte
	ExpiresAt time.Time
}

type Driver interface {
	Put(put Put) error
	Find(key string) (Result, error)
}

type Factory interface {
	Create(name string, cfg any) (Driver, error)
	Name() string
}
