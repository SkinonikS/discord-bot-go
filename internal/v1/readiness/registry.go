package readiness

import (
	"github.com/samber/lo"
	"go.uber.org/fx"
)

type Registry interface {
	IsReady() bool
	IsHealthy() bool
}

type registryImpl struct {
	handlers []Handler
}

type HandlerParams struct {
	fx.In

	Handlers []Handler `group:"readiness_handlers"`
}

func NewRegistry(p HandlerParams) Registry {
	return &registryImpl{
		handlers: p.Handlers,
	}
}

func (r *registryImpl) IsReady() bool {
	return lo.SomeBy(r.handlers, func(h Handler) bool {
		return h.IsReady()
	})
}

func (r *registryImpl) IsHealthy() bool {
	return lo.SomeBy(r.handlers, func(h Handler) bool {
		return h.IsHealthy()
	})
}
