package main

import (
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_CircuitBreaker(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	cb := New(
		log,
		WithSuccessThreshold(1),
		WithFailureThreshold(1),
		WithDelay(time.Second*3),
	)

	// Create two circuit breakers with different keys
	breaker1 := cb.Upsert("key1")
	breaker2 := cb.Upsert("key2")

	// Assert that both breakers are initially closed
	assert.True(t, breaker1.IsClosed(), "breaker1 should not be closed")
	assert.True(t, breaker2.IsClosed(), "breaker2 should not be closed")

	// Simulate a failed request for breaker1
	breaker1.RecordResult(&http.Response{StatusCode: 500})
	breaker2.RecordResult(&http.Response{StatusCode: 200})

	log.Info("Failures for breaker1", "failures", breaker1.Metrics().Failures())

	// Assert that breaker1 is locked
	assert.True(t, breaker1.IsOpen(), "breaker1 should be closed")
	// Assert that breaker2 is not affected
	assert.True(t, breaker2.IsClosed(), "breaker2 should not be closed")

	// sleep until breaker delay trips half open state. Added 10 for a reasonable buffer time.
	time.Sleep(breaker1.RemainingDelay() + 10)

	// Works if I uncomment this line
	// breaker1.HalfOpen()

	// Assert that breaker1 is half open after delay
	assert.True(t, breaker1.IsHalfOpen(), "breaker1 should be half open")

	// Simulate a successful request for breaker1 so we transition to closed
	breaker1.RecordResult(&http.Response{StatusCode: 200})

	// Assert that breaker1 is unlocked
	assert.True(t, breaker1.IsClosed(), "breaker1 should not be closed")
}
