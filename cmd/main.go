package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/failsafe-go/failsafe-go/circuitbreaker"
)

type Breaker struct {
	breakers         map[string]circuitbreaker.CircuitBreaker[*http.Response]
	log              *slog.Logger
	FailureThreshold uint
	Delay            time.Duration
	SuccessThreshold uint
}

type Option func(*Breaker)

func WithFailureThreshold(threshold uint) Option {
	return func(b *Breaker) {
		b.FailureThreshold = threshold
	}
}
func WithDelay(delay time.Duration) Option {
	return func(b *Breaker) {
		b.Delay = delay
	}
}
func WithSuccessThreshold(threshold uint) Option {
	return func(b *Breaker) {
		b.SuccessThreshold = threshold
	}
}

func New(log *slog.Logger, opts ...Option) *Breaker {
	b := &Breaker{
		log:              log,
		breakers:         make(map[string]circuitbreaker.CircuitBreaker[*http.Response]),
		SuccessThreshold: 2,
		FailureThreshold: 5,
		Delay:            time.Minute,
	}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

//nolint:bodyclose // this is not meant to be closed here
func (b *Breaker) Upsert(key string) circuitbreaker.CircuitBreaker[*http.Response] {
	breaker, exists := b.breakers[key]
	if exists {
		b.log.Debug("circuit breaker already exists for key", slog.String("key", key))
		return breaker
	}

	// Status codes we would like to tigger a circuit breaker with
	codes := []int{429, 500}

	newBreaker := circuitbreaker.Builder[*http.Response]().
		HandleIf(func(response *http.Response, _ error) bool {
			return response != nil && slices.Contains(codes, response.StatusCode)
		}).
		WithFailureThreshold(b.FailureThreshold).
		WithDelay(b.Delay).
		WithSuccessThreshold(b.SuccessThreshold).
		OnStateChanged(func(event circuitbreaker.StateChangedEvent) {
			message := fmt.Sprintf("circuit breaker for key %s state changed from %s to %s", key, event.OldState.String(), event.NewState.String())
			b.log.Debug(message, slog.String("event", fmt.Sprintf("%v", event)))
		}).
		Build()
	b.breakers[key] = newBreaker

	return newBreaker
}
