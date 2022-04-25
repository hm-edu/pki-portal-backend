package helper

import (
	"errors"
	"fmt"
	"time"
)

// WaitFor polls the given function 'f', once every 'interval', up to 'timeout'.
func WaitFor(timeout, interval time.Duration, f func() (bool, error)) error {
	var lastErr error
	timeUp := time.After(timeout)
	for {
		select {
		case <-timeUp:
			if lastErr == nil {
				return errors.New("time limit exceeded")
			}
			return fmt.Errorf("time limit exceeded: last error: %w", lastErr)
		default:
		}

		stop, err := f()
		if stop {
			return nil
		}
		if err != nil {
			lastErr = err
		}

		time.Sleep(interval)
	}
}
