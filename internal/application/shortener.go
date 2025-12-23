package application

import (
	"context"
	"fmt"
	"time"
	"url-shortener/internal/domain"
)

type ShortenerService struct {
	repo      domain.URLRepository
	generator domain.ShortCodeGenerator
}

func NewShortenerService(repo domain.URLRepository, generator domain.ShortCodeGenerator) *ShortenerService {
	return &ShortenerService{
		repo:      repo,
		generator: generator,
	}
}

func (s *ShortenerService) CreateShortURL(ctx context.Context, longURL string) (*domain.URL, error) {
	shortCode := s.generator.Generate()

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		exists, err := s.repo.Exists(ctx, shortCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check short code existence: %w", err)
		}
		if !exists {
			break
		}
		shortCode = s.generator.Generate()
	}

	url := &domain.URL{
		ShortCode: shortCode,
		LongURL:   longURL,
		CreatedAt: time.Now(),
	}

	if err := url.Validate(); err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	if err := s.repo.Save(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to save url: %w", err)
	}

	return url, nil
}

func (s *ShortenerService) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	url, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to get url: %w", err)
	}

	if url.IsExpired() {
		return "", domain.ErrURLNotFound
	}

	return url.LongURL, nil
}
