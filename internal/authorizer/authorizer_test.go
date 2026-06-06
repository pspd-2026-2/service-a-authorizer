package authorizer_test

import (
	"service-a/internal/authorizer"
	"service-a/internal/models"
	"testing"
)

func newAuthorizer() *authorizer.Authorizer {
	return authorizer.New()
}

func TestAuthorizeTransaction_Approved(t *testing.T) {
	svc := newAuthorizer()

	out := svc.AuthorizeTransaction(models.AuthorizeTransactionInput{
		Request: models.AuthorizationRequest{
			UserID:     "user-001",
			Amount:     500.00,
			CardNumber: "4111111111111111", // active, limit R$10.000
			IPAddress:  "192.168.1.1",
		},
	})

	if out.Response.Status != "authorized" {
		t.Errorf("expected authorized, got %s — message: %s", out.Response.Status, out.Response.Message)
	}
	if out.Response.AuthCode == "" {
		t.Error("expected non-empty authCode on approval")
	}
	if !out.Response.LimitSufficient {
		t.Error("expected limitSufficient=true")
	}
}

func TestAuthorizeTransaction_BlockedCard(t *testing.T) {
	svc := newAuthorizer()

	out := svc.AuthorizeTransaction(models.AuthorizeTransactionInput{
		Request: models.AuthorizationRequest{
			UserID:     "user-002",
			Amount:     100.00,
			CardNumber: "4000000000000002", // blocked
			IPAddress:  "10.0.0.1",
		},
	})

	if out.Response.Status != "declined" {
		t.Errorf("expected declined, got %s", out.Response.Status)
	}
	if out.Response.CardStatus != "blocked" {
		t.Errorf("expected cardStatus=blocked, got %s", out.Response.CardStatus)
	}
}

func TestAuthorizeTransaction_ExpiredCard(t *testing.T) {
	svc := newAuthorizer()

	out := svc.AuthorizeTransaction(models.AuthorizeTransactionInput{
		Request: models.AuthorizationRequest{
			UserID:     "user-003",
			Amount:     50.00,
			CardNumber: "4000000000000069", // expired
			IPAddress:  "10.0.0.2",
		},
	})

	if out.Response.Status != "declined" {
		t.Errorf("expected declined, got %s", out.Response.Status)
	}
	if out.Response.CardStatus != "expired" {
		t.Errorf("expected cardStatus=expired, got %s", out.Response.CardStatus)
	}
}

func TestAuthorizeTransaction_InsufficientLimit(t *testing.T) {
	svc := newAuthorizer()

	out := svc.AuthorizeTransaction(models.AuthorizeTransactionInput{
		Request: models.AuthorizationRequest{
			UserID:     "user-004",
			Amount:     9000.00, // limite disponível é R$8.000 (10.000 - 2.000)
			CardNumber: "4111111111111111",
			IPAddress:  "10.0.0.3",
		},
	})

	if out.Response.Status != "declined" {
		t.Errorf("expected declined, got %s", out.Response.Status)
	}
	if out.Response.LimitSufficient {
		t.Error("expected limitSufficient=false")
	}
}

func TestAuthorizeTransaction_CardNotFound(t *testing.T) {
	svc := newAuthorizer()

	out := svc.AuthorizeTransaction(models.AuthorizeTransactionInput{
		Request: models.AuthorizationRequest{
			UserID:     "user-005",
			Amount:     100.00,
			CardNumber: "9999999999999999", // não existe
			IPAddress:  "10.0.0.4",
		},
	})

	if out.Response.Status != "declined" {
		t.Errorf("expected declined, got %s", out.Response.Status)
	}
	if out.Response.CardStatus != "unknown" {
		t.Errorf("expected cardStatus=unknown, got %s", out.Response.CardStatus)
	}
}

func TestAuthorizeTransaction_InvalidAmount(t *testing.T) {
	svc := newAuthorizer()

	out := svc.AuthorizeTransaction(models.AuthorizeTransactionInput{
		Request: models.AuthorizationRequest{
			UserID:     "user-006",
			Amount:     0,
			CardNumber: "4111111111111111",
			IPAddress:  "10.0.0.5",
		},
	})

	if out.Response.Status != "declined" {
		t.Errorf("expected declined for zero amount, got %s", out.Response.Status)
	}
}

func TestAuthorizeTransaction_MaskedCardNumber(t *testing.T) {
	svc := newAuthorizer()

	// Testa o match via prefixo+sufixo (formato mascarado do gateway P)
	out := svc.AuthorizeTransaction(models.AuthorizeTransactionInput{
		Request: models.AuthorizationRequest{
			UserID:     "user-007",
			Amount:     100.00,
			CardNumber: "4111...1111",
			IPAddress:  "10.0.0.6",
		},
	})

	if out.Response.Status != "authorized" {
		t.Errorf("expected authorized with masked card, got %s — %s", out.Response.Status, out.Response.Message)
	}
}