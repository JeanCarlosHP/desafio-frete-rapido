package main

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/config"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/logger"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/server"
)

var cfg *config.Config

func init() {
	cfg = config.New()
	logger.New(cfg)
}

func main() {
	app := server.New(cfg)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
