package iam

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	domain "github.com/kvitrvn/go-ygg/internal/domain/iam"
	"golang.org/x/crypto/bcrypt"
)

type Store interface {
	WithinTx(ctx context.Context, fn func(store Store) error) error
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	CreateTenant(ctx context.Context, tenant *domain.Tenant) error
	CreateMembership(ctx context.Context, membership *domain.Membership) error
	GetMembershipByUserAndTenant(ctx context.Context, userID, tenantID string) (*domain.Membership, error)
	ListMembershipsByUserID(ctx context.Context, userID string) ([]domain.MembershipWithTenant, error)
	ListTenantMembers(ctx context.Context, tenantID string) ([]domain.TenantMember, error)
	RevokePendingInvitationsByTenantEmail(ctx context.Context, tenantID, email string, revokedAt time.Time) error
	CreateInvitation(ctx context.Context, invitation *domain.Invitation) error
	GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*domain.InvitationWithTenant, error)
	GetInvitationByTokenHashForUpdate(ctx context.Context, tokenHash string) (*domain.InvitationWithTenant, error)
	MarkInvitationAccepted(ctx context.Context, invitationID string, acceptedAt time.Time) error
	CreateSession(ctx context.Context, session *domain.Session) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error
	UpdateSessionActiveTenant(ctx context.Context, sessionID, tenantID string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenManager interface {
	NewToken() (raw string, hashed string, err error)
}

type SessionResult struct {
	SessionToken string
	Auth         *AuthContext
}

type AuthContext struct {
	User             *domain.User
	Session          *domain.Session
	Memberships      []domain.MembershipWithTenant
	ActiveMembership *domain.MembershipWithTenant
}

type SignUpInput struct {
	Username string
	Email    string
	Password string
}

type SignInInput struct {
	Login    string
	Password string
}

type SwitchTenantInput struct {
	SessionToken string
	TenantID     string
}

type CreateOrganizationInput struct {
	SessionToken string
	Name         string
}

type CreateInvitationInput struct {
	SessionToken string
	Email        string
	Role         domain.Role
}

type CreateInvitationResult struct {
	Invitation *domain.InvitationWithTenant
	InviteURL  string
}

type AcceptInvitationInput struct {
	SessionToken string
	Token        string
	Username     string
	Email        string
	Password     string
}

type Service struct {
	store         Store
	hasher        PasswordHasher
	tokenManager  TokenManager
	now           func() time.Time
	sessionTTL    time.Duration
	invitationTTL time.Duration
	appBaseURL    string
}

func NewService(store Store, hasher PasswordHasher, tokenManager TokenManager, sessionTTL, invitationTTL time.Duration, appBaseURL string) *Service {
	return &Service{
		store:         store,
		hasher:        hasher,
		tokenManager:  tokenManager,
		now:           time.Now,
		sessionTTL:    sessionTTL,
		invitationTTL: invitationTTL,
		appBaseURL:    strings.TrimRight(appBaseURL, "/"),
	}
}

func (s *Service) SignUp(ctx context.Context, input SignUpInput) (*SessionResult, error) {
	username := normalizeUsername(input.Username)
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)
	if validation := validateSignUpInput(username, email, password); validation != nil {
		return nil, validation
	}

	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := s.now()
	user := &domain.User{
		ID:           newID("usr"),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	tenant := &domain.Tenant{
		ID:         newID("ten"),
		Slug:       username,
		Name:       username,
		IsPersonal: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	membership := &domain.Membership{
		ID:        newID("mem"),
		UserID:    user.ID,
		TenantID:  tenant.ID,
		Role:      domain.RoleOwner,
		CreatedAt: now,
		UpdatedAt: now,
	}

	rawSessionToken, sessionTokenHash, err := s.tokenManager.NewToken()
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	session := &domain.Session{
		ID:             newID("ses"),
		UserID:         user.ID,
		TokenHash:      sessionTokenHash,
		ActiveTenantID: tenant.ID,
		ExpiresAt:      now.Add(s.sessionTTL),
		LastSeenAt:     now,
		CreatedAt:      now,
	}

	if err := s.store.WithinTx(ctx, func(store Store) error {
		validation := domain.NewValidationErrors()

		existing, err := store.GetUserByEmail(ctx, email)
		if err != nil {
			return err
		}
		if existing != nil {
			validation.Add("email", "This email is already used.")
		}

		existing, err = store.GetUserByUsername(ctx, username)
		if err != nil {
			return err
		}
		if existing != nil {
			validation.Add("username", "This username is already taken.")
		}
		if validation.Any() {
			return validation
		}

		if err := store.CreateUser(ctx, user); err != nil {
			return err
		}
		if err := store.CreateTenant(ctx, tenant); err != nil {
			return err
		}
		if err := store.CreateMembership(ctx, membership); err != nil {
			return err
		}
		if err := store.CreateSession(ctx, session); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("sign up: %w", err)
	}

	auth, err := s.ResolveSession(ctx, rawSessionToken)
	if err != nil {
		return nil, fmt.Errorf("resolve session after sign up: %w", err)
	}
	return &SessionResult{SessionToken: rawSessionToken, Auth: auth}, nil
}

func (s *Service) SignIn(ctx context.Context, input SignInInput) (*SessionResult, error) {
	login := strings.TrimSpace(input.Login)
	password := strings.TrimSpace(input.Password)
	if validation := validateSignInInput(login, password); validation != nil {
		return nil, validation
	}

	user, err := s.findUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("find user by login: %w", err)
	}
	if user == nil || s.hasher.Compare(user.PasswordHash, password) != nil {
		return nil, domain.ErrInvalidCredentials
	}

	memberships, err := s.store.ListMembershipsByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}
	if len(memberships) == 0 {
		return nil, domain.ErrForbidden
	}

	rawSessionToken, sessionTokenHash, err := s.tokenManager.NewToken()
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	now := s.now()
	session := &domain.Session{
		ID:             newID("ses"),
		UserID:         user.ID,
		TokenHash:      sessionTokenHash,
		ActiveTenantID: memberships[0].Tenant.ID,
		ExpiresAt:      now.Add(s.sessionTTL),
		LastSeenAt:     now,
		CreatedAt:      now,
	}

	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	auth, err := s.ResolveSession(ctx, rawSessionToken)
	if err != nil {
		return nil, fmt.Errorf("resolve session after sign in: %w", err)
	}
	return &SessionResult{SessionToken: rawSessionToken, Auth: auth}, nil
}

func (s *Service) CreateOrganization(ctx context.Context, input CreateOrganizationInput) (*AuthContext, error) {
	auth, err := s.ResolveSession(ctx, input.SessionToken)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	if validation := validateOrganizationInput(name); validation != nil {
		return nil, validation
	}

	now := s.now()
	tenant := &domain.Tenant{
		ID:         newID("ten"),
		Slug:       generateTenantSlug(name),
		Name:       name,
		IsPersonal: false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	membership := &domain.Membership{
		ID:        newID("mem"),
		UserID:    auth.User.ID,
		TenantID:  tenant.ID,
		Role:      domain.RoleOwner,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.WithinTx(ctx, func(store Store) error {
		if err := store.CreateTenant(ctx, tenant); err != nil {
			return err
		}
		if err := store.CreateMembership(ctx, membership); err != nil {
			return err
		}
		if err := store.UpdateSessionActiveTenant(ctx, auth.Session.ID, tenant.ID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}

	return s.ResolveSession(ctx, input.SessionToken)
}

func (s *Service) SignOut(ctx context.Context, sessionToken string) error {
	if strings.TrimSpace(sessionToken) == "" {
		return nil
	}
	if err := s.store.DeleteSessionByTokenHash(ctx, hashToken(sessionToken)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *Service) ResolveSession(ctx context.Context, sessionToken string) (*AuthContext, error) {
	if strings.TrimSpace(sessionToken) == "" {
		return nil, domain.ErrUnauthenticated
	}

	session, err := s.store.GetSessionByTokenHash(ctx, hashToken(sessionToken))
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if session == nil {
		return nil, domain.ErrUnauthenticated
	}
	if !session.ExpiresAt.After(s.now()) {
		_ = s.store.DeleteSessionByTokenHash(ctx, session.TokenHash)
		return nil, domain.ErrUnauthenticated
	}

	user, err := s.store.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUnauthenticated
	}

	memberships, err := s.store.ListMembershipsByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}
	if len(memberships) == 0 {
		return nil, domain.ErrForbidden
	}

	activeMembership := findMembership(memberships, session.ActiveTenantID)
	if activeMembership == nil {
		session.ActiveTenantID = memberships[0].Tenant.ID
		if err := s.store.UpdateSessionActiveTenant(ctx, session.ID, session.ActiveTenantID); err != nil {
			return nil, fmt.Errorf("repair session active tenant: %w", err)
		}
		activeMembership = &memberships[0]
	}

	return &AuthContext{
		User:             user,
		Session:          session,
		Memberships:      memberships,
		ActiveMembership: activeMembership,
	}, nil
}

func (s *Service) SwitchActiveTenant(ctx context.Context, input SwitchTenantInput) (*AuthContext, error) {
	auth, err := s.ResolveSession(ctx, input.SessionToken)
	if err != nil {
		return nil, err
	}
	if findMembership(auth.Memberships, input.TenantID) == nil {
		return nil, domain.ErrTenantNotAccessible
	}
	if err := s.store.UpdateSessionActiveTenant(ctx, auth.Session.ID, input.TenantID); err != nil {
		return nil, fmt.Errorf("update session active tenant: %w", err)
	}
	return s.ResolveSession(ctx, input.SessionToken)
}

func (s *Service) ListTenantMembers(ctx context.Context, sessionToken string) (*AuthContext, []domain.TenantMember, error) {
	auth, err := s.ResolveSession(ctx, sessionToken)
	if err != nil {
		return nil, nil, err
	}
	if auth.ActiveMembership.Tenant.IsPersonal {
		return nil, nil, domain.ErrForbidden
	}
	members, err := s.store.ListTenantMembers(ctx, auth.ActiveMembership.Tenant.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("list tenant members: %w", err)
	}
	return auth, members, nil
}

func (s *Service) CreateInvitation(ctx context.Context, input CreateInvitationInput) (*CreateInvitationResult, error) {
	auth, err := s.ResolveSession(ctx, input.SessionToken)
	if err != nil {
		return nil, err
	}
	email := normalizeEmail(input.Email)
	if validation := validateInvitationInput(email, input.Role); validation != nil {
		return nil, validation
	}
	if auth.ActiveMembership.Tenant.IsPersonal {
		return nil, domain.ErrForbidden
	}
	if !auth.ActiveMembership.Membership.Role.CanInvite(input.Role) {
		return nil, domain.ErrForbidden
	}

	rawInvitationToken, invitationTokenHash, err := s.tokenManager.NewToken()
	if err != nil {
		return nil, fmt.Errorf("generate invitation token: %w", err)
	}

	now := s.now()
	invitation := &domain.Invitation{
		ID:              newID("inv"),
		TenantID:        auth.ActiveMembership.Tenant.ID,
		Email:           email,
		Role:            input.Role,
		TokenHash:       invitationTokenHash,
		InvitedByUserID: auth.User.ID,
		ExpiresAt:       now.Add(s.invitationTTL),
		CreatedAt:       now,
	}

	if err := s.store.WithinTx(ctx, func(store Store) error {
		if err := store.RevokePendingInvitationsByTenantEmail(ctx, invitation.TenantID, invitation.Email, now); err != nil {
			return err
		}
		if err := store.CreateInvitation(ctx, invitation); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	return &CreateInvitationResult{
		Invitation: &domain.InvitationWithTenant{
			Invitation: *invitation,
			Tenant:     auth.ActiveMembership.Tenant,
		},
		InviteURL: s.invitationURL(rawInvitationToken),
	}, nil
}

func (s *Service) GetInvitation(ctx context.Context, rawToken string) (*domain.InvitationWithTenant, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, domain.ErrInvitationNotFound
	}
	invitation, err := s.store.GetInvitationByTokenHash(ctx, hashToken(rawToken))
	if err != nil {
		return nil, fmt.Errorf("get invitation: %w", err)
	}
	if invitation == nil {
		return nil, domain.ErrInvitationNotFound
	}
	if err := validateInvitation(invitation.Invitation, s.now()); err != nil {
		return nil, err
	}
	return invitation, nil
}

func (s *Service) AcceptInvitation(ctx context.Context, input AcceptInvitationInput) (*SessionResult, error) {
	if strings.TrimSpace(input.Token) == "" {
		return nil, domain.ErrInvitationNotFound
	}

	var currentAuth *AuthContext
	var err error
	if strings.TrimSpace(input.SessionToken) != "" {
		currentAuth, err = s.ResolveSession(ctx, input.SessionToken)
		if err != nil && err != domain.ErrUnauthenticated {
			return nil, err
		}
		if err == domain.ErrUnauthenticated {
			currentAuth = nil
		}
	}

	rawSessionToken := input.SessionToken
	invitationTokenHash := hashToken(input.Token)
	now := s.now()

	if err := s.store.WithinTx(ctx, func(store Store) error {
		invitation, err := store.GetInvitationByTokenHashForUpdate(ctx, invitationTokenHash)
		if err != nil {
			return err
		}
		if invitation == nil {
			return domain.ErrInvitationNotFound
		}
		if err := validateInvitation(invitation.Invitation, now); err != nil {
			return err
		}

		var user *domain.User
		if currentAuth != nil {
			if currentAuth.User.Email != invitation.Invitation.Email {
				return domain.ErrInvitationEmailMismatch
			}
			user = currentAuth.User
		} else {
			email := normalizeEmail(input.Email)
			password := strings.TrimSpace(input.Password)
			existingUser, err := store.GetUserByEmail(ctx, email)
			if err != nil {
				return err
			}
			if validation := validateInvitationAcceptanceInput(input.Username, email, password, invitation.Invitation.Email, existingUser == nil); validation != nil {
				return validation
			}
			if existingUser != nil {
				if s.hasher.Compare(existingUser.PasswordHash, input.Password) != nil {
					validation := domain.NewValidationErrors()
					validation.Add("password", "Incorrect password for this account.")
					return validation
				}
				user = existingUser
			} else {
				username := normalizeUsername(input.Username)
				existingUser, err := store.GetUserByUsername(ctx, username)
				if err != nil {
					return err
				}
				if existingUser != nil {
					validation := domain.NewValidationErrors()
					validation.Add("username", "This username is already taken.")
					return validation
				}
				passwordHash, err := s.hasher.Hash(input.Password)
				if err != nil {
					return fmt.Errorf("hash password: %w", err)
				}
				personalTenant := &domain.Tenant{
					ID:         newID("ten"),
					Slug:       username,
					Name:       username,
					IsPersonal: true,
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				user = &domain.User{
					ID:           newID("usr"),
					Username:     username,
					Email:        email,
					PasswordHash: passwordHash,
					CreatedAt:    now,
					UpdatedAt:    now,
				}
				if err := store.CreateUser(ctx, user); err != nil {
					return err
				}
				if err := store.CreateTenant(ctx, personalTenant); err != nil {
					return err
				}
				if err := store.CreateMembership(ctx, &domain.Membership{
					ID:        newID("mem"),
					UserID:    user.ID,
					TenantID:  personalTenant.ID,
					Role:      domain.RoleOwner,
					CreatedAt: now,
					UpdatedAt: now,
				}); err != nil {
					return err
				}
			}
		}

		membership, err := store.GetMembershipByUserAndTenant(ctx, user.ID, invitation.Tenant.ID)
		if err != nil {
			return err
		}
		if membership == nil {
			membership = &domain.Membership{
				ID:        newID("mem"),
				UserID:    user.ID,
				TenantID:  invitation.Tenant.ID,
				Role:      invitation.Invitation.Role,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := store.CreateMembership(ctx, membership); err != nil {
				return err
			}
		}

		if currentAuth != nil {
			if err := store.UpdateSessionActiveTenant(ctx, currentAuth.Session.ID, invitation.Tenant.ID); err != nil {
				return err
			}
		} else {
			rawToken, hashedToken, err := s.tokenManager.NewToken()
			if err != nil {
				return fmt.Errorf("generate session token: %w", err)
			}
			rawSessionToken = rawToken
			if err := store.CreateSession(ctx, &domain.Session{
				ID:             newID("ses"),
				UserID:         user.ID,
				TokenHash:      hashedToken,
				ActiveTenantID: invitation.Tenant.ID,
				ExpiresAt:      now.Add(s.sessionTTL),
				LastSeenAt:     now,
				CreatedAt:      now,
			}); err != nil {
				return err
			}
		}

		if err := store.MarkInvitationAccepted(ctx, invitation.Invitation.ID, now); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("accept invitation: %w", err)
	}

	auth, err := s.ResolveSession(ctx, rawSessionToken)
	if err != nil {
		return nil, fmt.Errorf("resolve session after invitation acceptance: %w", err)
	}
	return &SessionResult{SessionToken: rawSessionToken, Auth: auth}, nil
}

func (s *Service) invitationURL(rawToken string) string {
	return s.appBaseURL + "/invitations/" + url.PathEscape(rawToken)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && strings.EqualFold(addr.Address, email)
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func validUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}
	if username[0] == '-' || username[len(username)-1] == '-' {
		return false
	}
	for _, r := range username {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func validateSignUpInput(username, email, password string) error {
	validation := domain.NewValidationErrors()
	if !validUsername(username) {
		validation.Add("username", "Use 3-32 lowercase letters, digits or `-`.")
	}
	if email == "" {
		validation.Add("email", "Email is required.")
	} else if !validEmail(email) {
		validation.Add("email", "Enter a valid email address.")
	}
	if len(password) < 8 {
		validation.Add("password", "Use at least 8 characters.")
	}
	if validation.Any() {
		return validation
	}
	return nil
}

func validateSignInInput(login, password string) error {
	validation := domain.NewValidationErrors()
	if strings.TrimSpace(login) == "" {
		validation.Add("login", "Email or username is required.")
	}
	if strings.TrimSpace(password) == "" {
		validation.Add("password", "Password is required.")
	}
	if validation.Any() {
		return validation
	}
	return nil
}

func validateOrganizationInput(name string) error {
	validation := domain.NewValidationErrors()
	if strings.TrimSpace(name) == "" {
		validation.Add("name", "Organization name is required.")
	}
	if validation.Any() {
		return validation
	}
	return nil
}

func validateInvitationInput(email string, role domain.Role) error {
	validation := domain.NewValidationErrors()
	if email == "" {
		validation.Add("email", "Email is required.")
	} else if !validEmail(email) {
		validation.Add("email", "Enter a valid email address.")
	}
	if !role.Valid() {
		validation.Add("role", "Select a valid role.")
	}
	if validation.Any() {
		return validation
	}
	return nil
}

func validateInvitationAcceptanceInput(username, email, password, invitationEmail string, requireUsername bool) error {
	validation := domain.NewValidationErrors()
	if email == "" {
		validation.Add("email", "Email is required.")
	} else if !validEmail(email) {
		validation.Add("email", "Enter a valid email address.")
	} else if email != invitationEmail {
		validation.Add("email", "Email must match the invitation.")
	}
	if requireUsername && !validUsername(normalizeUsername(username)) {
		validation.Add("username", "Use 3-32 lowercase letters, digits or `-`.")
	}
	if len(password) < 8 {
		validation.Add("password", "Use at least 8 characters.")
	}
	if validation.Any() {
		return validation
	}
	return nil
}

func generateTenantSlug(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prevDash := false
	for _, r := range normalized {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			prevDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		slug = "tenant"
	}
	return slug + "-" + newID("")[:6]
}

func (s *Service) findUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	email := normalizeEmail(login)
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}
	username := normalizeUsername(login)
	if !validUsername(username) {
		return nil, nil
	}
	return s.store.GetUserByUsername(ctx, username)
}

func newID(prefix string) string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	id := hex.EncodeToString(buf[:])
	if prefix == "" {
		return id
	}
	return prefix + "_" + id
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func findMembership(memberships []domain.MembershipWithTenant, tenantID string) *domain.MembershipWithTenant {
	for i := range memberships {
		if memberships[i].Tenant.ID == tenantID {
			return &memberships[i]
		}
	}
	return nil
}

func validateInvitation(invitation domain.Invitation, now time.Time) error {
	if invitation.AcceptedAt != nil {
		return domain.ErrInvitationAccepted
	}
	if invitation.RevokedAt != nil {
		return domain.ErrInvitationRevoked
	}
	if !invitation.ExpiresAt.After(now) {
		return domain.ErrInvitationExpired
	}
	return nil
}

type BcryptHasher struct{}

func (BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

type SHA256TokenManager struct{}

func (SHA256TokenManager) NewToken() (string, string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", "", err
	}
	raw := hex.EncodeToString(buf[:])
	return raw, hashToken(raw), nil
}
