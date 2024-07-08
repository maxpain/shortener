package app

import (
	"log/slog"
	"time"

	jwtMiddleware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	compressMiddleware "github.com/gofiber/fiber/v2/middleware/compress"
	loggerMiddleware "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/handler"
)

func setupRoutes(
	app *fiber.App,
	cfg *config.Config,
	logger *slog.Logger,
	handler *handler.LinkHandler,
) {
	app.Use(compressMiddleware.New())
	app.Use(loggerMiddleware.New())
	app.Use(jwtMiddleware.New(jwtMiddleware.Config{
		SigningKey: jwtMiddleware.SigningKey{
			JWTAlg: "HS256",
			Key:    []byte(cfg.JwtSecret),
		},
		TokenLookup: "cookie:auth",
		ErrorHandler: func(c *fiber.Ctx, _ error) error {
			expiresAt := time.Now().Add(time.Hour * 72)
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"userID": uuid.New().String(),
				"exp":    expiresAt.Unix(),
			})

			t, err := token.SignedString([]byte(cfg.JwtSecret))
			if err != nil {
				logger.Error("Failed to sign JWT token", slog.Any("error", err))

				return c.SendStatus(fiber.StatusInternalServerError)
			}

			c.Cookie(&fiber.Cookie{
				Name:     "auth",
				Value:    t,
				Expires:  expiresAt,
				HTTPOnly: true,
			})

			c.Locals("user", token)

			return c.Next()
		},
	}))

	app.Get("/ping", handler.Ping)

	// Plain routes
	app.Get("/:hash", handler.Redirect)
	app.Post("/", handler.ShortenSinglePlain)

	// API routes
	app.Get("/api/user/urls", handler.GetUserLinks)
	app.Post("/api/shorten", handler.ShortenSingleJSON)
	app.Post("/api/shorten/batch", handler.ShortenBatchJSON)
}
