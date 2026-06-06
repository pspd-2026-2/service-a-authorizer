package authorizer

import (
	"context"
	"log"
	"service-a/internal/models"
	"service-a/internal/pb"
	"time"
)

type GRPCHandler struct {
	authorizer *Authorizer
}

func NewGRPCHandler(a *Authorizer) *GRPCHandler {
	return &GRPCHandler{
		authorizer: a,
	}
}

func (h *GRPCHandler) Authorize(ctx context.Context, req *pb.AuthorizationRequest) (*pb.AuthorizationResponse, error) {
	if req == nil {
		return nil, nil
	}

	start := time.Now()
	output := h.authorizer.AuthorizeTransaction(models.AuthorizeTransactionInput{
		Request: models.AuthorizationRequest{
			UserID:    req.UserId,
			Amount:    req.Amount,
			CardNumber: req.CardNumber,
			IPAddress: req.IpAddress,
		},
	})
	elapsed := time.Since(start)

	log.Printf("[AUTHORIZE] transactionId=%s status=%s userId=%s amount=%.2f duration=%s",
		output.Response.TransactionID,
		output.Response.Status,
		req.UserId,
		req.Amount,
		elapsed)

	resp := &pb.AuthorizationResponse{
		TransactionId:   output.Response.TransactionID,
		AuthCode:        output.Response.AuthCode,
		Status:          output.Response.Status,
		CardStatus:      output.Response.CardStatus,
		ApprovedLimit:   output.Response.ApprovedLimit,
		RequestedAmount: output.Response.RequestedAmount,
		LimitSufficient: output.Response.LimitSufficient,
		Message:         output.Response.Message,
	}

	return resp, output.Error
}

