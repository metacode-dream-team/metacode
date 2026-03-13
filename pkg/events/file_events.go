package events

import "github.com/google/uuid"

const (
	EventTypeAvatarUpdatedEvent            = "avatar.updated"
	EventTypeAvatarProcessingFinishedEvent = "avatar.processing.finished"
)

type AvatarUpdatedEvent struct {
	UserID        uuid.UUID `json:"user_id"`
	S3OriginalUrl string    `json:"s3_original_url"`
}

type AvatarProcessingFinishedEvent struct {
	UserID      uuid.UUID `json:"user_id"`
	S3SmallUrl  string    `json:"s3_small_url"`
	S3MediumUrl string    `json:"s3_medium_url"`
	S3LargeUrl  string    `json:"s3_large_url"`
}
