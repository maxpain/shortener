package repository

import (
	"context"
	"io"

	"github.com/maxpain/shortener/internal/model"
)

type Repository interface {
	io.Closer

	GetLink(ctx context.Context, hash string) (*model.StoredLink, error)
	GetUserLinks(ctx context.Context, userID string) ([]*model.StoredLink, error)
	SaveLinks(ctx context.Context, links []*model.StoredLink) ([]bool, error)
	MarkForDeletion(hashes []string, userID string) error

	Init(ctx context.Context) error
	Ping(ctx context.Context) error
}
