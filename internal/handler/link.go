package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/maxpain/shortener/internal/model"
)

type LinkUseCase interface {
	Shorten(ctx context.Context, links []*model.Link, baseURL string) ([]*model.ShortenedLink, error)
	Resolve(ctx context.Context, hash string) (string, error)
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
	originalURL := string(c.Body())

	if originalURL == "" {
		return c.Status(fiber.StatusBadRequest).SendString("URL is required")
	}

	shortenedLinks, err := h.useCase.Shorten(c.Context(), []*model.Link{{OriginalURL: originalURL}}, h.baseURL)
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
	var r struct {
		URL string `json:"url"`
	}

	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	if r.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "URL is required"})
	}

	shortenedLinks, err := h.useCase.Shorten(c.Context(), []*model.Link{{OriginalURL: r.URL}}, h.baseURL)
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
	var links []*model.Link

	if err := c.BodyParser(&links); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	if len(links) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "URLs are required"})
	}

	shortenedLinks, err := h.useCase.Shorten(c.Context(), links, h.baseURL)
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
