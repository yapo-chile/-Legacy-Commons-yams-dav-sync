package infrastructure

import (
	"fmt"

	"github.com/sony/gobreaker"
)

var circuitBreaker *gobreaker.CircuitBreaker

func init() {
	var st gobreaker.Settings
	fmt.Printf("\nCIRCUIT BREAKER ENABLED")
	st.Name = "HTTP SEND"
	st.ReadyToTrip = func(counts gobreaker.Counts) bool {
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 3 && failureRatio >= 0.6
	}
	st.Timeout = 30
	circuitBreaker = gobreaker.NewCircuitBreaker(st)
}
