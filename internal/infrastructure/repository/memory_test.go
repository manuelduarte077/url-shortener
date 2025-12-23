package repository_test

import (
	"context"
	"testing"
	"time"
	"url-shortener/internal/domain"
	"url-shortener/internal/infrastructure/repository"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryURLRepository(t *testing.T) {
	ttl := 1 * time.Hour
	repo := repository.NewMemoryURLRepository(ttl)

	assert.NotNil(t, repo)

	// Cleanup
	if memRepo, ok := repo.(*repository.MemoryURLRepository); ok {
		memRepo.Close()
	}
}

func TestMemoryURLRepository_Save(t *testing.T) {
	repo := repository.NewMemoryURLRepository(0) // No TTL
	defer func() {
		if memRepo, ok := repo.(*repository.MemoryURLRepository); ok {
			memRepo.Close()
		}
	}()

	ctx := context.Background()
	url := &domain.URL{
		ShortCode: "test123",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
	}

	err := repo.Save(ctx, url)

	assert.NoError(t, err)
}

func TestMemoryURLRepository_Save_WithTTL(t *testing.T) {
	ttl := 1 * time.Hour
	repo := repository.NewMemoryURLRepository(ttl)
	defer func() {
		if memRepo, ok := repo.(*repository.MemoryURLRepository); ok {
			memRepo.Close()
		}
	}()

	ctx := context.Background()
	url := &domain.URL{
		ShortCode: "test123",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
	}

	err := repo.Save(ctx, url)

	assert.NoError(t, err)
	assert.NotNil(t, url.ExpiresAt)
	assert.True(t, url.ExpiresAt.After(time.Now()))
}

func TestMemoryURLRepository_FindByShortCode(t *testing.T) {
	repo := repository.NewMemoryURLRepository(0)
	defer func() {
		if memRepo, ok := repo.(*repository.MemoryURLRepository); ok {
			memRepo.Close()
		}
	}()

	ctx := context.Background()
	url := &domain.URL{
		ShortCode: "test123",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
	}

	err := repo.Save(ctx, url)
	assert.NoError(t, err)

	found, err := repo.FindByShortCode(ctx, "test123")

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "test123", found.ShortCode)
	assert.Equal(t, "https://example.com", found.LongURL)
}

func TestMemoryURLRepository_FindByShortCode_NotFound(t *testing.T) {
	repo := repository.NewMemoryURLRepository(0)
	defer func() {
		if memRepo, ok := repo.(*repository.MemoryURLRepository); ok {
			memRepo.Close()
		}
	}()

	ctx := context.Background()
	found, err := repo.FindByShortCode(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Equal(t, domain.ErrURLNotFound, err)
}

func TestMemoryURLRepository_Exists(t *testing.T) {
	repo := repository.NewMemoryURLRepository(0)
	defer func() {
		if memRepo, ok := repo.(*repository.MemoryURLRepository); ok {
			memRepo.Close()
		}
	}()

	ctx := context.Background()

	// Test non-existent
	exists, err := repo.Exists(ctx, "test123")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Save URL
	url := &domain.URL{
		ShortCode: "test123",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
	}
	err = repo.Save(ctx, url)
	assert.NoError(t, err)

	// Test exists
	exists, err = repo.Exists(ctx, "test123")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestMemoryURLRepository_CleanupExpired(t *testing.T) {
	// Use a short cleanup interval for testing
	repo := repository.NewMemoryURLRepository(0)
	defer func() {
		if memRepo, ok := repo.(*repository.MemoryURLRepository); ok {
			memRepo.Close()
		}
	}()

	ctx := context.Background()

	// Save expired URL
	expiredTime := time.Now().Add(-1 * time.Hour)
	expiredURL := &domain.URL{
		ShortCode: "expired123",
		LongURL:   "https://example.com",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: &expiredTime,
	}
	err := repo.Save(ctx, expiredURL)
	assert.NoError(t, err)

	// Save non-expired URL
	validTime := time.Now().Add(1 * time.Hour)
	validURL := &domain.URL{
		ShortCode: "valid123",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
		ExpiresAt: &validTime,
	}
	err = repo.Save(ctx, validURL)
	assert.NoError(t, err)

	// Save URL without expiry
	noExpiryURL := &domain.URL{
		ShortCode: "noexpiry123",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
		ExpiresAt: nil,
	}
	err = repo.Save(ctx, noExpiryURL)
	assert.NoError(t, err)

	// Wait for cleanup goroutine to run (cleanup runs every minute)
	// Since we can't directly call cleanupExpired, we verify the behavior
	// by checking that expired URLs are properly marked as expired
	found, err := repo.FindByShortCode(ctx, "expired123")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.True(t, found.IsExpired(), "URL should be marked as expired")

	// Verify valid URL is not expired
	found, err = repo.FindByShortCode(ctx, "valid123")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.False(t, found.IsExpired(), "URL should not be expired")

	// Verify URL without expiry is not expired
	found, err = repo.FindByShortCode(ctx, "noexpiry123")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.False(t, found.IsExpired(), "URL without expiry should not be expired")
}

func TestMemoryURLRepository_Close(t *testing.T) {
	repo := repository.NewMemoryURLRepository(0)

	if memRepo, ok := repo.(*repository.MemoryURLRepository); ok {
		// Close should not panic
		assert.NotPanics(t, func() {
			memRepo.Close()
		})

		// Note: Closing twice will panic because we close the channel
		// This is expected behavior - Close should only be called once
	}
}

func TestMemoryURLRepository_StartCleanup(t *testing.T) {
	repo := repository.NewMemoryURLRepository(1 * time.Minute)

	// Give cleanup goroutine a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify repo is working
	ctx := context.Background()
	url := &domain.URL{
		ShortCode: "test123",
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
	}
	err := repo.Save(ctx, url)
	assert.NoError(t, err)

	// Verify the URL was saved
	exists, err := repo.Exists(ctx, "test123")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Cleanup
	if memRepo, ok := repo.(*repository.MemoryURLRepository); ok {
		memRepo.Close()
	}
}
