package events

import "github.com/google/uuid"

const (
	EventTypeUserRegistered = "user.registered"
	EventTypeUserUpdated    = "user.updated"
	EventTypeUserDeleted    = "user.deleted"
)

type UserRegistered struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}

type UserUpdated struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url"`
}

type UserDeleted struct {
	UserID uuid.UUID `json:"user_id"`
}
