package authorizer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"service-a/internal/models"
	"strings"
	"time"
)

type Authorizer struct {
	repo *cardRepository
}

func New() *Authorizer {
	return &Authorizer{
		repo: newCardRepository(),
	}
}

func (a *Authorizer) AuthorizeTransaction(input models.AuthorizeTransactionInput) models.AuthorizeTransactionOutput {
	req := input.Request

	time.Sleep(80 * time.Millisecond)

	card, err := a.repo.findCard(req.CardNumber)
	if err != nil {
		return models.AuthorizeTransactionOutput{
			Response: models.AuthorizationResponse{
				TransactionID: 		generateTransactionID(),
				Status:				"declined",
				CardStatus:     	"unknown",
				RequestedAmount:	req.Amount,
				Message:			fmt.Sprintf("card not found: %s", req.CardNumber),
			},
		}
	}

	txId := generateTransactionID()
	availableLimit := card.availableLimit()

	if card.Status != "active" {
		reason := cardStatusReason(card.Status)
		return models.AuthorizeTransactionOutput{
			Response: models.AuthorizationResponse{
				TransactionID:   txId,
				Status:          "declined",
				CardStatus:      card.Status,
				ApprovedLimit:   availableLimit,
				RequestedAmount: req.Amount,
				LimitSufficient: availableLimit >= req.Amount,
				Message:         reason,
			},
		}
	}

	if req.Amount <= 0 {
		return models.AuthorizeTransactionOutput{
			Response: models.AuthorizationResponse{
				TransactionID:   txId,
				Status:          "declined",
				CardStatus:      card.Status,
				ApprovedLimit:   availableLimit,
				RequestedAmount: req.Amount,
				LimitSufficient: false,
				Message:         "invalid transaction amount",
			},
		}
	}

	if availableLimit < req.Amount {
		return models.AuthorizeTransactionOutput{
			Response: models.AuthorizationResponse{
				TransactionID:   txId,
				Status:          "declined",
				CardStatus:      card.Status,
				ApprovedLimit:   availableLimit,
				RequestedAmount: req.Amount,
				LimitSufficient: false,
				Message:         fmt.Sprintf("insufficient limit: available %.2f, requested %.2f", availableLimit, req.Amount),
			},
		}
	}

	authCode, err := generateAuthCode()
	if err != nil {
		return models.AuthorizeTransactionOutput{
			Response: models.AuthorizationResponse{
				TransactionID:   txId,
				Status:          "declined",
				CardStatus:      card.Status,
				RequestedAmount: req.Amount,
				Message:         "internal error generating auth code",
			},
			Error: err,
		}
	}

	return models.AuthorizeTransactionOutput{
		Response: models.AuthorizationResponse{
			TransactionID:   txId,
			AuthCode:        authCode,
			Status:          "authorized",
			CardStatus:      card.Status,
			ApprovedLimit:   availableLimit,
			RequestedAmount: req.Amount,
			LimitSufficient: true,
			Message:         "transaction authorized successfully",
		},
	}
}

func generateTransactionID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("tx_%s_%d", hex.EncodeToString(b), time.Now().UnixMilli()%100000)
}

func generateAuthCode() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate auth code: %w", err)
	}
	return strings.ToUpper(hex.EncodeToString(b)), nil
}

func cardStatusReason(status string) string {
	switch status {
	case "blocked":
		return "card is blocked"
	case "expired":
		return "card is expired"
	default:
		return fmt.Sprintf("card status not eligible for transactions: %s", status)
	}
}