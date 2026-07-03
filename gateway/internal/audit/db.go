package audit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DBClient struct {
	db *sql.DB
}

type Transaction struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditLog struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	ActorID   string    `json:"actor_id"`
	TenantID  string    `json:"tenant_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details"`
	PrevHash  string    `json:"prev_hash"`
	BlockHash string    `json:"block_hash"`
}

// Map client identities to valid UUIDs to satisfy database constraints
var ClientToUserUUID = map[string]string{
	"client-alice": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
	"client-bob":   "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
	"client-evil":  "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee",
}

func GetUserUUID(sub string) string {
	if uuid, ok := ClientToUserUUID[sub]; ok {
		return uuid
	}
	return "00000000-0000-0000-0000-000000000000"
}

// NewDBClient initializes and checks a PostgreSQL connection pool
func NewDBClient(connStr string) (*DBClient, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection limits
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping database to verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DBClient{db: db}, nil
}

// Close closes the database connection pool
func (c *DBClient) Close() {
	c.db.Close()
}

// CreateTransaction inserts a new financial transaction under RLS context
func (c *DBClient) CreateTransaction(tenantID string, amount float64, description string, actorSub string) (*Transaction, error) {
	// Start transaction to bind SET LOCAL lifetime to this transaction only
	tx, err := c.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start database transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Inject Tenant Context for RLS
	_, err = tx.Exec("SELECT set_config('app.tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set RLS tenant context: %w", err)
	}

	// 2. Perform Transaction Insertion
	var t Transaction
	query := "INSERT INTO transactions (tenant_id, amount, description) VALUES ($1, $2, $3) RETURNING id, created_at"
	err = tx.QueryRow(query, tenantID, amount, description).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("transaction insert blocked or failed: %w", err)
	}

	t.TenantID = tenantID
	t.Amount = amount
	t.Description = description

	// 3. Write WORM Audit Log inside the same transaction
	actorUUID := GetUserUUID(actorSub)
	detailsMap := map[string]interface{}{
		"amount":      amount,
		"description": description,
		"tx_id":       t.ID,
	}
	detailsBytes, _ := json.Marshal(detailsMap)

	auditQuery := "INSERT INTO audit_logs (actor_id, tenant_id, action, resource, details) VALUES ($1, $2, $3, $4, $5)"
	_, err = tx.Exec(auditQuery, actorUUID, tenantID, "CREATE_TRANSACTION", "transactions", string(detailsBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to write audit log: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &t, nil
}

// GetBalance retrieves the total balance of a tenant under RLS context
func (c *DBClient) GetBalance(tenantID string, actorSub string) (float64, error) {
	tx, err := c.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start database transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Inject Tenant Context for RLS
	_, err = tx.Exec("SELECT set_config('app.tenant_id', $1, true)", tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to set RLS tenant context: %w", err)
	}

	// 2. Run query
	var balance float64
	query := "SELECT COALESCE(SUM(amount), 0) FROM transactions"
	err = tx.QueryRow(query).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to query balance: %w", err)
	}

	// 3. Write WORM Audit Log
	actorUUID := GetUserUUID(actorSub)
	auditQuery := "INSERT INTO audit_logs (actor_id, tenant_id, action, resource, details) VALUES ($1, $2, $3, $4, $5)"
	_, err = tx.Exec(auditQuery, actorUUID, tenantID, "GET_BALANCE", "transactions", "{}")
	if err != nil {
		return 0, fmt.Errorf("failed to write audit log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return balance, nil
}

// GetAuditLogs retrieves all audit logs for the tenant
func (c *DBClient) GetAuditLogs(tenantID string) ([]AuditLog, error) {
	tx, err := c.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start database transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Inject Tenant Context for RLS
	_, err = tx.Exec("SELECT set_config('app.tenant_id', $1, true)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set RLS tenant context: %w", err)
	}

	// 2. Query logs
	query := "SELECT id, timestamp, actor_id, tenant_id, action, resource, details, prev_hash, block_hash FROM audit_logs ORDER BY id ASC"
	rows, err := tx.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		var detailsRaw []byte
		err := rows.Scan(&l.ID, &l.Timestamp, &l.ActorID, &l.TenantID, &l.Action, &l.Resource, &detailsRaw, &l.PrevHash, &l.BlockHash)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log row: %w", err)
		}
		l.Details = string(detailsRaw)
		logs = append(logs, l)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return logs, nil
}
