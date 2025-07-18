package main

import (
	"log/slog"

	"github.com/jeancarloshp/desafio-frete-rapido/internal/quote"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/config"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/database"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/database/querier"
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

	db, err := database.NewConnection(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}
	defer db.Close()

	q := querier.New(db)

	fastDeliveryAPI := fastdeliveryapi.New(cfg)

	quoteRepository := quote.NewQuoteRepository(q)
	quoteController := quote.NewQuoteController(cfg, quoteRepository, fastDeliveryAPI)
	quoteHandler := quote.NewQuoteHandler(quoteController)

	v1.Post("/quote", quoteHandler.QuoteSimulationHandler)
	v1.Get("/metrics", quoteHandler.QuoteMetricsHandler)

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		slog.Error("failed to start server", "error", err)
	}
}
