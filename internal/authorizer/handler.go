package authorizer

import (
	"encoding/json"
	"log"
	"net/http"
	"service-a/internal/models"
	"time"
)

type Handler struct {
	authorizer *Authorizer
}

func NewHandler(a *Authorizer) *Handler {
	return &Handler{authorizer: a}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/authorize", h.handleAuthorize)
	mux.HandleFunc("/health", h.handleHealth)
}

func (h *Handler) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
 
	var req models.AuthorizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	defer r.Body.Close()
 
	if req.UserID == "" || req.CardNumber == "" {
		writeError(w, http.StatusBadRequest, "userId and cardNumber are required")
		return
	}
 
	start := time.Now()
	output := h.authorizer.AuthorizeTransaction(models.AuthorizeTransactionInput{
		Request: req,
	})
	elapsed := time.Since(start)
 
	log.Printf("[AUTHORIZE] transactionId=%s status=%s userId=%s amount=%.2f duration=%s",
		output.Response.TransactionID,
		output.Response.Status,
		req.UserID,
		req.Amount,
		elapsed,
	)
 
	statusCode := http.StatusOK
	if output.Response.Status == "declined" {
		statusCode = http.StatusUnprocessableEntity // 422
	}
 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(output.Response)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok","service":"authorization-service"}`))
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}