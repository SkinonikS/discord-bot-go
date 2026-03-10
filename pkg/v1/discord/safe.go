package discord

import "fmt"

func ListenWithError(cb func() error) (err error) {
	defer func() {
		if rvr := recover(); rvr != nil {
			err = fmt.Errorf("%v", rvr)
		}
	}()

	return cb()
}
