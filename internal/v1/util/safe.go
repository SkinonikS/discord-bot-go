package util

import (
	"fmt"
)

func Safe(fn func() error) (err error) {
	defer func() {
		if rvr := recover(); rvr != nil {
			err = fmt.Errorf("panic: %v", rvr)
		}
	}()
	return fn()
}
