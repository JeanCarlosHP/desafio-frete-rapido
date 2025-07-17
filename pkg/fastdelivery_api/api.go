package fastdeliveryapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/config"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/fastdelivery_api/models"
)

type FastDeliveryAPI struct {
	cfg    *config.Config
	client *http.Client
}

func New(cfg *config.Config) *FastDeliveryAPI {
	return &FastDeliveryAPI{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (api *FastDeliveryAPI) SimulateQuote(quoteRequest models.QuoteRequest) (*models.QuoteResponse, error) {
	body, err := json.Marshal(quoteRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/quote/simulate", api.cfg.FastDeliveryAPIBaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	slog.Debug("sending quote simulation request",
		"url", url,
		"body", string(body),
		"headers", req.Header,
	)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("failed to get quote simulation",
			"status_code", resp.StatusCode,
			"body", string(body),
			"response_body", resp.Body,
		)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var quoteResponse models.QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	v := validator.New()
	if err := v.Struct(quoteResponse); err != nil {
		return nil, fmt.Errorf("error validating fast delivery quote response: %w", err)
	}

	return &quoteResponse, nil
}
