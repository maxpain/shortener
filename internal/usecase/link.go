package usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"

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

type LinkUseCase struct {
	logger *slog.Logger
	repo   Repository
}

func New(repo Repository, logger *slog.Logger) *LinkUseCase {
	return &LinkUseCase{
		logger: logger.With(
			slog.String("usecase", "link"),
		),
		repo: repo,
	}
}

func (u *LinkUseCase) Shorten(
	ctx context.Context,
	linksToShorten []*model.Link,
	baseURL string,
	userID string,
) ([]*model.ShortenedLink, error) {
	linksToStore := make([]*model.StoredLink, 0, len(linksToShorten))
	shortenedLinks := make([]*model.ShortenedLink, 0, len(linksToShorten))

	for _, linkToShorten := range linksToShorten {
		storedLink := linkToShorten.GetStoredLink(userID)
		linksToStore = append(linksToStore, storedLink)

		shortenedLink, err := storedLink.GetShortenedLink(baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get shortened link: %w", err)
		}

		shortenedLinks = append(shortenedLinks, shortenedLink)
	}

	results, err := u.repo.SaveLinks(ctx, linksToStore)
	if err != nil {
		return nil, fmt.Errorf("failed to save links: %w", err)
	}

	for i, isSaved := range results {
		shortenedLinks[i].Saved = isSaved
	}

	return shortenedLinks, nil
}

func (u *LinkUseCase) Resolve(ctx context.Context, hash string) (string, error) {
	storedLink, err := u.repo.GetLink(ctx, hash)
	if err != nil {
		return "", fmt.Errorf("failed to get link from repo: %w", err)
	}

	if storedLink.IsDeleted {
		return "", model.ErrDeleted
	}

	return storedLink.OriginalURL, nil
}

func (u *LinkUseCase) GetUserLinks(ctx context.Context, baseURL string, userID string) ([]*model.UserLink, error) {
	links, err := u.repo.GetUserLinks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user links: %w", err)
	}

	userLinks := make([]*model.UserLink, 0, len(links))

	for _, link := range links {
		if link.IsDeleted {
			continue
		}

		shortenedLink, err := link.GetShortenedLink(baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get shortened link: %w", err)
		}

		userLinks = append(userLinks, &model.UserLink{
			OriginalURL: link.OriginalURL,
			ShortURL:    shortenedLink.ShortURL,
		})
	}

	return userLinks, nil
}

func (u *LinkUseCase) DeleteUserLinks(hashes []string, userID string) error {
	err := u.repo.MarkForDeletion(hashes, userID)
	if err != nil {
		return fmt.Errorf("failed to mark links for deletion: %w", err)
	}

	return nil
}

func (u *LinkUseCase) Ping(ctx context.Context) error {
	err := u.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping repository: %w", err)
	}

	return nil
}
