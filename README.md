# Circuit Breaker Demo

This repo documents an issue I am finding when using the [failsafe circuit breaker library](https://failsafe-go.dev/circuit-breaker/).

There seems to be a problem when transitioning from Open to Half-Open on a timeed delay.
