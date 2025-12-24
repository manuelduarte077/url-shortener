package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestMemoryRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		limit     int
		window    time.Duration
		requests  int
		wantAllow bool
	}{
		{
			name:      "allow requests within limit",
			limit:     5,
			window:    1 * time.Second,
			requests:  3,
			wantAllow: true,
		},
		{
			name:      "deny requests exceeding limit",
			limit:     5,
			window:    1 * time.Second,
			requests:  6,
			wantAllow: false,
		},
		{
			name:      "allow exactly limit requests",
			limit:     5,
			window:    1 * time.Second,
			requests:  5,
			wantAllow: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rl := NewMemoryRateLimiter(tt.limit, tt.window).(*MemoryRateLimiter)
			defer rl.Close()

			ctx := context.Background()
			identifier := "test-ip"

			var lastAllowed bool
			for i := 0; i < tt.requests; i++ {
				allowed, err := rl.Allow(ctx, identifier)
				if err != nil {
					t.Fatalf("Allow() error = %v", err)
				}
				lastAllowed = allowed
			}

			if lastAllowed != tt.wantAllow {
				t.Errorf("Allow() = %v, want %v", lastAllowed, tt.wantAllow)
			}
		})
	}
}

func TestMemoryRateLimiter_Allow_DifferentIdentifiers(t *testing.T) {
	t.Parallel()

	rl := NewMemoryRateLimiter(5, 1*time.Second).(*MemoryRateLimiter)
	defer rl.Close()

	ctx := context.Background()

	allowed1, err := rl.Allow(ctx, "ip1")
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if !allowed1 {
		t.Error("Allow() = false, want true for first request")
	}

	allowed2, err := rl.Allow(ctx, "ip2")
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if !allowed2 {
		t.Error("Allow() = false, want true for different identifier")
	}
}

func TestMemoryRateLimiter_Allow_WindowReset(t *testing.T) {
	t.Parallel()

	rl := NewMemoryRateLimiter(2, 100*time.Millisecond).(*MemoryRateLimiter)
	defer rl.Close()

	ctx := context.Background()
	identifier := "test-ip"

	// Exhaust the limit
	allowed1, _ := rl.Allow(ctx, identifier)
	allowed2, _ := rl.Allow(ctx, identifier)
	allowed3, _ := rl.Allow(ctx, identifier)

	if allowed1 != true || allowed2 != true {
		t.Error("First two requests should be allowed")
	}
	if allowed3 != false {
		t.Error("Third request should be denied")
	}

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again after window reset
	allowed4, err := rl.Allow(ctx, identifier)
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if !allowed4 {
		t.Error("Allow() = false, want true after window reset")
	}
}

func TestMemoryRateLimiter_Allow_ContextCancellation(t *testing.T) {
	t.Parallel()

	rl := NewMemoryRateLimiter(5, 1*time.Second).(*MemoryRateLimiter)
	defer rl.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should not panic with cancelled context
	_, err := rl.Allow(ctx, "test-ip")
	if err != nil {
		t.Logf("Allow() with cancelled context returned error (acceptable): %v", err)
	}
}

func TestMemoryRateLimiter_Cleanup(t *testing.T) {
	t.Parallel()

	rl := NewMemoryRateLimiter(5, 50*time.Millisecond).(*MemoryRateLimiter)
	ctx := context.Background()

	// Create a bucket
	rl.Allow(ctx, "test-ip")

	// Verify bucket exists
	rl.mu.RLock()
	if len(rl.buckets) == 0 {
		t.Error("Bucket should exist")
	}
	rl.mu.RUnlock()

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify bucket was cleaned up
	rl.mu.RLock()
	if len(rl.buckets) > 0 {
		t.Error("Bucket should be cleaned up after expiration")
	}
	rl.mu.RUnlock()

	rl.Close()
}

func TestMemoryRateLimiter_Close(t *testing.T) {
	t.Parallel()

	rl := NewMemoryRateLimiter(5, 1*time.Second).(*MemoryRateLimiter)

	// Close should not panic
	rl.Close()

	// Closing again should not panic
	rl.Close()
}
