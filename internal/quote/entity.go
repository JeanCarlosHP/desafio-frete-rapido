package quote

type QuoteRequest struct {
	Recipient Recipient `json:"recipient" validate:"required"`
	Volumes   []Volume  `json:"volumes" validate:"required,dive,required"`
}

type Recipient struct {
	Address Address `json:"address" validate:"required"`
}

type Address struct {
	ZipCode string `json:"zipcode" validate:"required,len=8"`
}

type Volume struct {
	Category      int     `json:"category" validate:"required"`
	Amount        int     `json:"amount" validate:"required"`
	UnitaryWeight float64 `json:"unitary_weight" validate:"required"`
	Price         float64 `json:"price" validate:"required"`
	SKU           string  `json:"sku" validate:"required"`
	Height        float64 `json:"height" validate:"required"`
	Width         float64 `json:"width" validate:"required"`
	Length        float64 `json:"length" validate:"required"`
}

type QuoteResponse struct {
	Carriers []Carrier `json:"carriers"`
}

type Carrier struct {
	Name     string  `json:"name"`
	Service  string  `json:"service"`
	Deadline int     `json:"deadline"`
	Price    float64 `json:"price"`
}

type QuoteMetrics struct {
	CarrierQuotes    []CarrierQuotes `json:"carrier_quotes"`
	CheapestShipping float64         `json:"cheapest_shipping"`
	HighestShipping  float64         `json:"highest_shipping"`
}

type CarrierQuotes struct {
	CarrierName  string  `json:"carrier_name"`
	TotalQuotes  int     `json:"total_quotes"`
	TotalPrice   float64 `json:"total_price"`
	AveragePrice float64 `json:"average_price"`
}
