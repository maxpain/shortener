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

	return storedLink.OriginalURL, nil
}

func (u *LinkUseCase) GetUserLinks(ctx context.Context, baseURL string, userID string) ([]*model.UserLink, error) {
	links, err := u.repo.GetUserLinks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user links: %w", err)
	}

	userLinks := make([]*model.UserLink, 0, len(links))

	for _, link := range links {
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

func (u *LinkUseCase) Ping(ctx context.Context) error {
	err := u.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping repository: %w", err)
	}

	return nil
}
