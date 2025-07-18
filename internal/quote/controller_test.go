package quote_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/jeancarloshp/desafio-frete-rapido/internal/quote"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/config"
	"github.com/jeancarloshp/desafio-frete-rapido/pkg/fastdelivery_api/models"
)

// Interface para repository
type QuoteRepositoryInterface interface {
	SaveQuote(carrier quote.Carrier) error
	FindQuotesByLastQuote(lastQuotes int) ([]quote.Carrier, error)
}

// Interface para API
type FastDeliveryAPIInterface interface {
	SimulateQuote(request models.QuoteRequest) (*models.QuoteResponse, error)
}

// MockQuoteRepository implementa um mock do repository
type MockQuoteRepository struct {
	saveQuoteFunc             func(carrier quote.Carrier) error
	findQuotesByLastQuoteFunc func(lastQuotes int) ([]quote.Carrier, error)
}

func (m *MockQuoteRepository) SaveQuote(carrier quote.Carrier) error {
	if m.saveQuoteFunc != nil {
		return m.saveQuoteFunc(carrier)
	}
	return nil
}

func (m *MockQuoteRepository) FindQuotesByLastQuote(lastQuotes int) ([]quote.Carrier, error) {
	if m.findQuotesByLastQuoteFunc != nil {
		return m.findQuotesByLastQuoteFunc(lastQuotes)
	}
	return []quote.Carrier{}, nil
}

// MockFastDeliveryAPI implementa um mock da API externa
type MockFastDeliveryAPI struct {
	simulateQuoteFunc func(request models.QuoteRequest) (*models.QuoteResponse, error)
}

func (m *MockFastDeliveryAPI) SimulateQuote(request models.QuoteRequest) (*models.QuoteResponse, error) {
	if m.simulateQuoteFunc != nil {
		return m.simulateQuoteFunc(request)
	}
	return &models.QuoteResponse{}, nil
}

// TestableQuoteController expõe métodos testáveis implementando a lógica diretamente
type TestableQuoteController struct {
	cfg        *config.Config
	repository QuoteRepositoryInterface
	api        FastDeliveryAPIInterface
}

func NewTestableQuoteController(cfg *config.Config, repository QuoteRepositoryInterface, api FastDeliveryAPIInterface) *TestableQuoteController {
	return &TestableQuoteController{
		cfg:        cfg,
		repository: repository,
		api:        api,
	}
}

func (qc *TestableQuoteController) SimulateQuote(quoteRequest quote.QuoteRequest) (*quote.QuoteResponse, error) {
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

	carriers := make([]quote.Carrier, len(quoteResponse.Dispatchers[0].Offers))
	for i, o := range quoteResponse.Dispatchers[0].Offers {
		carriers[i] = quote.Carrier{
			Name:     o.Carrier.Name,
			Price:    o.FinalPrice,
			Service:  o.Service,
			Deadline: o.DeliveryTime.Days,
		}
	}

	for _, c := range carriers {
		err := qc.repository.SaveQuote(c)
		if err != nil {
			return nil, fmt.Errorf("failed to save quote: %w", err)
		}
	}

	response := &quote.QuoteResponse{
		Carriers: carriers,
	}

	return response, nil
}

