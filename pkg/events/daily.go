package events

import (
	"github.com/google/uuid"
)

type Source string

const (
	SourceGithub     Source = "github"
	SourceLeetcode   Source = "leetcode"
	SourceMonkeytype Source = "monkeytype"
)

type TodayContributedEvent struct {
	UserID uuid.UUID `json:"user_id"`
	Source Source    `json:"source"`
	Count  int       `json:"count"`
}
