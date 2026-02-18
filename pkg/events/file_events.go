package events

import "github.com/google/uuid"

const (
	EventTypeAvatarUpdatedEvent            = "avatar.avatar.updated"
	EventTypeAvatarProcessingFinishedEvent = "avatar.avatar.processing.finished"
)

type AvatarUpdatedEvent struct {
	UserID        uuid.UUID `json:"user_id"`
	S3OriginalKey string    `json:"s3_original_key"`
}

type AvatarProcessingFinishedEvent struct {
	UserID      uuid.UUID `json:"user_id"`
	S3SmallKey  string    `json:"s3_small_key"`
	S3MediumKey string    `json:"s3_medium_key"`
	S3LargeKey  string    `json:"s3_large_key"`
}
