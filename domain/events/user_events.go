package events

// UserCreatedEvent is published when a user is created
type UserCreatedEvent struct {
	BaseEvent
	Data UserCreatedData `json:"data"`
}

// UserCreatedData contains the data for user.created event
type UserCreatedData struct {
	UserID      string `json:"userId"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	Role        string `json:"role"`
	IsActive    bool   `json:"isActive"`
}

// UserUpdatedEvent is published when a user is updated
type UserUpdatedEvent struct {
	BaseEvent
	Data UserUpdatedData `json:"data"`
}

// UserUpdatedData contains the data for user.updated event
// Fields are pointers to distinguish between "not changed" (nil) and "changed to empty" ("")
type UserUpdatedData struct {
	UserID      string  `json:"userId"`
	Email       *string `json:"email,omitempty"`
	Username    *string `json:"username,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
	Avatar      *string `json:"avatar,omitempty"`
	Role        *string `json:"role,omitempty"`
	IsActive    *bool   `json:"isActive,omitempty"`
}

// UserDeletedEvent is published when a user is deleted
type UserDeletedEvent struct {
	BaseEvent
	Data UserDeletedData `json:"data"`
}

// UserDeletedData contains the data for user.deleted event
type UserDeletedData struct {
	UserID string `json:"userId"`
}

// UserPasswordChangedEvent is published when a user changes their password
type UserPasswordChangedEvent struct {
	BaseEvent
	Data UserPasswordChangedData `json:"data"`
}

// UserPasswordChangedData contains the data for user.password_changed event
type UserPasswordChangedData struct {
	UserID string `json:"userId"`
}
