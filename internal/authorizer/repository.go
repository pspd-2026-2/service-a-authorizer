package authorizer

import (
	"fmt"
	"strings"
	"sync"
)

type CardRecord struct {
	MaskedNumber	string
	Status			string
	CreditLimit		float64
	CurrentDebt		float64
}

type cardRepository struct {
	mu		sync.RWMutex
	cards	map[string]*CardRecord
}

func newCardRepository() *cardRepository {
	return &cardRepository{
		cards: map[string]*CardRecord{
			"4111111111111111": {MaskedNumber: "4111...1111", Status: "active",  CreditLimit: 10000.00, CurrentDebt: 2000.00},
			"4000000000000002": {MaskedNumber: "4000...0002", Status: "blocked", CreditLimit: 5000.00,  CurrentDebt: 5000.00},
			"4000000000000069": {MaskedNumber: "4000...0069", Status: "expired", CreditLimit: 3000.00,  CurrentDebt: 0.00},
			"5500000000000004": {MaskedNumber: "5500...0004", Status: "active",  CreditLimit: 20000.00, CurrentDebt: 500.00},
		},
	}
}

func (r *cardRepository) findCard(rawNumber string) (*CardRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalized := strings.ReplaceAll(rawNumber, " ", "")
	
	if card, ok := r.cards[normalized]; ok {
		return card, nil
	}

	if idx := strings.Index(normalized, "..."); idx != -1 {
		prefix := normalized[:idx]
		suffix := normalized[idx+3:]
		
		for pan, card := range r.cards {
			if strings.HasPrefix(pan, prefix) && strings.HasSuffix(pan, suffix) {
				return card, nil
			}
		}
	}

	if len(normalized) >= 10 {
		prefix := normalized[:6]
		suffix := normalized[len(normalized)-4:]

		for pan, card := range r.cards {
			if strings.HasPrefix(pan, prefix) && strings.HasSuffix(pan, suffix) {
				return card, nil
			}
		}
	}

	return nil, fmt.Errorf("card not found %s", rawNumber)
}

func (r *CardRecord) availableLimit() float64 {
	return r.CreditLimit - r.CurrentDebt
}