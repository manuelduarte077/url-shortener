package repository

import (
	"context"
	"sync"
	"time"
	"url-shortener/internal/domain"
)

type MemoryURLRepository struct {
	mu            sync.RWMutex
	urls          map[string]*domain.URL
	ttl           time.Duration
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

func NewMemoryURLRepository(ttl time.Duration) domain.URLRepository {
	repo := &MemoryURLRepository{
		urls:        make(map[string]*domain.URL),
		ttl:         ttl,
		stopCleanup: make(chan struct{}),
	}

	repo.startCleanup()

	return repo
}

func (r *MemoryURLRepository) Save(ctx context.Context, url *domain.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.ttl > 0 {
		expiresAt := time.Now().Add(r.ttl)
		url.ExpiresAt = &expiresAt
	}

	r.urls[url.ShortCode] = url
	return nil
}

func (r *MemoryURLRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.urls[shortCode]
	if !exists {
		return nil, domain.ErrURLNotFound
	}

	return url, nil
}

func (r *MemoryURLRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.urls[shortCode]
	return exists, nil
}

func (r *MemoryURLRepository) startCleanup() {
	r.cleanupTicker = time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-r.cleanupTicker.C:
				r.cleanupExpired()
			case <-r.stopCleanup:
				return
			}
		}
	}()
}

func (r *MemoryURLRepository) cleanupExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for code, url := range r.urls {
		if url.IsExpired() {
			delete(r.urls, code)
		}
	}
	_ = now
}

func (r *MemoryURLRepository) Close() {
	if r.cleanupTicker != nil {
		r.cleanupTicker.Stop()
	}
	close(r.stopCleanup)
}
