package readiness

type Handler interface {
	IsReady() bool
	IsHealthy() bool
}
