package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventTypeGitHubAccountLinked          = "github.account.linked"
	EventTypeGitHubAccountUnlinked        = "github.account.unlinked"
	EventTypeGitHubProfileUpdated         = "github.profile.updated"
	EventTypeGitHubHistoryImportCompleted = "github.history.import.completed"
	EventTypeGitHubHistoryImportFailed    = "github.history.import.failed"
	EventTypeGitHubCurrentYearRefreshed   = "github.current-year.refreshed"

	ReasonsGithubInitialLink    = "initial_link"
	ReasonsGithubBackgroundSync = "background_sync"
)

type GitHubAccountLinked struct {
	UserID         uuid.UUID `json:"user_id"`
	GitHubUserID   string    `json:"github_user_id"`
	GitHubUsername string    `json:"github_username"`
	LinkedAt       time.Time `json:"linked_at"`
}

type GitHubAccountUnlinked struct {
	UserID         uuid.UUID `json:"user_id"`
	GitHubUsername string    `json:"github_username,omitempty"`
	UnlinkedAt     time.Time `json:"unlinked_at"`
}

type GitHubProfileUpdated struct {
	UserID             uuid.UUID `json:"user_id"`
	GitHubUsername     string    `json:"github_username"`
	TotalContributions int       `json:"total_contributions"`
	CurrentYear        int       `json:"current_year"`
	UpdatedAt          time.Time `json:"updated_at"`
	// Optional: update reason
	Reason string `json:"reason,omitempty"` // "initial_link", "manual_refresh", "background_sync"...
}

type GitHubHistoryImportCompleted struct {
	UserID         uuid.UUID `json:"user_id"`
	GitHubUsername string    `json:"github_username"`
	YearsImported  []int     `json:"years_imported"`
	TotalYears     int       `json:"total_years"`
	CompletedAt    time.Time `json:"completed_at"`
}

type GitHubHistoryImportFailed struct {
	UserID         uuid.UUID `json:"user_id"`
	GitHubUsername string    `json:"github_username"`
	YearsAttempted []int     `json:"years_attempted"`
	Error          string    `json:"error"`
	ErrorCode      string    `json:"error_code,omitempty"` // "rate_limit", "unauthorized", "timeout"...
	FailedAt       time.Time `json:"failed_at"`
}

type GitHubCurrentYearRefreshed struct {
	UserID             uuid.UUID `json:"user_id"`
	GitHubUsername     string    `json:"github_username"`
	Year               int       `json:"year"`
	TotalContributions int       `json:"total_contributions"`
	DaysWithActivity   int       `json:"days_with_activity,omitempty"`
	RefreshedAt        time.Time `json:"refreshed_at"`
}
