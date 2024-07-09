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
	logger    *slog.Logger
	links     sync.Map
	userLinks sync.Map
	file      *os.File
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

		err := r.saveLinkToMemory(&link)
		if err != nil {
			return fmt.Errorf("failed to save link to memory: %w", err)
		}
	}

	return nil
}

func (r *Repository) GetLink(_ context.Context, hash string) (*model.StoredLink, error) {
	if l, ok := r.links.Load(hash); ok {
		link, ok := l.(*model.StoredLink)

		if !ok {
			return nil, errCastLink
		}

		return link, nil
	}

	return nil, model.ErrNotFound
}

func (r *Repository) GetUserLinks(_ context.Context, userID string) ([]*model.StoredLink, error) {
	if l, ok := r.userLinks.Load(userID); ok {
		links, ok := l.([]*model.StoredLink)

		if !ok {
			return nil, errCastLink
		}

		return links, nil
	}

	r.logger.Debug("user links not found",
		slog.String("user_id", userID),
	)

	return []*model.StoredLink{}, nil
}

func (r *Repository) SaveLinks(ctx context.Context, linksToStore []*model.StoredLink) ([]bool, error) {
	results := make([]bool, 0, len(linksToStore))

	for _, link := range linksToStore {
		isExists := true

		if _, err := r.GetLink(ctx, link.Hash); err != nil {
			if errors.Is(err, model.ErrNotFound) {
				isExists = false
			} else {
				return nil, fmt.Errorf("failed to get link: %w", err)
			}
		}

		if !isExists {
			if err := r.saveLinkToMemory(link); err != nil {
				return nil, fmt.Errorf("failed to save link to memory: %w", err)
			}

			if err := r.saveLinkToFile(link); err != nil {
				return nil, fmt.Errorf("failed to save link to file: %w", err)
			}
		}

		results = append(results, !isExists)
	}

	return results, nil
}

func (r *Repository) MarkForDeletion(hashes []string, userID string) error {
	return nil
}

func (r *Repository) saveLinkToMemory(link *model.StoredLink) error {
	r.logger.Debug("saving link to memory",
		slog.Group("link",
			slog.String("hash", link.Hash),
			slog.String("original_url", link.OriginalURL),
			slog.String("correlation_id", link.CorrelationID),
			slog.String("user_id", link.UserID),
		),
	)

	r.links.Store(link.Hash, link)
	userLinks := []*model.StoredLink{link}

	if l, ok := r.userLinks.Load(link.UserID); ok {
		links, ok := l.([]*model.StoredLink)

		if !ok {
			return errCastLink
		}

		userLinks = append(links, userLinks...)
	}

	r.userLinks.Store(link.UserID, userLinks)

	return nil
}

func (r *Repository) saveLinkToFile(link *model.StoredLink) error {
	if r.file != nil {
		err := json.NewEncoder(r.file).Encode(link)
		if err != nil {
			return fmt.Errorf("failed to encode link: %w", err)
		}
	}

	return nil
}

func (r *Repository) Ping(_ context.Context) error {
	return nil
}

func (r *Repository) Close() error {
	if r.file != nil {
		return r.file.Close()
	}

	return nil
}
