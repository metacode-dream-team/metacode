package events

import (
	"time"

	"github.com/google/uuid"
)

// ────────────────────────────────────────────────
// Константы имён событий (Event Types)
// ────────────────────────────────────────────────

const (
	EventTypeLeetCodeAccountBound           = "leetcode.account.bound"
	EventTypeLeetCodeAccountUnbound         = "leetcode.account.unbound"
	EventTypeLeetCodeVerificationSucceeded  = "leetcode.verification.succeeded"
	EventTypeLeetCodeVerificationFailed     = "leetcode.verification.failed"
	EventTypeLeetCodeProfileUpdated         = "leetcode.profile.updated"
	EventTypeLeetCodeHistoryImportCompleted = "leetcode.history.import.completed"
	EventTypeLeetCodeHistoryImportFailed    = "leetcode.history.import.failed"
	EventTypeLeetCodeCurrentYearRefreshed   = "leetcode.current-year.refreshed"
)

// ────────────────────────────────────────────────
// Payload-structures of events
// ────────────────────────────────────────────────

type LeetCodeAccountBound struct {
	UserID           uuid.UUID `json:"user_id"`
	LeetCodeUsername string    `json:"leetcode_username"`
	BoundAt          time.Time `json:"bound_at"`
	Verified         bool      `json:"verified"`
}

type LeetCodeAccountUnbound struct {
	UserID           uuid.UUID `json:"user_id"`
	LeetCodeUsername string    `json:"leetcode_username,omitempty"`
	UnboundAt        time.Time `json:"unbound_at"`
	Reason           string    `json:"reason,omitempty"` // "manual", "verification_timeout", ...
}

type LeetCodeVerificationSucceeded struct {
	UserID           uuid.UUID `json:"user_id"`
	LeetCodeUsername string    `json:"leetcode_username"`
	VerifiedAt       time.Time `json:"verified_at"`
}

type LeetCodeVerificationFailed struct {
	UserID           uuid.UUID `json:"user_id"`
	LeetCodeUsername string    `json:"leetcode_username,omitempty"`
	FailedAt         time.Time `json:"failed_at"`
	Reason           string    `json:"reason"` // "timeout", "token_not_found", "bio_fetch_failed", "persistence_error", ...
	ErrorCode        string    `json:"error_code,omitempty"`
}

type LeetCodeProfileUpdated struct {
	UserID           uuid.UUID `json:"user_id"`
	LeetCodeUsername string    `json:"leetcode_username"`
	TotalSolved      int       `json:"total_solved"`
	Verified         bool      `json:"verified"`
	UpdatedAt        time.Time `json:"updated_at"`
	ChangeReason     string    `json:"change_reason,omitempty"` // "initial_bind", "force_refresh", "background_sync"
}

type LeetCodeHistoryImportCompleted struct {
	UserID           uuid.UUID `json:"user_id"`
	LeetCodeUsername string    `json:"leetcode_username"`
	YearsImported    []int     `json:"years_imported"`
	TotalYears       int       `json:"total_years"`
	CompletedAt      time.Time `json:"completed_at"`
}

type LeetCodeHistoryImportFailed struct {
	UserID           uuid.UUID `json:"user_id"`
	LeetCodeUsername string    `json:"leetcode_username"`
	YearsAttempted   []int     `json:"years_attempted"`
	Error            string    `json:"error"`
	ErrorCode        string    `json:"error_code,omitempty"` // "rate_limit", "not_found", "timeout"
	FailedAt         time.Time `json:"failed_at"`
}

type LeetCodeCurrentYearRefreshed struct {
	UserID           uuid.UUID `json:"user_id"`
	LeetCodeUsername string    `json:"leetcode_username"`
	Year             int       `json:"year"`
	QuestionsSolved  int       `json:"questions_solved"`
	ActiveDays       int       `json:"active_days,omitempty"`
	RefreshedAt      time.Time `json:"refreshed_at"`
}
