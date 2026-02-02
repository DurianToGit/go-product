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

	failures         int
	maxFailures      int
	openTimeout      time.Duration
	lastOpenTime     time.Time
	halfOpenInFlight bool
}

func New(maxFailures int, openTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		openTimeout: openTimeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	// 1) 进入前状态检查
	cb.mu.Lock()

	now := time.Now()

	switch cb.state {
	case StateOpen:
		// Open: 冷却期未到 -> 直接拒绝
		if now.Sub(cb.lastOpenTime) < cb.openTimeout {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
		// 冷却期到 -> 进入 HalfOpen（准备试探）
		cb.state = StateHalfOpen
		cb.halfOpenInFlight = false
	case StateHalfOpen:
		// HalfOpen: 只允许一个请求试探
		if cb.halfOpenInFlight {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
		cb.halfOpenInFlight = true
	case StateClosed:
		// 正常放行
	}

	cb.mu.Unlock()

	// 2) 执行真实下游调用（锁外执行）
	err := fn()

	// 3) 执行后更新状态
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		// HalfOpen: 试探结束，释放 inFlight
		cb.halfOpenInFlight = false

		if err != nil {
			// 试探失败：回到 Open，重新开始冷却
			cb.failures = cb.maxFailures
			cb.state = StateOpen
			cb.lastOpenTime = time.Now()
			return err
		}

		// 试探成功：恢复 Closed
		cb.failures = 0
		cb.state = StateClosed
		return nil
	}

	// Closed/Open 常规统计
	if err != nil {
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			cb.lastOpenTime = time.Now()
		}
		return err
	}

	// 成功：清零失败
	cb.failures = 0
	cb.state = StateClosed
	return nil
}
