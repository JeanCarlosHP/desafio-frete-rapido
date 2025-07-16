package server

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/config"
)

func New(config *config.Config) *fiber.App {
	cfg := fiber.Config{
		DisableStartupMessage: false,
		ErrorHandler:          errorHandler,
	}

	app := fiber.New(cfg)

	app.Use(cors.New(cors.Config{
		AllowHeaders: config.HTTPCorsAllowedHeaders,
		AllowMethods: config.HTTPCorsAllowedMethods,
		AllowOrigins: config.HTTPCorsAllowedOrigins,
	}))

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e any) {
			buf := make([]byte, 4096)
			buf = buf[:runtime.Stack(buf, false)]
			slog.Error(fmt.Sprintf("Panic recovered: %v\nStack trace:\n%s", e, buf))
		},
	}))

	app.Use(requestid.New())

	app.Use(healthcheck.New())
	app.Use(logger.New())

	return app
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
