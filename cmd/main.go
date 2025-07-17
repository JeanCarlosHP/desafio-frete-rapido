package main

import (
	"log/slog"

	"github.com/jeancarloshp/desafio-frete-rapido/internal/quote"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/config"
	fastdeliveryapi "github.com/jeancarloshp/desafio-frete-rapido/pkg/fastdelivery_api"
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
	v1 := app.Group("/v1")

	fastDeliveryAPI := fastdeliveryapi.New(cfg)

	quoteController := quote.NewQuoteController(cfg, fastDeliveryAPI)
	quoteHandler := quote.NewQuoteHandler(quoteController)

	v1.Post("/quote", quoteHandler.QuoteSimulationHandler)

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		slog.Error("failed to start server", "error", err)
	}
}
