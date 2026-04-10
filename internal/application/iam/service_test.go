package iam

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	domain "github.com/kvitrvn/go-ygg/internal/domain/iam"
)

type fakeStore struct {
	usersByID         map[string]*domain.User
	userIDByEmail     map[string]string
	userIDByUsername  map[string]string
	tenantsByID       map[string]*domain.Tenant
	tenantIDBySlug    map[string]string
	membershipsByID   map[string]*domain.Membership
	invitationsByHash map[string]*domain.Invitation
	sessionsByHash    map[string]*domain.Session
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		usersByID:         map[string]*domain.User{},
		userIDByEmail:     map[string]string{},
		userIDByUsername:  map[string]string{},
		tenantsByID:       map[string]*domain.Tenant{},
		tenantIDBySlug:    map[string]string{},
		membershipsByID:   map[string]*domain.Membership{},
		invitationsByHash: map[string]*domain.Invitation{},
		sessionsByHash:    map[string]*domain.Session{},
	}
}

func (s *fakeStore) WithinTx(_ context.Context, fn func(Store) error) error {
	return fn(s)
}

func (s *fakeStore) CreateUser(_ context.Context, user *domain.User) error {
	if _, exists := s.userIDByEmail[user.Email]; exists {
		return domain.ErrConflict
	}
	if _, exists := s.userIDByUsername[user.Username]; exists {
		return domain.ErrConflict
	}
	copy := *user
	s.usersByID[user.ID] = &copy
	s.userIDByEmail[user.Email] = user.ID
	s.userIDByUsername[user.Username] = user.ID
	return nil
}

func (s *fakeStore) GetUserByID(_ context.Context, id string) (*domain.User, error) {
	user := s.usersByID[id]
	if user == nil {
		return nil, nil
	}
	copy := *user
	return &copy, nil
}

func (s *fakeStore) GetUserByEmail(_ context.Context, email string) (*domain.User, error) {
	id, ok := s.userIDByEmail[email]
	if !ok {
		return nil, nil
	}
	return s.GetUserByID(context.Background(), id)
}

func (s *fakeStore) GetUserByUsername(_ context.Context, username string) (*domain.User, error) {
	id, ok := s.userIDByUsername[username]
	if !ok {
		return nil, nil
	}
	return s.GetUserByID(context.Background(), id)
}

func (s *fakeStore) CreateTenant(_ context.Context, tenant *domain.Tenant) error {
	if _, exists := s.tenantIDBySlug[tenant.Slug]; exists {
		return domain.ErrConflict
	}
	copy := *tenant
	s.tenantsByID[tenant.ID] = &copy
	s.tenantIDBySlug[tenant.Slug] = tenant.ID
	return nil
}

func (s *fakeStore) CreateMembership(_ context.Context, membership *domain.Membership) error {
	for _, existing := range s.membershipsByID {
		if existing.UserID == membership.UserID && existing.TenantID == membership.TenantID {
			return domain.ErrConflict
		}
	}
	copy := *membership
	s.membershipsByID[membership.ID] = &copy
	return nil
}

func (s *fakeStore) GetMembershipByUserAndTenant(_ context.Context, userID, tenantID string) (*domain.Membership, error) {
	for _, membership := range s.membershipsByID {
		if membership.UserID == userID && membership.TenantID == tenantID {
			copy := *membership
			return &copy, nil
		}
	}
	return nil, nil
}

func (s *fakeStore) ListMembershipsByUserID(_ context.Context, userID string) ([]domain.MembershipWithTenant, error) {
	var memberships []domain.MembershipWithTenant
	for _, membership := range s.membershipsByID {
		if membership.UserID != userID {
			continue
		}
		tenant := s.tenantsByID[membership.TenantID]
		membershipCopy := *membership
		tenantCopy := *tenant
		memberships = append(memberships, domain.MembershipWithTenant{
			Membership: membershipCopy,
			Tenant:     tenantCopy,
		})
	}
	sort.Slice(memberships, func(i, j int) bool {
		return memberships[i].Membership.CreatedAt.Before(memberships[j].Membership.CreatedAt)
	})
	return memberships, nil
}

