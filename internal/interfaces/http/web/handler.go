package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	appiam "github.com/kvitrvn/go-ygg/internal/application/iam"
	domain "github.com/kvitrvn/go-ygg/internal/domain/iam"
	"github.com/kvitrvn/go-ygg/internal/interfaces/http/templates"
)

type Handler struct {
	iam        *appiam.Service
	cookie     CookieConfig
	sessionTTL time.Duration
}

func NewHandler(iam *appiam.Service, cookie CookieConfig, sessionTTL time.Duration) *Handler {
	return &Handler{
		iam:        iam,
		cookie:     cookie,
		sessionTTL: sessionTTL,
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if AuthFromContext(r.Context()) != nil {
		Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	render(w, r, http.StatusOK, templates.HomePage(IsHTMXRequest(r)))
}

func (h *Handler) ShowSignUp(w http.ResponseWriter, r *http.Request) {
	if AuthFromContext(r.Context()) != nil {
		Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	render(w, r, http.StatusOK, templates.SignUpPage(templates.AuthFormPageData{}, IsHTMXRequest(r)))
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	result, err := h.iam.SignUp(r.Context(), appiam.SignUpInput{
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	})
	if err != nil {
		fieldErrors := validationErrors(err)
		render(w, r, statusForError(err), templates.SignUpPage(templates.AuthFormPageData{
			Username:    r.FormValue("username"),
			Email:       r.FormValue("email"),
			Error:       authErrorMessage(err, fieldErrors),
			FieldErrors: fieldErrors,
		}, IsHTMXRequest(r)))
		return
	}

	SetSessionCookie(w, h.cookie, result.SessionToken, timeDuration(h.sessionTTL))
	Redirect(w, r, "/app", http.StatusSeeOther)
}

func (h *Handler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	if AuthFromContext(r.Context()) != nil {
		Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	render(w, r, http.StatusOK, templates.LoginPage(templates.AuthFormPageData{}, IsHTMXRequest(r)))
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	result, err := h.iam.SignIn(r.Context(), appiam.SignInInput{
		Login:    r.FormValue("login"),
		Password: r.FormValue("password"),
	})
	if err != nil {
		fieldErrors := validationErrors(err)
		render(w, r, statusForError(err), templates.LoginPage(templates.AuthFormPageData{
			Login:       r.FormValue("login"),
			Error:       authErrorMessage(err, fieldErrors),
			FieldErrors: fieldErrors,
		}, IsHTMXRequest(r)))
		return
	}

	SetSessionCookie(w, h.cookie, result.SessionToken, timeDuration(h.sessionTTL))
	Redirect(w, r, "/app", http.StatusSeeOther)
}

func (h *Handler) SignOut(w http.ResponseWriter, r *http.Request) {
	_ = h.iam.SignOut(r.Context(), ReadSessionToken(r, h.cookie))
	ClearSessionCookie(w, h.cookie)
	Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	auth := AuthFromContext(r.Context())
	render(w, r, http.StatusOK, templates.DashboardPage(templates.DashboardPageData{Auth: auth}, IsHTMXRequest(r)))
}

func (h *Handler) SwitchTenant(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	_, err := h.iam.SwitchActiveTenant(r.Context(), appiam.SwitchTenantInput{
		SessionToken: ReadSessionToken(r, h.cookie),
		TenantID:     r.FormValue("tenant_id"),
	})
	if err != nil {
		auth := AuthFromContext(r.Context())
		render(w, r, statusForError(err), templates.DashboardPage(templates.DashboardPageData{
			Auth:  auth,
			Error: authErrorMessage(err, nil),
		}, IsHTMXRequest(r)))
		return
	}
	Redirect(w, r, "/app", http.StatusSeeOther)
}

func (h *Handler) Members(w http.ResponseWriter, r *http.Request) {
	auth, members, err := h.iam.ListTenantMembers(r.Context(), ReadSessionToken(r, h.cookie))
	if err != nil {
		http.Error(w, authErrorMessage(err, nil), statusForError(err))
		return
	}
	render(w, r, http.StatusOK, templates.MembersPage(templates.MembersPageData{
		Auth:    auth,
		Members: members,
	}, IsHTMXRequest(r)))
}

func (h *Handler) ShowOrganizationCreate(w http.ResponseWriter, r *http.Request) {
	render(w, r, http.StatusOK, templates.OrganizationCreatePage(templates.OrganizationCreatePageData{
		Auth: AuthFromContext(r.Context()),
	}, IsHTMXRequest(r)))
}

func (h *Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	_, err := h.iam.CreateOrganization(r.Context(), appiam.CreateOrganizationInput{
		SessionToken: ReadSessionToken(r, h.cookie),
		Name:         r.FormValue("name"),
	})
	if err != nil {
		fieldErrors := validationErrors(err)
		render(w, r, statusForError(err), templates.OrganizationCreatePage(templates.OrganizationCreatePageData{
			Auth:        AuthFromContext(r.Context()),
			Name:        r.FormValue("name"),
			Error:       authErrorMessage(err, fieldErrors),
			FieldErrors: fieldErrors,
		}, IsHTMXRequest(r)))
		return
	}

	Redirect(w, r, "/app", http.StatusSeeOther)
}

func (h *Handler) ShowInvitationCreate(w http.ResponseWriter, r *http.Request) {
	auth := AuthFromContext(r.Context())
	if auth == nil || auth.ActiveMembership.Tenant.IsPersonal {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	render(w, r, http.StatusOK, templates.InviteCreatePage(templates.InvitationCreatePageData{
		Auth: auth,
		Role: domain.RoleMember,
	}, IsHTMXRequest(r)))
}

func (h *Handler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	result, err := h.iam.CreateInvitation(r.Context(), appiam.CreateInvitationInput{
		SessionToken: ReadSessionToken(r, h.cookie),
		Email:        r.FormValue("email"),
		Role:         domain.Role(r.FormValue("role")),
	})
	if err != nil {
		fieldErrors := validationErrors(err)
		render(w, r, statusForError(err), templates.InviteCreatePage(templates.InvitationCreatePageData{
			Auth:        AuthFromContext(r.Context()),
			Email:       r.FormValue("email"),
			Role:        domain.Role(r.FormValue("role")),
			Error:       authErrorMessage(err, fieldErrors),
			FieldErrors: fieldErrors,
		}, IsHTMXRequest(r)))
		return
	}

	render(w, r, http.StatusOK, templates.InviteCreatePage(templates.InvitationCreatePageData{
		Auth:      AuthFromContext(r.Context()),
		InviteURL: result.InviteURL,
		Email:     result.Invitation.Invitation.Email,
		Role:      result.Invitation.Invitation.Role,
	}, IsHTMXRequest(r)))
}

func (h *Handler) ShowInvitation(w http.ResponseWriter, r *http.Request) {
	invitation, err := h.iam.GetInvitation(r.Context(), r.PathValue("token"))
	if err != nil {
		fieldErrors := validationErrors(err)
		render(w, r, statusForError(err), templates.InvitationAcceptPage(templates.InvitationAcceptPageData{
			Auth:        AuthFromContext(r.Context()),
			RawToken:    r.PathValue("token"),
			Error:       authErrorMessage(err, fieldErrors),
			FieldErrors: fieldErrors,
		}, IsHTMXRequest(r)))
		return
	}

	auth := AuthFromContext(r.Context())
	canUseCurrentSession := auth != nil && auth.User.Email == invitation.Invitation.Email
	render(w, r, http.StatusOK, templates.InvitationAcceptPage(templates.InvitationAcceptPageData{
		Auth:                 auth,
		Invitation:           invitation,
		RawToken:             r.PathValue("token"),
		Email:                invitation.Invitation.Email,
		CanUseCurrentSession: canUseCurrentSession,
	}, IsHTMXRequest(r)))
}

func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	token := r.PathValue("token")
	result, err := h.iam.AcceptInvitation(r.Context(), appiam.AcceptInvitationInput{
		SessionToken: ReadSessionToken(r, h.cookie),
		Token:        token,
		Username:     r.FormValue("username"),
		Email:        r.FormValue("email"),
		Password:     r.FormValue("password"),
	})
	if err != nil {
		invitation, invitationErr := h.iam.GetInvitation(r.Context(), token)
		if invitationErr != nil && !errors.Is(invitationErr, domain.ErrInvitationAccepted) {
			invitation = nil
		}
		fieldErrors := validationErrors(err)
		render(w, r, statusForError(err), templates.InvitationAcceptPage(templates.InvitationAcceptPageData{
			Auth:        AuthFromContext(r.Context()),
			Invitation:  invitation,
			RawToken:    token,
			Username:    r.FormValue("username"),
			Email:       r.FormValue("email"),
			Error:       authErrorMessage(err, fieldErrors),
			FieldErrors: fieldErrors,
		}, IsHTMXRequest(r)))
		return
	}

	SetSessionCookie(w, h.cookie, result.SessionToken, timeDuration(h.sessionTTL))
	Redirect(w, r, "/app", http.StatusSeeOther)
}

func render(w http.ResponseWriter, r *http.Request, status int, component interface {
	Render(ctx context.Context, w io.Writer) error
}) {
	AddHTMLVary(w)
	if IsHTMXRequest(r) && status >= http.StatusBadRequest && status < http.StatusInternalServerError {
		status = http.StatusOK
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = component.Render(r.Context(), w)
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrInvitationEmailMismatch):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrInvitationNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrInvitationExpired), errors.Is(err, domain.ErrInvitationRevoked), errors.Is(err, domain.ErrInvitationAccepted):
		return http.StatusGone
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func validationErrors(err error) map[string]string {
	var validation *domain.ValidationErrors
	if errors.As(err, &validation) && validation != nil && validation.Any() {
		return validation.Fields
	}
	if errors.Is(err, domain.ErrInvitationEmailMismatch) {
		return map[string]string{"email": "Email must match the invitation."}
	}
	return nil
}

func authErrorMessage(err error, fieldErrors map[string]string) string {
	if len(fieldErrors) > 0 {
		return ""
	}
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return "Check the form fields and try again."
	case errors.Is(err, domain.ErrInvalidCredentials):
		return "Invalid email, username or password."
	case errors.Is(err, domain.ErrUnauthenticated):
		return "You must sign in first."
	case errors.Is(err, domain.ErrForbidden):
		return "You do not have access to this action."
	case errors.Is(err, domain.ErrConflict):
		return "This action conflicts with existing data."
	case errors.Is(err, domain.ErrInvitationNotFound):
		return "Invitation not found."
	case errors.Is(err, domain.ErrInvitationExpired):
		return "Invitation expired."
	case errors.Is(err, domain.ErrInvitationRevoked):
		return "Invitation revoked."
	case errors.Is(err, domain.ErrInvitationAccepted):
		return "Invitation already accepted."
	case errors.Is(err, domain.ErrInvitationEmailMismatch):
		return "The invitation email does not match the current account."
	case errors.Is(err, domain.ErrTenantNotAccessible):
		return "Tenant not accessible."
	default:
		return fmt.Sprintf("Unexpected error: %v", err)
	}
}

func timeDuration(duration time.Duration) time.Duration {
	return duration
}