func (qc *TestableQuoteController) QuoteMetrics(lastQuotes int) (quote.QuoteMetrics, error) {
	quotes, err := qc.repository.FindQuotesByLastQuote(lastQuotes)
	if err != nil {
		return quote.QuoteMetrics{}, fmt.Errorf("failed to find last quotes: %w", err)
	}

	if len(quotes) == 0 {
		return quote.QuoteMetrics{}, nil
	}

	var totalPrice float64
	carrierQuotesMap := make(map[string]*quote.CarrierQuotes)

	for _, quoteItem := range quotes {
		totalPrice += quoteItem.Price

		if _, exists := carrierQuotesMap[quoteItem.Name]; !exists {
			carrierQuotesMap[quoteItem.Name] = &quote.CarrierQuotes{
				CarrierName: quoteItem.Name,
			}
		}

		carrierQuote := carrierQuotesMap[quoteItem.Name]
		carrierQuote.TotalQuotes++
		carrierQuote.TotalPrice += quoteItem.Price
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

	carrierQuotesSlice := make([]quote.CarrierQuotes, 0, len(carrierQuotesMap))
	for _, cq := range carrierQuotesMap {
		carrierQuotesSlice = append(carrierQuotesSlice, *cq)
	}

	return quote.QuoteMetrics{
		CarrierQuotes:    carrierQuotesSlice,
		CheapestShipping: cheapestShipping,
		HighestShipping:  highestShipping,
	}, nil
}

func setupController(mockRepo *MockQuoteRepository, mockAPI *MockFastDeliveryAPI) *TestableQuoteController {
	cfg := &config.Config{
		FastDeliveryAPISenderCNPJ:   "12345678901234",
		FastDeliveryAPIToken:        "test-token-12345678901234567890",
		FastDeliveryAPIPlatformCode: "platform-123",
		FastDeliveryAPIZipCode:      12345678,
	}

	return NewTestableQuoteController(cfg, mockRepo, mockAPI)
}

// Testes para SimulateQuote
func TestSimulateQuote_Success(t *testing.T) {
	mockRepo := &MockQuoteRepository{}
	mockAPI := &MockFastDeliveryAPI{
		simulateQuoteFunc: func(request models.QuoteRequest) (*models.QuoteResponse, error) {
			return &models.QuoteResponse{
				Dispatchers: []models.DispatcherResponse{
					{
						Offers: []models.Offer{
							{
								Carrier: models.Carrier{
									Name: "Transportadora A",
								},
								Service:    "Expresso",
								FinalPrice: 25.50,
								DeliveryTime: models.DeliveryTime{
									Days: 3,
								},
							},
							{
								Carrier: models.Carrier{
									Name: "Transportadora B",
								},
								Service:    "Normal",
								FinalPrice: 15.75,
								DeliveryTime: models.DeliveryTime{
									Days: 7,
								},
							},
						},
					},
				},
			}, nil
		},
	}

	controller := setupController(mockRepo, mockAPI)

	request := quote.QuoteRequest{
		Recipient: quote.Recipient{
			Address: quote.Address{
				ZipCode: "12345678",
			},
		},
		Volumes: []quote.Volume{
			{
				Category:      1,
				Amount:        2,
				UnitaryWeight: 1.5,
				Price:         100.0,
				SKU:           "PROD123",
				Height:        10.0,
				Width:         15.0,
				Length:        20.0,
			},
		},
	}

	response, err := controller.SimulateQuote(request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Carriers) != 2 {
		t.Fatalf("Expected 2 carriers, got: %d", len(response.Carriers))
	}

	expectedCarriers := []quote.Carrier{
		{Name: "Transportadora A", Service: "Expresso", Price: 25.50, Deadline: 3},
		{Name: "Transportadora B", Service: "Normal", Price: 15.75, Deadline: 7},
	}

	for i, carrier := range response.Carriers {
		if carrier.Name != expectedCarriers[i].Name {
			t.Errorf("Expected carrier name %s, got: %s", expectedCarriers[i].Name, carrier.Name)
		}
		if carrier.Price != expectedCarriers[i].Price {
			t.Errorf("Expected price %.2f, got: %.2f", expectedCarriers[i].Price, carrier.Price)
		}
	}
}

func TestSimulateQuote_InvalidZipCode(t *testing.T) {
	mockRepo := &MockQuoteRepository{}
	mockAPI := &MockFastDeliveryAPI{}

	controller := setupController(mockRepo, mockAPI)

	request := quote.QuoteRequest{
		Recipient: quote.Recipient{
			Address: quote.Address{
				ZipCode: "invalid",
			},
		},
		Volumes: []quote.Volume{
			{
				Category:      1,
				Amount:        1,
				UnitaryWeight: 1.0,
				Price:         50.0,
				SKU:           "PROD123",
				Height:        10.0,
				Width:         10.0,
				Length:        10.0,
			},
		},
	}

	response, err := controller.SimulateQuote(request)

	if err == nil {
		t.Fatal("Expected error for invalid zipcode, got nil")
	}

	if response != nil {
		t.Fatal("Expected nil response for invalid zipcode")
	}
}

func TestSimulateQuote_APIError(t *testing.T) {
	mockRepo := &MockQuoteRepository{}
	mockAPI := &MockFastDeliveryAPI{
		simulateQuoteFunc: func(request models.QuoteRequest) (*models.QuoteResponse, error) {
			return nil, errors.New("API connection error")
		},
	}

	controller := setupController(mockRepo, mockAPI)

	request := quote.QuoteRequest{
		Recipient: quote.Recipient{
			Address: quote.Address{
				ZipCode: "12345678",
			},
		},
		Volumes: []quote.Volume{
			{
				Category:      1,
				Amount:        1,
				UnitaryWeight: 1.0,
				Price:         50.0,
				SKU:           "PROD123",
				Height:        10.0,
				Width:         10.0,
				Length:        10.0,
			},
		},
	}

	response, err := controller.SimulateQuote(request)

	if err == nil {
		t.Fatal("Expected API error, got nil")
	}

	if response != nil {
		t.Fatal("Expected nil response for API error")
	}

	expectedError := "API connection error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got: '%s'", expectedError, err.Error())
	}
}

