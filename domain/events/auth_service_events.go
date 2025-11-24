package events

// AuthServiceUserEvent - Schema V2 (Minimal Identity Event)
// ตรงตาม EVENT_SCHEMA_V2.md จาก Auth Service
// เก็บเฉพาะ identity data (id, email, username) ไม่มี profile data
type AuthServiceUserEvent struct {
	// Minimal Identity Data
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Action   string `json:"action"` // "created" | "updated" | "deleted"

	// Observability Metadata
	RequestID   string `json:"request_id"`
	Timestamp   string `json:"timestamp"`    // ISO 8601 format
	ServiceName string `json:"service_name"` // "gofiber-auth"
	Sequence    uint64 `json:"sequence,omitempty"`
}

// IsValid validates the event
func (e *AuthServiceUserEvent) IsValid() bool {
	return e.ID != "" &&
		e.Email != "" &&
		e.Username != "" &&
		e.Action != ""
}
