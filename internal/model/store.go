package model

// Store is the top-level data structure persisted to disk.
type Store struct {
	Devices  []Device     `json:"devices"`
	Tokens   []Token      `json:"tokens"`
	AuditLog []AuditEntry `json:"audit_log"`
}