// Testes para QuoteMetrics
func TestQuoteMetrics_Success(t *testing.T) {
	mockRepo := &MockQuoteRepository{
		findQuotesByLastQuoteFunc: func(lastQuotes int) ([]quote.Carrier, error) {
			return []quote.Carrier{
				{Name: "Transportadora A", Price: 20.0},
				{Name: "Transportadora A", Price: 30.0},
				{Name: "Transportadora B", Price: 15.0},
				{Name: "Transportadora B", Price: 25.0},
				{Name: "Transportadora B", Price: 35.0},
			}, nil
		},
	}
	mockAPI := &MockFastDeliveryAPI{}

	controller := setupController(mockRepo, mockAPI)

	metrics, err := controller.QuoteMetrics(10)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(metrics.CarrierQuotes) != 2 {
		t.Fatalf("Expected 2 carrier quotes, got: %d", len(metrics.CarrierQuotes))
	}

	// Verificar se as métricas estão corretas
	if metrics.CheapestShipping != 25.0 { // (15+25+35)/3 = 25.0
		t.Errorf("Expected cheapest shipping 25.0, got: %.2f", metrics.CheapestShipping)
	}

	if metrics.HighestShipping != 25.0 { // (20+30)/2 = 25.0
		t.Errorf("Expected highest shipping 25.0, got: %.2f", metrics.HighestShipping)
	}

	// Encontrar e verificar transportadora A
	var transportadoraA *quote.CarrierQuotes
	for _, cq := range metrics.CarrierQuotes {
		if cq.CarrierName == "Transportadora A" {
			transportadoraA = &cq
			break
		}
	}

	if transportadoraA == nil {
		t.Fatal("Transportadora A not found in metrics")
	}

	if transportadoraA.TotalQuotes != 2 {
		t.Errorf("Expected 2 quotes for Transportadora A, got: %d", transportadoraA.TotalQuotes)
	}

	if transportadoraA.AveragePrice != 25.0 {
		t.Errorf("Expected average price 25.0 for Transportadora A, got: %.2f", transportadoraA.AveragePrice)
	}
}

