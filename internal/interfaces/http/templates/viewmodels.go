package templates

import (
	appiam "github.com/kvitrvn/go-ygg/internal/application/iam"
	domain "github.com/kvitrvn/go-ygg/internal/domain/iam"
)

type AuthFormPageData struct {
	Username    string
	Login       string
	Email       string
	Error       string
	FieldErrors map[string]string
}

type DashboardPageData struct {
	Auth  *appiam.AuthContext
	Error string
}

type MembersPageData struct {
	Auth    *appiam.AuthContext
	Members []domain.TenantMember
}

type OrganizationCreatePageData struct {
	Auth        *appiam.AuthContext
	Name        string
	Error       string
	FieldErrors map[string]string
}

type InvitationCreatePageData struct {
	Auth        *appiam.AuthContext
	Email       string
	Role        domain.Role
	InviteURL   string
	Error       string
	FieldErrors map[string]string
}

type InvitationAcceptPageData struct {
	Auth                 *appiam.AuthContext
	Invitation           *domain.InvitationWithTenant
	RawToken             string
	Username             string
	Email                string
	Error                string
	FieldErrors          map[string]string
	CanUseCurrentSession bool
}

func fieldError(errors map[string]string, field string) string {
	if errors == nil {
		return ""
	}
	return errors[field]
}

func hasFieldError(errors map[string]string, field string) bool {
	return fieldError(errors, field) != ""
}

func tenantOptionLabel(membership domain.MembershipWithTenant) string {
	if membership.Tenant.IsPersonal {
		return membership.Tenant.Name + " (personal)"
	}
	return membership.Tenant.Name + " (" + string(membership.Membership.Role) + ")"
}

func tenantKindLabel(tenant domain.Tenant) string {
	if tenant.IsPersonal {
		return "Personal tenant"
	}
	return "Organization"
}
