package api

import (
	"encoding/json"
	"net/http"

	"gateway/internal/audit"
	"gateway/internal/middleware"
)

type APIHandlers struct {
	dbClient *audit.DBClient
}

type TransferRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

func NewAPIHandlers(dbClient *audit.DBClient) *APIHandlers {
	return &APIHandlers{dbClient: dbClient}
}

// GetBalanceHandler retrieves the current balance for the authenticated tenant
func (h *APIHandlers) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := middleware.GetClaimsFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing claims context"})
		return
	}

	balance, err := h.dbClient.GetBalance(claims.TenantID, claims.Sub)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant_id": claims.TenantID,
		"sub":       claims.Sub,
		"role":      claims.Role,
		"balance":   balance,
	})
}

// CreateTransferHandler processes a new financial transaction under RLS context
func (h *APIHandlers) CreateTransferHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := middleware.GetClaimsFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing claims context"})
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if req.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "amount must be greater than zero"})
		return
	}

	tx, err := h.dbClient.CreateTransaction(claims.TenantID, req.Amount, req.Description, claims.Sub)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "success",
		"transaction": tx,
	})
}

// GetAuditLogsHandler retrieves the immutable audit log ledger for the authenticated tenant
func (h *APIHandlers) GetAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claims, ok := middleware.GetClaimsFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing claims context"})
		return
	}

	logs, err := h.dbClient.GetAuditLogs(claims.TenantID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant_id":  claims.TenantID,
		"audit_logs": logs,
	})
}