func (s *fakeStore) ListTenantMembers(_ context.Context, tenantID string) ([]domain.TenantMember, error) {
	var members []domain.TenantMember
	for _, membership := range s.membershipsByID {
		if membership.TenantID != tenantID {
			continue
		}
		user := s.usersByID[membership.UserID]
		members = append(members, domain.TenantMember{
			MembershipID: membership.ID,
			UserID:       user.ID,
			Email:        user.Email,
			Role:         membership.Role,
			CreatedAt:    membership.CreatedAt,
		})
	}
	sort.Slice(members, func(i, j int) bool {
		return members[i].Email < members[j].Email
	})
	return members, nil
}

func (s *fakeStore) RevokePendingInvitationsByTenantEmail(_ context.Context, tenantID, email string, revokedAt time.Time) error {
	for _, invitation := range s.invitationsByHash {
		if invitation.TenantID == tenantID && invitation.Email == email && invitation.AcceptedAt == nil && invitation.RevokedAt == nil {
			ts := revokedAt
			invitation.RevokedAt = &ts
		}
	}
	return nil
}

func (s *fakeStore) CreateInvitation(_ context.Context, invitation *domain.Invitation) error {
	copy := *invitation
	s.invitationsByHash[invitation.TokenHash] = &copy
	return nil
}

func (s *fakeStore) GetInvitationByTokenHash(_ context.Context, tokenHash string) (*domain.InvitationWithTenant, error) {
	return s.getInvitation(tokenHash), nil
}

func (s *fakeStore) GetInvitationByTokenHashForUpdate(_ context.Context, tokenHash string) (*domain.InvitationWithTenant, error) {
	return s.getInvitation(tokenHash), nil
}

func (s *fakeStore) getInvitation(tokenHash string) *domain.InvitationWithTenant {
	invitation := s.invitationsByHash[tokenHash]
	if invitation == nil {
		return nil
	}
	invitationCopy := *invitation
	tenantCopy := *s.tenantsByID[invitation.TenantID]
	return &domain.InvitationWithTenant{
		Invitation: invitationCopy,
		Tenant:     tenantCopy,
	}
}

func (s *fakeStore) MarkInvitationAccepted(_ context.Context, invitationID string, acceptedAt time.Time) error {
	for _, invitation := range s.invitationsByHash {
		if invitation.ID == invitationID {
			ts := acceptedAt
			invitation.AcceptedAt = &ts
			return nil
		}
	}
	return nil
}

func (s *fakeStore) CreateSession(_ context.Context, session *domain.Session) error {
	copy := *session
	s.sessionsByHash[session.TokenHash] = &copy
	return nil
}

func (s *fakeStore) GetSessionByTokenHash(_ context.Context, tokenHash string) (*domain.Session, error) {
	session := s.sessionsByHash[tokenHash]
	if session == nil {
		return nil, nil
	}
	copy := *session
	return &copy, nil
}

func (s *fakeStore) DeleteSessionByTokenHash(_ context.Context, tokenHash string) error {
	delete(s.sessionsByHash, tokenHash)
	return nil
}

