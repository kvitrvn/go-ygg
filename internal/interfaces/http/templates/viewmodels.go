package templates

import (
	appiam "github.com/kvitrvn/go-ygg/internal/application/iam"
	domain "github.com/kvitrvn/go-ygg/internal/domain/iam"
)

type AuthFormPageData struct {
	Username string
	Login    string
	Email    string
	Error    string
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
	Auth  *appiam.AuthContext
	Name  string
	Error string
}

type InvitationCreatePageData struct {
	Auth      *appiam.AuthContext
	Email     string
	Role      domain.Role
	InviteURL string
	Error     string
}

type InvitationAcceptPageData struct {
	Auth                 *appiam.AuthContext
	Invitation           *domain.InvitationWithTenant
	RawToken             string
	Username             string
	Email                string
	Error                string
	CanUseCurrentSession bool
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
