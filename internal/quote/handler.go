package quote

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type QuoteHandler struct {
	quoteController *QuoteController
}

func NewQuoteHandler(quoteController *QuoteController) *QuoteHandler {
	handler := &QuoteHandler{
		quoteController: quoteController,
	}

	return handler
}

func (qh *QuoteHandler) QuoteSimulationHandler(c *fiber.Ctx) error {
	var quoteRequest QuoteRequest
	if err := c.BodyParser(&quoteRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	v := validator.New()
	if err := v.Struct(quoteRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation failed",
			"details": err.Error(),
		})
	}

	quoteResponse, err := qh.quoteController.SimulateQuote(quoteRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to process quote request",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(quoteResponse)
}

func (qh *QuoteHandler) QuoteMetricsHandler(c *fiber.Ctx) error {
	lastQuotes := c.QueryInt("last_quotes", 10)

	if lastQuotes < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "last_quotes must be a positive integer",
		})
	}

	metrics, err := qh.quoteController.QuoteMetrics(lastQuotes)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to retrieve quote metrics",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(metrics)
}