func (s *fakeStore) UpdateSessionActiveTenant(_ context.Context, sessionID, tenantID string) error {
	for _, session := range s.sessionsByHash {
		if session.ID == sessionID {
			session.ActiveTenantID = tenantID
			return nil
		}
	}
	return nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

func (fakeHasher) Compare(hash, password string) error {
	if hash != "hash:"+password {
		return domain.ErrInvalidCredentials
	}
	return nil
}

type fakeTokenManager struct {
	seq int
}

func (m *fakeTokenManager) NewToken() (string, string, error) {
	m.seq++
	raw := fmt.Sprintf("token-%d", m.seq)
	return raw, hashToken(raw), nil
}

func TestSignUpCreatesPersonalTenantAndSession(t *testing.T) {
	store := newFakeStore()
	service := NewService(store, fakeHasher{}, &fakeTokenManager{}, 24*time.Hour, 48*time.Hour, "http://localhost:8080")

	result, err := service.SignUp(context.Background(), SignUpInput{
		Username: "Owner",
		Email:    "owner@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignUp() error = %v", err)
	}

	if result.Auth.User.Username != "owner" {
		t.Fatalf("result.Auth.User.Username = %q, want owner", result.Auth.User.Username)
	}
	if len(result.Auth.Memberships) != 1 {
		t.Fatalf("len(result.Auth.Memberships) = %d, want 1", len(result.Auth.Memberships))
	}
	if result.Auth.ActiveMembership.Membership.Role != domain.RoleOwner {
		t.Fatalf("result.Auth.ActiveMembership.Membership.Role = %q, want %q", result.Auth.ActiveMembership.Membership.Role, domain.RoleOwner)
	}
	if !result.Auth.ActiveMembership.Tenant.IsPersonal {
		t.Fatal("result.Auth.ActiveMembership.Tenant.IsPersonal = false, want true")
	}
	if result.Auth.ActiveMembership.Tenant.Slug != "owner" {
		t.Fatalf("result.Auth.ActiveMembership.Tenant.Slug = %q, want owner", result.Auth.ActiveMembership.Tenant.Slug)
	}
	if result.SessionToken != "token-1" {
		t.Fatalf("result.SessionToken = %q, want token-1", result.SessionToken)
	}
}

func TestSignUpReturnsFieldValidationErrors(t *testing.T) {
	store := newFakeStore()
	service := NewService(store, fakeHasher{}, &fakeTokenManager{}, 24*time.Hour, 48*time.Hour, "http://localhost:8080")

	_, err := service.SignUp(context.Background(), SignUpInput{
		Username: "A",
		Email:    "not-an-email",
		Password: "short",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("errors.Is(err, ErrInvalidInput) = false, err = %v", err)
	}

	var validation *domain.ValidationErrors
	if !errors.As(err, &validation) {
		t.Fatalf("errors.As(err, *ValidationErrors) = false, err = %v", err)
	}
	if validation.Fields["username"] == "" {
		t.Fatal("username field error is empty")
	}
	if validation.Fields["email"] == "" {
		t.Fatal("email field error is empty")
	}
	if validation.Fields["password"] == "" {
		t.Fatal("password field error is empty")
	}
}

func TestSignUpReturnsBothUniquenessErrors(t *testing.T) {
	store := newFakeStore()
	service := NewService(store, fakeHasher{}, &fakeTokenManager{}, 24*time.Hour, 48*time.Hour, "http://localhost:8080")

	if _, err := service.SignUp(context.Background(), SignUpInput{
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
	}); err != nil {
		t.Fatalf("seed SignUp() error = %v", err)
	}

	_, err := service.SignUp(context.Background(), SignUpInput{
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("errors.Is(err, ErrInvalidInput) = false, err = %v", err)
	}

	var validation *domain.ValidationErrors
	if !errors.As(err, &validation) {
		t.Fatalf("errors.As(err, *ValidationErrors) = false, err = %v", err)
	}
	if validation.Fields["username"] == "" {
		t.Fatal("username uniqueness error is empty")
	}
	if validation.Fields["email"] == "" {
		t.Fatal("email uniqueness error is empty")
	}
}

func TestSignInAcceptsEmailOrUsername(t *testing.T) {
	store := newFakeStore()
	tokens := &fakeTokenManager{}
	service := NewService(store, fakeHasher{}, tokens, 24*time.Hour, 48*time.Hour, "http://localhost:8080")

	if _, err := service.SignUp(context.Background(), SignUpInput{
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
	}); err != nil {
		t.Fatalf("SignUp() error = %v", err)
	}

	byEmail, err := service.SignIn(context.Background(), SignInInput{
		Login:    "owner@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignIn(email) error = %v", err)
	}
	if byEmail.Auth.User.Username != "owner" {
		t.Fatalf("SignIn(email).Auth.User.Username = %q, want owner", byEmail.Auth.User.Username)
	}

	byUsername, err := service.SignIn(context.Background(), SignInInput{
		Login:    "owner",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignIn(username) error = %v", err)
	}
	if byUsername.Auth.User.Email != "owner@example.com" {
		t.Fatalf("SignIn(username).Auth.User.Email = %q, want owner@example.com", byUsername.Auth.User.Email)
	}
}

func TestCreateOrganizationCreatesCollaborativeTenantAndSwitchesSession(t *testing.T) {
	store := newFakeStore()
	service := NewService(store, fakeHasher{}, &fakeTokenManager{}, 24*time.Hour, 48*time.Hour, "http://localhost:8080")

	signUpResult, err := service.SignUp(context.Background(), SignUpInput{
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignUp() error = %v", err)
	}

	auth, err := service.CreateOrganization(context.Background(), CreateOrganizationInput{
		SessionToken: signUpResult.SessionToken,
		Name:         "Acme",
	})
	if err != nil {
		t.Fatalf("CreateOrganization() error = %v", err)
	}

	if len(auth.Memberships) != 2 {
		t.Fatalf("len(auth.Memberships) = %d, want 2", len(auth.Memberships))
	}
	if auth.ActiveMembership.Tenant.Name != "Acme" {
		t.Fatalf("auth.ActiveMembership.Tenant.Name = %q, want Acme", auth.ActiveMembership.Tenant.Name)
	}
	if auth.ActiveMembership.Tenant.IsPersonal {
		t.Fatal("auth.ActiveMembership.Tenant.IsPersonal = true, want false")
	}
}

func TestCreateInvitationReturnsFieldValidationErrors(t *testing.T) {
	store := newFakeStore()
	tokens := &fakeTokenManager{}
	service := NewService(store, fakeHasher{}, tokens, 24*time.Hour, 48*time.Hour, "http://localhost:8080")

	signUpResult, err := service.SignUp(context.Background(), SignUpInput{
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignUp() error = %v", err)
	}
	if _, err := service.CreateOrganization(context.Background(), CreateOrganizationInput{
		SessionToken: signUpResult.SessionToken,
		Name:         "Alpha",
	}); err != nil {
		t.Fatalf("CreateOrganization() error = %v", err)
	}

	_, err = service.CreateInvitation(context.Background(), CreateInvitationInput{
		SessionToken: signUpResult.SessionToken,
		Email:        "bad-email",
		Role:         domain.Role("bogus"),
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("errors.Is(err, ErrInvalidInput) = false, err = %v", err)
	}

	var validation *domain.ValidationErrors
	if !errors.As(err, &validation) {
		t.Fatalf("errors.As(err, *ValidationErrors) = false, err = %v", err)
	}
	if validation.Fields["email"] == "" {
		t.Fatal("email field error is empty")
	}
	if validation.Fields["role"] == "" {
		t.Fatal("role field error is empty")
	}
}

func TestCreateInvitationRejectsPersonalTenant(t *testing.T) {
	store := newFakeStore()
	tokens := &fakeTokenManager{}
	service := NewService(store, fakeHasher{}, tokens, 24*time.Hour, 48*time.Hour, "http://localhost:8080")

	signUpResult, err := service.SignUp(context.Background(), SignUpInput{
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignUp() error = %v", err)
	}

	_, err = service.CreateInvitation(context.Background(), CreateInvitationInput{
		SessionToken: signUpResult.SessionToken,
		Email:        "member@example.com",
		Role:         domain.RoleAdmin,
	})
	if err != domain.ErrForbidden && !strings.Contains(fmt.Sprint(err), domain.ErrForbidden.Error()) {
		t.Fatalf("CreateInvitation() error = %v, want forbidden", err)
	}
}

func TestAcceptInvitationWithExistingSessionAddsMembershipAndSwitchesTenant(t *testing.T) {
	store := newFakeStore()
	tokens := &fakeTokenManager{}
	service := NewService(store, fakeHasher{}, tokens, 24*time.Hour, 48*time.Hour, "http://localhost:8080")

	owner, err := service.SignUp(context.Background(), SignUpInput{
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignUp(owner) error = %v", err)
	}
	if _, err := service.CreateOrganization(context.Background(), CreateOrganizationInput{
		SessionToken: owner.SessionToken,
		Name:         "Alpha",
	}); err != nil {
		t.Fatalf("CreateOrganization() error = %v", err)
	}

	invite, err := service.CreateInvitation(context.Background(), CreateInvitationInput{
		SessionToken: owner.SessionToken,
		Email:        "bob@example.com",
		Role:         domain.RoleMember,
	})
	if err != nil {
		t.Fatalf("CreateInvitation() error = %v", err)
	}

	bob, err := service.SignUp(context.Background(), SignUpInput{
		Username: "bob",
		Email:    "bob@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignUp(bob) error = %v", err)
	}

	accepted, err := service.AcceptInvitation(context.Background(), AcceptInvitationInput{
		SessionToken: bob.SessionToken,
		Token:        "token-2",
	})
	if err != nil {
		t.Fatalf("AcceptInvitation() error = %v", err)
	}

	if accepted.SessionToken != bob.SessionToken {
		t.Fatalf("accepted.SessionToken = %q, want %q", accepted.SessionToken, bob.SessionToken)
	}
	if accepted.Auth.ActiveMembership.Tenant.ID != invite.Invitation.Tenant.ID {
		t.Fatalf("accepted.Auth.ActiveMembership.Tenant.ID = %q, want %q", accepted.Auth.ActiveMembership.Tenant.ID, invite.Invitation.Tenant.ID)
	}
	if len(accepted.Auth.Memberships) != 2 {
		t.Fatalf("len(accepted.Auth.Memberships) = %d, want 2", len(accepted.Auth.Memberships))
	}
}

func TestAcceptInvitationCreatesPersonalTenantForNewUser(t *testing.T) {
	store := newFakeStore()
	tokens := &fakeTokenManager{}
	service := NewService(store, fakeHasher{}, tokens, 24*time.Hour, 48*time.Hour, "http://localhost:8080")

	owner, err := service.SignUp(context.Background(), SignUpInput{
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("SignUp(owner) error = %v", err)
	}
	if _, err := service.CreateOrganization(context.Background(), CreateOrganizationInput{
		SessionToken: owner.SessionToken,
		Name:         "Alpha",
	}); err != nil {
		t.Fatalf("CreateOrganization() error = %v", err)
	}

	invite, err := service.CreateInvitation(context.Background(), CreateInvitationInput{
		SessionToken: owner.SessionToken,
		Email:        "bob@example.com",
		Role:         domain.RoleMember,
	})
	if err != nil {
		t.Fatalf("CreateInvitation() error = %v", err)
	}
	if !strings.HasSuffix(invite.InviteURL, "/invitations/token-2") {
		t.Fatalf("invite.InviteURL = %q, want suffix /invitations/token-2", invite.InviteURL)
	}

	accepted, err := service.AcceptInvitation(context.Background(), AcceptInvitationInput{
		Token:    "token-2",
		Username: "bob",
		Email:    "bob@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("AcceptInvitation() error = %v", err)
	}

	if accepted.Auth.User.Username != "bob" {
		t.Fatalf("accepted.Auth.User.Username = %q, want bob", accepted.Auth.User.Username)
	}
	if len(accepted.Auth.Memberships) != 2 {
		t.Fatalf("len(accepted.Auth.Memberships) = %d, want 2", len(accepted.Auth.Memberships))
	}
	if accepted.Auth.ActiveMembership.Tenant.ID != invite.Invitation.Tenant.ID {
		t.Fatalf("accepted.Auth.ActiveMembership.Tenant.ID = %q, want %q", accepted.Auth.ActiveMembership.Tenant.ID, invite.Invitation.Tenant.ID)
	}

	personalCount := 0
	for _, membership := range accepted.Auth.Memberships {
		if membership.Tenant.IsPersonal {
			personalCount++
		}
	}
	if personalCount != 1 {
		t.Fatalf("personal membership count = %d, want 1", personalCount)
	}
}
