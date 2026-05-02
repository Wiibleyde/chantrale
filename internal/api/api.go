package api

import (
	"errors"
	"time"

	apibed "LsmsBot/internal/api/features/bed"
	apiduty "LsmsBot/internal/api/features/duty"
	apimortuary "LsmsBot/internal/api/features/mortuary"
	apiradio "LsmsBot/internal/api/features/radio"
	"LsmsBot/internal/api/middleware"
	"LsmsBot/internal/config"
	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cache"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// Run starts the REST API server in a background goroutine and returns the
// Fiber app so the caller can shut it down gracefully.
// If API_KEY is not set, the server is not started.
func Run(client *bot.Client) *fiber.App {
	cfg := config.Load()

	if cfg.APIKey == "" {
		logger.Warn("API_KEY n'est pas défini — le serveur API ne démarrera pas.")
		return nil
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "Erreur interne du serveur."
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
				msg = e.Message
			}
			return c.Status(code).JSON(fiber.Map{"error": msg})
		},
	})

	// Rate limiter: max 30 requests per 10 seconds per IP.
	// Protects against flooding before hitting auth or handlers.
	apiLimiter := limiter.New(limiter.Config{
		Max:        30,
		Expiration: 10 * time.Second,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Trop de requêtes. Veuillez réessayer dans quelques secondes.",
			})
		},
	})

	// Cache: responses are cached for 10 seconds per unique URL.
	// This prevents hammering the Discord REST API (GetMembers, GetMessage)
	// regardless of how frequently the API is polled.
	apiCache := cache.New(cache.Config{
		Expiration: 10 * time.Second,
	})

	api := app.Group("/api", apiLimiter, middleware.Auth(cfg.APIKey), apiCache)
	apiduty.Register(api, client)
	apibed.Register(api, client)
	apimortuary.Register(api, client)
	apiradio.Register(api, client)

	port := cfg.APIPort
	if port == "" {
		port = "3000"
	}

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("Erreur du serveur API", "error", err)
		}
	}()

	logger.Info("Serveur API démarré", "port", port)
	return app
}
