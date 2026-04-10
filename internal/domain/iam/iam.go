package iam

import (
	"errors"
	"time"
)

var (
	ErrInvalidInput            = errors.New("invalid input")
	ErrConflict                = errors.New("conflict")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUnauthenticated         = errors.New("unauthenticated")
	ErrForbidden               = errors.New("forbidden")
	ErrInvitationNotFound      = errors.New("invitation not found")
	ErrInvitationExpired       = errors.New("invitation expired")
	ErrInvitationRevoked       = errors.New("invitation revoked")
	ErrInvitationAccepted      = errors.New("invitation already accepted")
	ErrInvitationEmailMismatch = errors.New("invitation email mismatch")
	ErrTenantNotAccessible     = errors.New("tenant not accessible")
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

func (r Role) Valid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	default:
		return false
	}
}

func (r Role) CanInvite(target Role) bool {
	switch r {
	case RoleOwner:
		return target == RoleAdmin || target == RoleMember
	case RoleAdmin:
		return target == RoleMember
	default:
		return false
	}
}

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Tenant struct {
	ID         string
	Slug       string
	Name       string
	IsPersonal bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Membership struct {
	ID        string
	UserID    string
	TenantID  string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MembershipWithTenant struct {
	Membership Membership
	Tenant     Tenant
}

type TenantMember struct {
	UserID       string
	Email        string
	MembershipID string
	Role         Role
	CreatedAt    time.Time
}

type Invitation struct {
	ID              string
	TenantID        string
	Email           string
	Role            Role
	TokenHash       string
	InvitedByUserID string
	ExpiresAt       time.Time
	AcceptedAt      *time.Time
	RevokedAt       *time.Time
	CreatedAt       time.Time
}

type InvitationWithTenant struct {
	Invitation Invitation
	Tenant     Tenant
}

type Session struct {
	ID             string
	UserID         string
	TokenHash      string
	ActiveTenantID string
	ExpiresAt      time.Time
	LastSeenAt     time.Time
	CreatedAt      time.Time
}
