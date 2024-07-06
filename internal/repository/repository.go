package repository

import (
	"context"

	"github.com/maxpain/shortener/internal/model"
)

type Repository interface {
	Get(ctx context.Context, hash string) (*model.StoredLink, error)
	Save(ctx context.Context, links []*model.StoredLink) ([]bool, error)
	Init(ctx context.Context) error
	Ping(ctx context.Context) error
	Close()
}
