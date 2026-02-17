/*
Package retry provides a simple retry mechanism with exponential backoff.
It is as abstract as possible to allow for different retry strategies.
*/
package retry

import (
	"context"
	"log/slog"
	"time"
)

type Config struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// DefaultConfig returns a Config with sensible default values
func DefaultConfig() Config {
	return Config{
		MaxRetries:  3,
		InitialWait: 1 * time.Second,
		MaxWait:     10 * time.Second,
	}
}

// WithBackoff executes the given operation with exponential backoff retry logic
func WithBackoff(ctx context.Context, cfg Config, operation func() error) error {
	var err error
	wait := cfg.InitialWait

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			slog.Debug("retry", "retry", attempt, "from", cfg.MaxRetries, "duration", wait)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}

			// Exponential backoff with max wait cap
			wait *= 2
			if wait > cfg.MaxWait {
				wait = cfg.MaxWait
			}
		}

		if err = operation(); err == nil {
			return nil
		}

		slog.Debug("retry failed", "error", err, "retry", attempt+1, "from", cfg.MaxRetries, "duration", wait)
	}

	return err
}
