package events

import "github.com/google/uuid"

const EventTypeUserRegistered = "user.registered"

type UserRegistered struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
}
