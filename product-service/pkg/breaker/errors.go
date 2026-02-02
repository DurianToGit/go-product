package breaker

import "errors"

func IsCircuitOpen(err error) bool {
	return errors.Is(err, ErrCircuitOpen)
}
