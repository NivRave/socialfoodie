package db

import (
	"context"
	"fmt"
)

// InsertAuditLog inserts an audit log into the database.
func (db *DB) InsertAuditLog(ctx context.Context, traceID, eventType, eventSource string, details map[string]interface{}) error {
	query := `
		INSERT INTO audit_logs (trace_id, event_type, event_source, details)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Pool.Exec(ctx, query, traceID, eventType, eventSource, details)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}