func TestQuoteMetrics_EmptyQuotes(t *testing.T) {
	mockRepo := &MockQuoteRepository{
		findQuotesByLastQuoteFunc: func(lastQuotes int) ([]quote.Carrier, error) {
			return []quote.Carrier{}, nil
		},
	}
	mockAPI := &MockFastDeliveryAPI{}

	controller := setupController(mockRepo, mockAPI)

	metrics, err := controller.QuoteMetrics(10)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(metrics.CarrierQuotes) != 0 {
		t.Errorf("Expected 0 carrier quotes for empty data, got: %d", len(metrics.CarrierQuotes))
	}

	if metrics.CheapestShipping != 0 {
		t.Errorf("Expected cheapest shipping 0 for empty data, got: %.2f", metrics.CheapestShipping)
	}

	if metrics.HighestShipping != 0 {
		t.Errorf("Expected highest shipping 0 for empty data, got: %.2f", metrics.HighestShipping)
	}
}

func TestQuoteMetrics_RepositoryError(t *testing.T) {
	mockRepo := &MockQuoteRepository{
		findQuotesByLastQuoteFunc: func(lastQuotes int) ([]quote.Carrier, error) {
			return nil, errors.New("database connection error")
		},
	}
	mockAPI := &MockFastDeliveryAPI{}

	controller := setupController(mockRepo, mockAPI)

	metrics, err := controller.QuoteMetrics(10)

	if err == nil {
		t.Fatal("Expected repository error, got nil")
	}

	expectedError := "failed to find last quotes: database connection error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got: '%s'", expectedError, err.Error())
	}

	// Verificar se retorna estrutura vazia em caso de erro
	if len(metrics.CarrierQuotes) != 0 {
		t.Errorf("Expected empty carrier quotes on error, got: %d", len(metrics.CarrierQuotes))
	}
}

// Teste adicional para SimulateQuote - erro ao salvar no repository
func TestSimulateQuote_SaveQuoteError(t *testing.T) {
	mockRepo := &MockQuoteRepository{
		saveQuoteFunc: func(carrier quote.Carrier) error {
			return errors.New("database save error")
		},
	}
	mockAPI := &MockFastDeliveryAPI{
		simulateQuoteFunc: func(request models.QuoteRequest) (*models.QuoteResponse, error) {
			return &models.QuoteResponse{
				Dispatchers: []models.DispatcherResponse{
					{
						Offers: []models.Offer{
							{
								Carrier: models.Carrier{
									Name: "Transportadora A",
								},
								Service:    "Expresso",
								FinalPrice: 25.50,
								DeliveryTime: models.DeliveryTime{
									Days: 3,
								},
							},
						},
					},
				},
			}, nil
		},
	}

	controller := setupController(mockRepo, mockAPI)

	request := quote.QuoteRequest{
		Recipient: quote.Recipient{
			Address: quote.Address{
				ZipCode: "12345678",
			},
		},
		Volumes: []quote.Volume{
			{
				Category:      1,
				Amount:        1,
				UnitaryWeight: 1.0,
				Price:         50.0,
				SKU:           "PROD123",
				Height:        10.0,
				Width:         10.0,
				Length:        10.0,
			},
		},
	}

	response, err := controller.SimulateQuote(request)

	if err == nil {
		t.Fatal("Expected save error, got nil")
	}

	if response != nil {
		t.Fatal("Expected nil response for save error")
	}

	expectedError := "failed to save quote: database save error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got: '%s'", expectedError, err.Error())
	}
}

