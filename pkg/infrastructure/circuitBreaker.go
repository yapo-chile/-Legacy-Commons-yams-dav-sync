package infrastructure

import (
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker common variables
var (
	// StateClosed represents Circuit breaker closed state
	StateClosed = gobreaker.StateClosed

	// StateHalfOpen represents Circuit breaker halft-open state
	StateHalfOpen = gobreaker.StateHalfOpen

	// StateOpen represents Circuit breaker open state
	StateOpen = gobreaker.StateOpen

	// ErrTooManyRequests is returned when the CB state is half open and the requests count is over the cb maxRequests
	ErrTooManyRequests = gobreaker.ErrTooManyRequests

	// ErrOpenState is returned when the CB state is open
	ErrOpenState = gobreaker.ErrOpenState
)

// NewCircuitBreaker initializes circuit breaker wrapper
// name is the circuit breaker
// consecutiveFailures is the maximum of consecutive errors allowed before open state
// failureRatio is the maximum error ratio (errors vs requests qty) allowed before open state
// Interval is the cyclic period of the closed state for the CircuitBreaker to clear the internal Counts.
// If Interval is 0, the CircuitBreaker doesn't clear internal Counts during the closed state.
// Timeout is the period of the open state, after which the state of the CircuitBreaker becomes half-open.
func NewCircuitBreaker(name string, consecutiveFailures uint32, failureRatio float64, timeout, interval int) CircuitBreaker {
	settings := gobreaker.Settings{
		Name:     name,
		Timeout:  time.Duration(timeout) * (time.Second),
		Interval: time.Duration(interval) * (time.Second),

		// If ReadyToTrip returns true, the CircuitBreaker will be placed into the open state
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			errorRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return errorRatio >= failureRatio || counts.ConsecutiveFailures > consecutiveFailures
		},
	}

	return gobreaker.NewCircuitBreaker(settings)
}

// CircuitBreaker allows circuit breaker operations
type CircuitBreaker interface {
	// Execute wrapps a function. If the function returns too many errors, circuit breaker
	// will return "circuit breaker open" error
	Execute(req func() (interface{}, error)) (interface{}, error)
	// Name returns circuit breaker name
	Name() string
	// State returns the current status of the circuit breaker
	State() gobreaker.State
}
