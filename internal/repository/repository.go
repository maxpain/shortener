package repository

import (
	"context"

	"github.com/maxpain/shortener/internal/model"
)

type Repository interface {
	GetLink(ctx context.Context, hash string) (*model.StoredLink, error)
	GetUserLinks(ctx context.Context, userID string) ([]*model.StoredLink, error)
	SaveLinks(ctx context.Context, links []*model.StoredLink) ([]bool, error)

	Init(ctx context.Context) error
	Ping(ctx context.Context) error
	Close()
}
