package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/maxpain/shortener/internal/model"
)

type LinkUseCase interface {
	Shorten(ctx context.Context, links []*model.Link, baseURL string, userID string) ([]*model.ShortenedLink, error)
	Resolve(ctx context.Context, hash string) (string, error)
	GetUserLinks(ctx context.Context, baseURL string, userID string) ([]*model.UserLink, error)
	Ping(ctx context.Context) error
}

type LinkHandler struct {
	logger  *slog.Logger
	baseURL string
	useCase LinkUseCase
}

func New(u LinkUseCase, logger *slog.Logger, baseURL string) *LinkHandler {
	return &LinkHandler{
		logger: logger.With(
			slog.String("handler", "link"),
		),
		baseURL: baseURL,
		useCase: u,
	}
}

func (h *LinkHandler) getUserIdFromContext(c *fiber.Ctx) (string, error) {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return "", errors.New("failed to get token from context")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		h.logger.Debug("Failed to get claims from token")

		return "", errors.New("failed to get claims from token")
	}

	userID, ok := claims["userID"].(string)
	if !ok {
		h.logger.Debug("Failed to get userID from claims")

		return "", errors.New("failed to get userID from claims")
	}

	return userID, nil
}

func (h *LinkHandler) Ping(c *fiber.Ctx) error {
	err := h.useCase.Ping(c.Context())
	if err != nil {
		h.logger.Error("Ping failed", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *LinkHandler) Redirect(c *fiber.Ctx) error {
	shortURL := c.Params("hash")

	if shortURL == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Short URL is required")
	}

	originalURL, err := h.useCase.Resolve(c.Context(), shortURL)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).SendString("URL not found")
		}

		h.logger.Error("Failed to resolve URL", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Redirect(originalURL, fiber.StatusTemporaryRedirect)
}

func (h *LinkHandler) ShortenSinglePlain(c *fiber.Ctx) error {
	userID, err := h.getUserIdFromContext(c)
	if err != nil {
		h.logger.Error("Failed to get user ID from context", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	originalURL := string(c.Body())

	if originalURL == "" {
		return c.Status(fiber.StatusBadRequest).SendString("URL is required")
	}

	shortenedLinks, err := h.useCase.Shorten(
		c.Context(),
		[]*model.Link{{OriginalURL: originalURL}},
		h.baseURL,
		userID,
	)
	if err != nil {
		h.logger.Error("Failed to shorten URL", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if len(shortenedLinks) == 0 {
		h.logger.Error("Empty response from use case")

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	status := fiber.StatusCreated

	if !shortenedLinks[0].Saved {
		status = fiber.StatusConflict
	}

	return c.Status(status).SendString(shortenedLinks[0].ShortURL)
}

func (h *LinkHandler) ShortenSingleJSON(c *fiber.Ctx) error {
	userID, err := h.getUserIdFromContext(c)
	if err != nil {
		h.logger.Error("Failed to get user ID from context", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	var r struct {
		URL string `json:"url"`
	}

	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	if r.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "URL is required"})
	}

	shortenedLinks, err := h.useCase.Shorten(c.Context(), []*model.Link{{OriginalURL: r.URL}}, h.baseURL, userID)
	if err != nil {
		h.logger.Error("Failed to shorten URL", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if len(shortenedLinks) == 0 {
		h.logger.Error("Empty response from use case")

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	status := fiber.StatusCreated

	if !shortenedLinks[0].Saved {
		status = fiber.StatusConflict
	}

	return c.Status(status).JSON(fiber.Map{"result": shortenedLinks[0].ShortURL})
}

func (h *LinkHandler) ShortenBatchJSON(c *fiber.Ctx) error {
	userID, err := h.getUserIdFromContext(c)
	if err != nil {
		h.logger.Error("Failed to get user ID from context", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	var links []*model.Link

	if err := c.BodyParser(&links); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	if len(links) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "URLs are required"})
	}

	shortenedLinks, err := h.useCase.Shorten(c.Context(), links, h.baseURL, userID)
	if err != nil {
		h.logger.Error("Failed to shorten URLs", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if len(shortenedLinks) == 0 {
		h.logger.Error("Empty response from use case")

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	status := fiber.StatusCreated

	for _, link := range shortenedLinks {
		if !link.Saved {
			status = fiber.StatusConflict

			break
		}
	}

	return c.Status(status).JSON(shortenedLinks)
}

func (h *LinkHandler) GetUserLinks(c *fiber.Ctx) error {
	userID, err := h.getUserIdFromContext(c)
	if err != nil {
		h.logger.Error("Failed to get user ID from context", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	links, err := h.useCase.GetUserLinks(c.Context(), h.baseURL, userID)
	if err != nil {
		h.logger.Error("Failed to get user links", slog.Any("error", err))

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(links)
}
