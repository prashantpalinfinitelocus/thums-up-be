package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests")
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	name             string
	maxRequests      uint32
	interval         time.Duration
	timeout          time.Duration
	failureThreshold uint32
	successThreshold uint32

	state           State
	counts          Counts
	expiry          time.Time
	lastStateChange time.Time
}

// Counts holds the statistics
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// Config holds circuit breaker configuration
type Config struct {
	Name             string
	MaxRequests      uint32        // Max requests allowed in half-open state
	Interval         time.Duration // Period for measuring failures
	Timeout          time.Duration // Time to stay in open state before trying half-open
	FailureThreshold uint32        // Failures needed to open circuit
	SuccessThreshold uint32        // Successes needed in half-open to close circuit
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:             cfg.Name,
		maxRequests:      cfg.MaxRequests,
		interval:         cfg.Interval,
		timeout:          cfg.Timeout,
		failureThreshold: cfg.FailureThreshold,
		successThreshold: cfg.SuccessThreshold,
		state:            StateClosed,
		expiry:           time.Now().Add(cfg.Interval),
		lastStateChange:  time.Now(),
	}

	return cb
}

// Execute wraps a function call with circuit breaker logic
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// Check if circuit allows the request
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn(ctx)

	// Record the result
	cb.afterRequest(err == nil)

	return err
}

func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.currentState(now)

	if state == StateOpen {
		return ErrCircuitOpen
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return ErrTooManyRequests
	}

	cb.counts.Requests++
	return nil
}

func (cb *CircuitBreaker) afterRequest(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.currentState(now)

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	if state == StateHalfOpen && cb.counts.ConsecutiveSuccesses >= cb.successThreshold {
		cb.setState(StateClosed, now)
		log.WithField("circuit_breaker", cb.name).Info("Circuit breaker closed")
	}
}

func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	if state == StateClosed && cb.counts.ConsecutiveFailures >= cb.failureThreshold {
		cb.setState(StateOpen, now)
		log.WithField("circuit_breaker", cb.name).Warn("Circuit breaker opened")
	} else if state == StateHalfOpen {
		cb.setState(StateOpen, now)
		log.WithField("circuit_breaker", cb.name).Warn("Circuit breaker reopened")
	}
}

func (cb *CircuitBreaker) currentState(now time.Time) State {
	switch cb.state {
	case StateClosed:
		if cb.expiry.Before(now) {
			cb.resetCounts(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
			log.WithField("circuit_breaker", cb.name).Info("Circuit breaker half-open")
		}
	}
	return cb.state
}

func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	cb.state = state
	cb.lastStateChange = now
	cb.resetCounts(now)

	switch state {
	case StateClosed:
		cb.expiry = now.Add(cb.interval)
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	case StateHalfOpen:
		cb.expiry = time.Time{}
	}
}

func (cb *CircuitBreaker) resetCounts(now time.Time) {
	cb.counts = Counts{}
	cb.expiry = now.Add(cb.interval)
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.currentState(time.Now())
}

// GetCounts returns a copy of current counts
func (cb *CircuitBreaker) GetCounts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.counts
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.setState(StateClosed, time.Now())
	log.WithField("circuit_breaker", cb.name).Info("Circuit breaker reset")
}
