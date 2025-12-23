package domain

import "context"

type URLRepository interface {
	Save(ctx context.Context, url *URL) error
	FindByShortCode(ctx context.Context, shortCode string) (*URL, error)
	Exists(ctx context.Context, shortCode string) (bool, error)
}
