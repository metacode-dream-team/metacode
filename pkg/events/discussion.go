package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeDiscussionCreated = "discussion.created"
)

type DiscussionCreated struct {
	ID         uuid.UUID `json:"id"`
	AuthorID   uuid.UUID `json:"author_id"`
	PreviewURL *string   `json:"preview_url"`
	CreatedAt  time.Time `json:"created_at"`
}
