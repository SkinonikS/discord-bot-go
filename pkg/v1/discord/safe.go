package discord

func ListenWithError(cb func() error) error {
	return cb()
}
