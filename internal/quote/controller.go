package quote

import (
	"fmt"
	"strconv"

	"github.com/jeancarloshp/desafio-frete-rapido/pkg/config"
	fastdeliveryapi "github.com/jeancarloshp/desafio-frete-rapido/pkg/fastdelivery_api"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/fastdelivery_api/models"
)

type QuoteController struct {
	cfg             *config.Config
	quoteRepository *QuoteRepository
	api             *fastdeliveryapi.FastDeliveryAPI
}

func NewQuoteController(cfg *config.Config, quoteRepository *QuoteRepository, api *fastdeliveryapi.FastDeliveryAPI) *QuoteController {
	return &QuoteController{
		cfg:             cfg,
		quoteRepository: quoteRepository,
		api:             api,
	}
}

func (qc *QuoteController) SimulateQuote(quoteRequest QuoteRequest) (*QuoteResponse, error) {
	zipcode, err := strconv.Atoi(quoteRequest.Recipient.Address.ZipCode)
	if err != nil {
		return nil, err
	}

	volumes := make([]models.Volume, len(quoteRequest.Volumes))
	for i, v := range quoteRequest.Volumes {
		category := strconv.Itoa(v.Category)

		volumes[i] = models.Volume{
			Category:      category,
			Amount:        v.Amount,
			UnitaryWeight: v.UnitaryWeight,
			UnitaryPrice:  v.Price,
			SKU:           v.SKU,
			Height:        v.Height,
			Width:         v.Width,
			Length:        v.Length,
		}
	}

	fastDeliveryQuoteRequest := models.QuoteRequest{
		Shipper: models.Shipper{
			RegisteredNumber: qc.cfg.FastDeliveryAPISenderCNPJ,
			Token:            qc.cfg.FastDeliveryAPIToken,
			PlatformCode:     qc.cfg.FastDeliveryAPIPlatformCode,
		},
		Recipient: models.Recipient{
			Type:    0,
			Country: "BRA",
			Zipcode: zipcode,
		},
		Dispatchers: []models.Dispatcher{
			{
				RegisteredNumber: qc.cfg.FastDeliveryAPISenderCNPJ,
				Zipcode:          qc.cfg.FastDeliveryAPIZipCode,
				Volumes:          volumes,
			},
		},
		SimulationType: []int{0},
	}

	quoteResponse, err := qc.api.SimulateQuote(fastDeliveryQuoteRequest)
	if err != nil {
		return nil, err
	}

	carriers := make([]Carrier, len(quoteResponse.Dispatchers[0].Offers))
	for i, o := range quoteResponse.Dispatchers[0].Offers {
		carriers[i] = Carrier{
			Name:     o.Carrier.Name,
			Price:    o.FinalPrice,
			Service:  o.Service,
			Deadline: o.DeliveryTime.Days,
		}
	}

	for _, c := range carriers {
		err := qc.quoteRepository.SaveQuote(c)
		if err != nil {
			return nil, fmt.Errorf("failed to save quote: %w", err)
		}
	}

	response := &QuoteResponse{
		Carriers: carriers,
	}

	return response, nil
}

func (qc *QuoteController) QuoteMetrics(lastQuotes int) (QuoteMetrics, error) {
	quotes, err := qc.quoteRepository.FindQuotesByLastQuote(lastQuotes)
	if err != nil {
		return QuoteMetrics{}, fmt.Errorf("failed to find last quotes: %w", err)
	}

	if len(quotes) == 0 {
		return QuoteMetrics{}, nil
	}

	var totalPrice float64
	carrierQuotesMap := make(map[string]*CarrierQuotes)

	for _, quote := range quotes {
		totalPrice += quote.Price

		if _, exists := carrierQuotesMap[quote.Name]; !exists {
			carrierQuotesMap[quote.Name] = &CarrierQuotes{
				CarrierName: quote.Name,
			}
		}

		carrierQuote := carrierQuotesMap[quote.Name]
		carrierQuote.TotalQuotes++
		carrierQuote.TotalPrice += quote.Price
	}

	var cheapestShipping, highestShipping float64
	for _, cq := range carrierQuotesMap {
		cq.AveragePrice = cq.TotalPrice / float64(cq.TotalQuotes)
		if cheapestShipping == 0 || cq.AveragePrice < cheapestShipping {
			cheapestShipping = cq.AveragePrice
		}
		if cq.AveragePrice > highestShipping {
			highestShipping = cq.AveragePrice
		}
	}

	carrierQuotesSlice := make([]CarrierQuotes, 0, len(carrierQuotesMap))
	for _, cq := range carrierQuotesMap {
		carrierQuotesSlice = append(carrierQuotesSlice, *cq)
	}

	return QuoteMetrics{
		CarrierQuotes:    carrierQuotesSlice,
		CheapestShipping: cheapestShipping,
		HighestShipping:  highestShipping,
	}, nil
}
