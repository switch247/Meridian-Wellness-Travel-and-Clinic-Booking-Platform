package domain

import "time"

type User struct {
	ID                int64      `json:"id"`
	Username          string     `json:"username"`
	PasswordHash      string     `json:"-"`
	FailedAttempts    int        `json:"-"`
	LockedUntil       *time.Time `json:"-"`
	EncryptedPhone    string     `json:"-"`
	EncryptedAddress  string     `json:"-"`
	CreatedAt         time.Time  `json:"createdAt"`
	LastPasswordReset *time.Time `json:"lastPasswordReset,omitempty"`
}

type Role string

const (
	RoleTraveler  Role = "traveler"
	RoleCoach     Role = "coach"
	RoleOps       Role = "operations"
	RoleAdmin     Role = "admin"
	RoleClinician Role = "clinician"
)

type Address struct {
	Line1         string `json:"line1"`
	Line2         string `json:"line2"`
	City          string `json:"city"`
	State         string `json:"state"`
	PostalCode    string `json:"postalCode"`
	NormalizedKey string `json:"normalizedKey"`
	InCoverage    bool   `json:"inCoverage"`
	Duplicate     bool   `json:"duplicate"`
}

type PermissionAudit struct {
	ID         int64     `json:"id"`
	ActorID    int64     `json:"actorId"`
	TargetID   int64     `json:"targetId"`
	Action     string    `json:"action"`
	BeforeJSON string    `json:"before"`
	AfterJSON  string    `json:"after"`
	CreatedAt  time.Time `json:"createdAt"`
}
