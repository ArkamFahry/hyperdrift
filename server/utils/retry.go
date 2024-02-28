package utils

import (
	"context"
	"math"
	"math/rand"
	"time"
)

func Retry(ctx context.Context, fn func() error, maxRetries int) error {
	const baseDelay = 100 * time.Millisecond
	const maxDelay = 5 * time.Second

	backoff := baseDelay
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := fn(); err == nil {
			return nil
		}

		backoff = time.Duration(math.Min(float64(baseDelay)*math.Pow(2, float64(i)), float64(maxDelay)))
		time.Sleep(backoff + time.Duration(rand.Intn(100))*time.Millisecond)
	}

	return fn()
}