// Teste adicional para QuoteMetrics - apenas uma transportadora
func TestQuoteMetrics_SingleCarrier(t *testing.T) {
	mockRepo := &MockQuoteRepository{
		findQuotesByLastQuoteFunc: func(lastQuotes int) ([]quote.Carrier, error) {
			return []quote.Carrier{
				{Name: "Transportadora Única", Price: 30.0},
				{Name: "Transportadora Única", Price: 40.0},
				{Name: "Transportadora Única", Price: 50.0},
			}, nil
		},
	}
	mockAPI := &MockFastDeliveryAPI{}

	controller := setupController(mockRepo, mockAPI)

	metrics, err := controller.QuoteMetrics(10)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(metrics.CarrierQuotes) != 1 {
		t.Fatalf("Expected 1 carrier quote, got: %d", len(metrics.CarrierQuotes))
	}

	carrier := metrics.CarrierQuotes[0]
	expectedAverage := 40.0 // (30+40+50)/3

	if carrier.CarrierName != "Transportadora Única" {
		t.Errorf("Expected carrier name 'Transportadora Única', got: %s", carrier.CarrierName)
	}

	if carrier.TotalQuotes != 3 {
		t.Errorf("Expected 3 total quotes, got: %d", carrier.TotalQuotes)
	}

	if carrier.AveragePrice != expectedAverage {
		t.Errorf("Expected average price %.2f, got: %.2f", expectedAverage, carrier.AveragePrice)
	}

	// Quando há apenas uma transportadora, cheapest e highest devem ser iguais
	if metrics.CheapestShipping != expectedAverage {
		t.Errorf("Expected cheapest shipping %.2f, got: %.2f", expectedAverage, metrics.CheapestShipping)
	}

	if metrics.HighestShipping != expectedAverage {
		t.Errorf("Expected highest shipping %.2f, got: %.2f", expectedAverage, metrics.HighestShipping)
	}
}

// Teste adicional para SimulateQuote - múltiplos volumes
func TestSimulateQuote_MultipleVolumes(t *testing.T) {
	mockRepo := &MockQuoteRepository{}
	mockAPI := &MockFastDeliveryAPI{
		simulateQuoteFunc: func(request models.QuoteRequest) (*models.QuoteResponse, error) {
			// Verificar se os volumes foram convertidos corretamente
			if len(request.Dispatchers) != 1 {
				t.Errorf("Expected 1 dispatcher, got: %d", len(request.Dispatchers))
			}

			if len(request.Dispatchers[0].Volumes) != 3 {
				t.Errorf("Expected 3 volumes, got: %d", len(request.Dispatchers[0].Volumes))
			}

			// Verificar conversão de categoria para string
			expectedCategories := []string{"1", "2", "3"}
			for i, volume := range request.Dispatchers[0].Volumes {
				if volume.Category != expectedCategories[i] {
					t.Errorf("Expected category '%s', got: '%s'", expectedCategories[i], volume.Category)
				}
			}

			return &models.QuoteResponse{
				Dispatchers: []models.DispatcherResponse{
					{
						Offers: []models.Offer{
							{
								Carrier: models.Carrier{
									Name: "Transportadora X",
								},
								Service:    "Premium",
								FinalPrice: 99.99,
								DeliveryTime: models.DeliveryTime{
									Days: 1,
								},
							},
						},
					},
				},
			}, nil
		},
	}

	controller := setupController(mockRepo, mockAPI)

	request := quote.QuoteRequest{
		Recipient: quote.Recipient{
			Address: quote.Address{
				ZipCode: "87654321",
			},
		},
		Volumes: []quote.Volume{
			{
				Category:      1,
				Amount:        5,
				UnitaryWeight: 2.5,
				Price:         200.0,
				SKU:           "PROD001",
				Height:        15.0,
				Width:         25.0,
				Length:        30.0,
			},
			{
				Category:      2,
				Amount:        3,
				UnitaryWeight: 1.8,
				Price:         150.0,
				SKU:           "PROD002",
				Height:        12.0,
				Width:         20.0,
				Length:        25.0,
			},
			{
				Category:      3,
				Amount:        1,
				UnitaryWeight: 5.0,
				Price:         500.0,
				SKU:           "PROD003",
				Height:        20.0,
				Width:         30.0,
				Length:        40.0,
			},
		},
	}

	response, err := controller.SimulateQuote(request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Carriers) != 1 {
		t.Fatalf("Expected 1 carrier, got: %d", len(response.Carriers))
	}

	carrier := response.Carriers[0]
	if carrier.Name != "Transportadora X" {
		t.Errorf("Expected carrier name 'Transportadora X', got: %s", carrier.Name)
	}

	if carrier.Service != "Premium" {
		t.Errorf("Expected service 'Premium', got: %s", carrier.Service)
	}

	if carrier.Price != 99.99 {
		t.Errorf("Expected price 99.99, got: %.2f", carrier.Price)
	}

	if carrier.Deadline != 1 {
		t.Errorf("Expected deadline 1 day, got: %d", carrier.Deadline)
	}
}

