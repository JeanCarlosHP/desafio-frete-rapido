package quote

import (
	"strconv"

	"github.com/jeancarloshp/desafio-frete-rapido/pkg/config"
	fastdeliveryapi "github.com/jeancarloshp/desafio-frete-rapido/pkg/fastdelivery_api"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/fastdelivery_api/models"
)

type QuoteController struct {
	cfg *config.Config
	api *fastdeliveryapi.FastDeliveryAPI
}

func NewQuoteController(cfg *config.Config, api *fastdeliveryapi.FastDeliveryAPI) *QuoteController {
	return &QuoteController{
		cfg: cfg,
		api: api,
	}
}

func (qc *QuoteController) Process(quoteRequest QuoteRequest) (*QuoteResponse, error) {
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
			Price:    o.CostPrice,
			Service:  o.Service,
			Deadline: o.DeliveryTime.Days,
		}
	}

	response := &QuoteResponse{
		Carriers: carriers,
	}

	return response, nil
}
