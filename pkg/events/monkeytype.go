package events

import (
	"time"

	"github.com/google/uuid"
)

// -----------------------------------------------------------------------------
// Event Type Constants
// -----------------------------------------------------------------------------

const (
	EventTypeMonkeytypeAccountBound = "monkeytype.account.bound"

	EventTypeMonkeytypeAccountUnbound = "monkeytype.account.unbound"

	EventTypeMonkeytypeVerificationSucceeded = "monkeytype.verification.succeeded"

	EventTypeMonkeytypeVerificationFailed = "monkeytype.verification.failed"

	EventTypeMonkeytypeProfileUpdated = "monkeytype.profile.updated"

	EventTypeMonkeytypeCurrentStatsRefreshed = "monkeytype.current-stats.refreshed"
)

// -----------------------------------------------------------------------------
// Event Payloads
// -----------------------------------------------------------------------------

// MonkeytypeAccountBound
// Fired right after successful BindMonkeytype call (before verification)
type MonkeytypeAccountBound struct {
	UserID             uuid.UUID `json:"user_id"`
	MonkeytypeUsername string    `json:"monkeytype_username"`
	BoundAt            time.Time `json:"bound_at"`
	Verified           bool      `json:"verified"` // usually false at this stage
}

// MonkeytypeAccountUnbound
// Fired after UnbindMonkeytype
type MonkeytypeAccountUnbound struct {
	UserID             uuid.UUID `json:"user_id"`
	MonkeytypeUsername string    `json:"monkeytype_username,omitempty"`
	UnboundAt          time.Time `json:"unbound_at"`
	Reason             string    `json:"reason,omitempty"` // "manual", "verification_timeout", "user_requested", ...
}

// MonkeytypeVerificationSucceeded
// Fired when token is found in bio and Keycloak attribute is updated
type MonkeytypeVerificationSucceeded struct {
	UserID             uuid.UUID `json:"user_id"`
	MonkeytypeUsername string    `json:"monkeytype_username"`
	VerifiedAt         time.Time `json:"verified_at"`
}

// MonkeytypeVerificationFailed
// Fired on timeout, manual cancel, bio fetch error, Keycloak failure, etc
type MonkeytypeVerificationFailed struct {
	UserID             uuid.UUID `json:"user_id"`
	MonkeytypeUsername string    `json:"monkeytype_username,omitempty"`
	FailedAt           time.Time `json:"failed_at"`
	Reason             string    `json:"reason"` // "timeout", "cancelled", "bio_error", "keycloak_error", ...
	ErrorCode          string    `json:"error_code,omitempty"`
}

// MonkeytypeProfileUpdated
// Fired after successful profile fetch & persistence (initial bind or force refresh)
type MonkeytypeProfileUpdated struct {
	UserID             uuid.UUID `json:"user_id"`
	MonkeytypeUsername string    `json:"monkeytype_username"`
	WpmBest            float64   `json:"wpm_best,omitempty"`
	AccuracyBest       float64   `json:"accuracy_best,omitempty"`
	TestsCompleted     int       `json:"tests_completed,omitempty"`
	Verified           bool      `json:"verified"`
	UpdatedAt          time.Time `json:"updated_at"`
	ChangeReason       string    `json:"change_reason,omitempty"` // "initial_bind", "force_refresh", "background_sync"
}

// MonkeytypeCurrentStatsRefreshed
// Fired when current typing stats are refreshed (can be called periodically / on demand)
type MonkeytypeCurrentStatsRefreshed struct {
	UserID             uuid.UUID `json:"user_id"`
	MonkeytypeUsername string    `json:"monkeytype_username"`
	TestsToday         int       `json:"tests_today,omitempty"`
	RefreshedAt        time.Time `json:"refreshed_at"`
}
