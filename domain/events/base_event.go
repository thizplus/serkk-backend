package events

import "time"

// BaseEvent contains common fields for all events
type BaseEvent struct {
	EventID         string                 `json:"eventId"`         // Unique event ID (UUID)
	EventType       string                 `json:"eventType"`       // e.g., "user.created"
	EventVersion    string                 `json:"eventVersion"`    // e.g., "v1"
	Timestamp       time.Time              `json:"timestamp"`       // When the event occurred
	ProducerService string                 `json:"producerService"` // Service that produced this event
	SequenceNum     int64                  `json:"sequenceNum,omitempty"` // Sequence number for ordering (optional)
	Metadata        map[string]interface{} `json:"metadata,omitempty"`    // Additional metadata
}

// GetEventID returns the event ID
func (e BaseEvent) GetEventID() string {
	return e.EventID
}

// GetEventType returns the event type
func (e BaseEvent) GetEventType() string {
	return e.EventType
}

// GetEventVersion returns the event version
func (e BaseEvent) GetEventVersion() string {
	return e.EventVersion
}

// GetTimestamp returns the event timestamp
func (e BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetSequenceNum returns the sequence number
func (e BaseEvent) GetSequenceNum() int64 {
	return e.SequenceNum
}
