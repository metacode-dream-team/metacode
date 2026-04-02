package events

import "github.com/google/uuid"

const (
	EventTypeAchievementGranted = "achievement.granted"
)

type AchievementGrantedEvent struct {
	UserID        uuid.UUID `json:"user_id"`
	AchievementID uuid.UUID `json:"achievement_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	IconURL       *string   `json:"icon_url"`
}
