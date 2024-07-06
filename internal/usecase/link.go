package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/maxpain/shortener/internal/model"
	"github.com/maxpain/shortener/internal/repository"
)

type LinkUseCase struct {
	logger *slog.Logger
	repo   repository.Repository
}

func New(repo repository.Repository, logger *slog.Logger) *LinkUseCase {
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
) ([]*model.ShortenedLink, error) {
	linksToStore := make([]*model.StoredLink, 0, len(linksToShorten))
	shortenedLinks := make([]*model.ShortenedLink, 0, len(linksToShorten))

	for _, linkToShorten := range linksToShorten {
		storedLink := linkToShorten.GetStoredLink()
		linksToStore = append(linksToStore, storedLink)

		shortenedLink, err := storedLink.GetShortenedLink(baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get shortened link: %w", err)
		}

		shortenedLinks = append(shortenedLinks, shortenedLink)
	}

	results, err := u.repo.Save(ctx, linksToStore)
	if err != nil {
		return nil, fmt.Errorf("failed to save links: %w", err)
	}

	for i, isSaved := range results {
		shortenedLinks[i].Saved = isSaved
	}

	return shortenedLinks, nil
}

func (u *LinkUseCase) Resolve(
	ctx context.Context,
	hash string,
) (string, error) {
	storedLink, err := u.repo.Get(ctx, hash)
	if err != nil {
		return "", fmt.Errorf("failed to get link from repo: %w", err)
	}

	return storedLink.OriginalURL, nil
}

func (u *LinkUseCase) Ping(ctx context.Context) error {
	err := u.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping repository: %w", err)
	}

	return nil
}
