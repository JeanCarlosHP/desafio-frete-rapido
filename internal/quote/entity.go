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
