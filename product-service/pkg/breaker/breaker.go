package breaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu sync.Mutex

	state State

	failures     int
	maxFailures  int
	openTimeout  time.Duration
	lastOpenTime time.Time
}

func New(maxFailures int, openTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		openTimeout: openTimeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	// Open 状态直接拒绝
	if cb.state == StateOpen {
		if time.Since(cb.lastOpenTime) > cb.openTimeout {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}

	cb.mu.Unlock()

	// 真正执行下游调用
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			cb.lastOpenTime = time.Now()
		}
		return err
	}

	// 成功，恢复
	cb.failures = 0
	cb.state = StateClosed
	return nil
}
