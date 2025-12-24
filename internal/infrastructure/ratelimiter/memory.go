package ratelimiter

import (
	"context"
	"sync"
	"time"
	"url-shortener/internal/domain"
)

type MemoryRateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*bucket
	limit    int
	window   time.Duration
	cleanup  *time.Ticker
	stopChan chan struct{}
}

type bucket struct {
	count     int
	resetTime time.Time
}

func NewMemoryRateLimiter(limit int, window time.Duration) domain.RateLimiter {
	rl := &MemoryRateLimiter{
		buckets:  make(map[string]*bucket),
		limit:    limit,
		window:   window,
		stopChan: make(chan struct{}),
	}

	rl.cleanup = time.NewTicker(window)
	go rl.cleanupExpired()

	return rl
}

func (rl *MemoryRateLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[identifier]

	if !exists || now.After(b.resetTime) {
		rl.buckets[identifier] = &bucket{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true, nil
	}

	if b.count >= rl.limit {
		return false, nil
	}

	b.count++
	return true, nil
}

func (rl *MemoryRateLimiter) cleanupExpired() {
	for {
		select {
		case <-rl.cleanup.C:
			rl.mu.Lock()
			now := time.Now()
			for id, b := range rl.buckets {
				if now.After(b.resetTime) {
					delete(rl.buckets, id)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopChan:
			return
		}
	}
}

func (rl *MemoryRateLimiter) Close() {
	if rl.cleanup != nil {
		rl.cleanup.Stop()
	}
	select {
	case <-rl.stopChan:
	default:
		close(rl.stopChan)
	}
}
