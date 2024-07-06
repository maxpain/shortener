package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/maxpain/shortener/internal/model"
)

var errCastLink = errors.New("failed to cast link")

type Repository struct {
	logger *slog.Logger
	links  sync.Map
	file   *os.File
}

// Create a new memory repository with optional persistence to the file.
func New(file *os.File, logger *slog.Logger) *Repository {
	return &Repository{
		logger: logger.With(
			slog.String("repository", "memory"),
		),
		file: file,
	}
}

func (r *Repository) Init(_ context.Context) error {
	if r.file == nil {
		return nil
	}

	decoder := json.NewDecoder(r.file)

	for {
		var link model.StoredLink

		if err := decoder.Decode(&link); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("failed to decode link: %w", err)
		}

		r.links.Store(link.Hash, &link)
		r.logger.Debug("loaded link",
			slog.Group("link",
				slog.String("hash", link.Hash),
				slog.String("original_url", link.OriginalURL),
				slog.String("correlation_id", link.CorrelationID),
			),
		)
	}

	return nil
}

func (r *Repository) Get(_ context.Context, hash string) (*model.StoredLink, error) {
	if l, ok := r.links.Load(hash); ok {
		link, ok := l.(*model.StoredLink)

		if !ok {
			return nil, errCastLink
		}

		return link, nil
	}

	return nil, model.ErrNotFound
}

func (r *Repository) Save(ctx context.Context, linksToStore []*model.StoredLink) ([]bool, error) {
	results := make([]bool, 0, len(linksToStore))

	for _, link := range linksToStore {
		isExists := true

		_, err := r.Get(ctx, link.Hash)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				isExists = false
			} else {
				return nil, fmt.Errorf("failed to get link: %w", err)
			}
		}

		if !isExists {
			r.links.Store(link.Hash, link)

			if r.file != nil {
				err := json.NewEncoder(r.file).Encode(link)
				if err != nil {
					return nil, fmt.Errorf("failed to encode link: %w", err)
				}
			}
		}

		results = append(results, !isExists)
	}

	return results, nil
}

func (r *Repository) Ping(_ context.Context) error {
	return nil
}

func (r *Repository) Close() {
	if r.file != nil {
		r.file.Close()
	}
}
