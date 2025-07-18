package quote

import (
	"context"
	"fmt"

	"github.com/jeancarloshp/desafio-frete-rapido/pkg/database/querier"
)

type QuoteRepository struct {
	conn *querier.Queries
}

func NewQuoteRepository(conn *querier.Queries) *QuoteRepository {
	return &QuoteRepository{
		conn: conn,
	}
}

func (r *QuoteRepository) SaveQuote(carrier Carrier) error {
	_, err := r.conn.CreateQuote(context.Background(), querier.CreateQuoteParams{
		CarrierName: carrier.Name,
		Service:     carrier.Service,
		Price:       carrier.Price,
		Deadline:    carrier.Deadline,
	})
	if err != nil {
		return fmt.Errorf("failed to save quote: %w", err)
	}

	return nil
}

func (r *QuoteRepository) FindQuotesByLastQuote(lastQuote int) ([]Carrier, error) {
	quotes, err := r.conn.FindLastQuotes(context.Background(), lastQuote)
	if err != nil {
		return nil, fmt.Errorf("failed to find quotes: %w", err)
	}

	carriers := make([]Carrier, len(quotes))
	for i, q := range quotes {
		carriers[i] = Carrier{
			Name:     q.CarrierName,
			Service:  q.Service,
			Price:    q.Price,
			Deadline: q.Deadline,
		}
	}

	return carriers, nil
}