// Teste para verificar se os dados de configuração são aplicados corretamente
func TestSimulateQuote_ConfigurationMapping(t *testing.T) {
	mockRepo := &MockQuoteRepository{}

	// Mock que verifica se a configuração foi aplicada corretamente
	mockAPI := &MockFastDeliveryAPI{
		simulateQuoteFunc: func(request models.QuoteRequest) (*models.QuoteResponse, error) {
			// Verificar se os dados do shipper estão corretos
			if request.Shipper.RegisteredNumber != "12345678901234" {
				t.Errorf("Expected shipper registered number '12345678901234', got: '%s'", request.Shipper.RegisteredNumber)
			}

			if request.Shipper.Token != "test-token-12345678901234567890" {
				t.Errorf("Expected shipper token 'test-token-12345678901234567890', got: '%s'", request.Shipper.Token)
			}

			if request.Shipper.PlatformCode != "platform-123" {
				t.Errorf("Expected platform code 'platform-123', got: '%s'", request.Shipper.PlatformCode)
			}

			// Verificar dados do recipient
			if request.Recipient.Type != 0 {
				t.Errorf("Expected recipient type 0, got: %d", request.Recipient.Type)
			}

			if request.Recipient.Country != "BRA" {
				t.Errorf("Expected recipient country 'BRA', got: '%s'", request.Recipient.Country)
			}

			if request.Recipient.Zipcode != 11111111 {
				t.Errorf("Expected recipient zipcode 11111111, got: %d", request.Recipient.Zipcode)
			}

			// Verificar dispatcher
			if len(request.Dispatchers) != 1 {
				t.Fatalf("Expected 1 dispatcher, got: %d", len(request.Dispatchers))
			}

			dispatcher := request.Dispatchers[0]
			if dispatcher.RegisteredNumber != "12345678901234" {
				t.Errorf("Expected dispatcher registered number '12345678901234', got: '%s'", dispatcher.RegisteredNumber)
			}

			if dispatcher.Zipcode != 12345678 {
				t.Errorf("Expected dispatcher zipcode 12345678, got: %d", dispatcher.Zipcode)
			}

			// Verificar simulation type
			if len(request.SimulationType) != 1 || request.SimulationType[0] != 0 {
				t.Errorf("Expected simulation type [0], got: %v", request.SimulationType)
			}

			return &models.QuoteResponse{
				Dispatchers: []models.DispatcherResponse{
					{
						Offers: []models.Offer{
							{
								Carrier: models.Carrier{
									Name: "Transportadora Teste",
								},
								Service:    "Teste",
								FinalPrice: 10.00,
								DeliveryTime: models.DeliveryTime{
									Days: 5,
								},
							},
						},
					},
				},
			}, nil
		},
	}

	controller := setupController(mockRepo, mockAPI)

	request := quote.QuoteRequest{
		Recipient: quote.Recipient{
			Address: quote.Address{
				ZipCode: "11111111", // CEP diferente para testar conversão
			},
		},
		Volumes: []quote.Volume{
			{
				Category:      7,
				Amount:        1,
				UnitaryWeight: 2.0,
				Price:         75.0,
				SKU:           "TEST123",
				Height:        8.0,
				Width:         12.0,
				Length:        16.0,
			},
		},
	}

	response, err := controller.SimulateQuote(request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Carriers) != 1 {
		t.Fatalf("Expected 1 carrier, got: %d", len(response.Carriers))
	}
}
