package middleware

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sony/gobreaker"
)

// breakers holds a map of circuit breakers, one for each service name.
var (
	breakers     = make(map[string]*gobreaker.CircuitBreaker)
	breakersLock sync.RWMutex
)

// getBreaker retrieves or creates a circuit breaker for a given service name.
func getBreaker(name string) *gobreaker.CircuitBreaker {
	breakersLock.RLock()
	breaker, exists := breakers[name]
	breakersLock.RUnlock()

	if !exists {
		breakersLock.Lock()
		defer breakersLock.Unlock()
		// Double-check in case another goroutine created it while we were waiting for the lock.
		breaker, exists = breakers[name]
		if !exists {
			st := gobreaker.Settings{
				Name:        name,
				MaxRequests: 3,                // Number of requests to probe in half-open state.
				Interval:    0,                // The clearing of counts is done by Timeout.
				Timeout:     30 * time.Second, // Time the breaker remains open before going to half-open.
				ReadyToTrip: func(counts gobreaker.Counts) bool {
					// Open the circuit if we have 5 or more consecutive failures.
					return counts.ConsecutiveFailures > 5
				},
				OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
					log.Printf("Circuit Breaker '%s' changed state from %s to %s\n", name, from, to)
				},
			}
			breaker = gobreaker.NewCircuitBreaker(st)
			breakers[name] = breaker
			log.Printf("Created new Circuit Breaker for service: %s", name)
		}
	}
	return breaker
}

// CircuitBreaker is a middleware that wraps the request with a circuit breaker.
func CircuitBreaker(next http.Handler, serviceName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		breaker := getBreaker(serviceName)

		_, err := breaker.Execute(func() (any, error) {
			// To detect failures, we need to capture the status code of the response.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			// Consider server-side errors (5xx) as failures.
			// Client-side errors (4xx) are not the service's fault, so they are successes.
			if ww.Status() >= http.StatusInternalServerError {
				// Return a specific error to signal a failure to the circuit breaker.
				return nil, fmt.Errorf("server error: status %d", ww.Status())
			}

			// Signal success to the circuit breaker.
			return nil, nil
		})

		// If err is not nil, it means the circuit is open or the request failed.
		if err != nil {
			// When the circuit is open, immediately return a 503 Service Unavailable.
			log.Printf("Circuit breaker open for service '%s', blocking request. Error: %v", serviceName, err)
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		}
	})
}
