package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	appiam "github.com/kvitrvn/go-ygg/internal/application/iam"
	domain "github.com/kvitrvn/go-ygg/internal/domain/iam"
	"github.com/lib/pq"
)

type dbtx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type IAMStore struct {
	db *sql.DB
	q  dbtx
}

func OpenPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

func NewIAMStore(db *sql.DB) *IAMStore {
	return &IAMStore{db: db, q: db}
}

var _ appiam.Store = (*IAMStore)(nil)

func (s *IAMStore) WithinTx(ctx context.Context, fn func(store appiam.Store) error) error {
	return s.withinTx(ctx, func(store *IAMStore) error {
		return fn(store)
	})
}

func (s *IAMStore) withinTx(ctx context.Context, fn func(store *IAMStore) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	txStore := &IAMStore{db: s.db, q: tx}
	if err := fn(txStore); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("%w: rollback tx: %v", err, rollbackErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (s *IAMStore) CreateUser(ctx context.Context, user *domain.User) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	return mapPQError(err)
}

func (s *IAMStore) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	row := s.q.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)
	var user domain.User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *IAMStore) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := s.q.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`, email)
	var user domain.User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *IAMStore) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := s.q.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE LOWER(username) = LOWER($1)
	`, username)
	var user domain.User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *IAMStore) CreateTenant(ctx context.Context, tenant *domain.Tenant) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO tenants (id, slug, name, is_personal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tenant.ID, tenant.Slug, tenant.Name, tenant.IsPersonal, tenant.CreatedAt, tenant.UpdatedAt)
	return mapPQError(err)
}

func (s *IAMStore) CreateMembership(ctx context.Context, membership *domain.Membership) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO memberships (id, user_id, tenant_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, membership.ID, membership.UserID, membership.TenantID, membership.Role, membership.CreatedAt, membership.UpdatedAt)
	return mapPQError(err)
}

func (s *IAMStore) GetMembershipByUserAndTenant(ctx context.Context, userID, tenantID string) (*domain.Membership, error) {
	row := s.q.QueryRowContext(ctx, `
		SELECT id, user_id, tenant_id, role, created_at, updated_at
		FROM memberships
		WHERE user_id = $1 AND tenant_id = $2
	`, userID, tenantID)
	var membership domain.Membership
	if err := row.Scan(&membership.ID, &membership.UserID, &membership.TenantID, &membership.Role, &membership.CreatedAt, &membership.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &membership, nil
}

func (s *IAMStore) ListMembershipsByUserID(ctx context.Context, userID string) ([]domain.MembershipWithTenant, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT
			m.id, m.user_id, m.tenant_id, m.role, m.created_at, m.updated_at,
			t.id, t.slug, t.name, t.is_personal, t.created_at, t.updated_at
		FROM memberships m
		INNER JOIN tenants t ON t.id = m.tenant_id
		WHERE m.user_id = $1
		ORDER BY m.created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []domain.MembershipWithTenant
	for rows.Next() {
		var membership domain.MembershipWithTenant
		if err := rows.Scan(
			&membership.Membership.ID,
			&membership.Membership.UserID,
			&membership.Membership.TenantID,
			&membership.Membership.Role,
			&membership.Membership.CreatedAt,
			&membership.Membership.UpdatedAt,
			&membership.Tenant.ID,
			&membership.Tenant.Slug,
			&membership.Tenant.Name,
			&membership.Tenant.IsPersonal,
			&membership.Tenant.CreatedAt,
			&membership.Tenant.UpdatedAt,
		); err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}
	return memberships, rows.Err()
}

func (s *IAMStore) ListTenantMembers(ctx context.Context, tenantID string) ([]domain.TenantMember, error) {
	rows, err := s.q.QueryContext(ctx, `
		SELECT m.id, u.id, u.email, m.role, m.created_at
		FROM memberships m
		INNER JOIN users u ON u.id = m.user_id
		WHERE m.tenant_id = $1
		ORDER BY u.email ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TenantMember
	for rows.Next() {
		var member domain.TenantMember
		if err := rows.Scan(&member.MembershipID, &member.UserID, &member.Email, &member.Role, &member.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (s *IAMStore) RevokePendingInvitationsByTenantEmail(ctx context.Context, tenantID, email string, revokedAt time.Time) error {
	_, err := s.q.ExecContext(ctx, `
		UPDATE invitations
		SET revoked_at = $3
		WHERE tenant_id = $1
		  AND LOWER(email) = LOWER($2)
		  AND accepted_at IS NULL
		  AND revoked_at IS NULL
	`, tenantID, email, revokedAt)
	return err
}

func (s *IAMStore) CreateInvitation(ctx context.Context, invitation *domain.Invitation) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO invitations (id, tenant_id, email, role, token_hash, invited_by_user_id, expires_at, accepted_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, invitation.ID, invitation.TenantID, invitation.Email, invitation.Role, invitation.TokenHash, invitation.InvitedByUserID, invitation.ExpiresAt, invitation.AcceptedAt, invitation.RevokedAt, invitation.CreatedAt)
	return mapPQError(err)
}

func (s *IAMStore) GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*domain.InvitationWithTenant, error) {
	return s.getInvitation(ctx, tokenHash, false)
}

func (s *IAMStore) GetInvitationByTokenHashForUpdate(ctx context.Context, tokenHash string) (*domain.InvitationWithTenant, error) {
	return s.getInvitation(ctx, tokenHash, true)
}

func (s *IAMStore) getInvitation(ctx context.Context, tokenHash string, forUpdate bool) (*domain.InvitationWithTenant, error) {
	query := `
		SELECT
			i.id, i.tenant_id, i.email, i.role, i.token_hash, i.invited_by_user_id, i.expires_at, i.accepted_at, i.revoked_at, i.created_at,
			t.id, t.slug, t.name, t.is_personal, t.created_at, t.updated_at
		FROM invitations i
		INNER JOIN tenants t ON t.id = i.tenant_id
		WHERE i.token_hash = $1
	`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	row := s.q.QueryRowContext(ctx, query, tokenHash)

	var invitation domain.InvitationWithTenant
	var acceptedAt sql.NullTime
	var revokedAt sql.NullTime
	if err := row.Scan(
		&invitation.Invitation.ID,
		&invitation.Invitation.TenantID,
		&invitation.Invitation.Email,
		&invitation.Invitation.Role,
		&invitation.Invitation.TokenHash,
		&invitation.Invitation.InvitedByUserID,
		&invitation.Invitation.ExpiresAt,
		&acceptedAt,
		&revokedAt,
		&invitation.Invitation.CreatedAt,
		&invitation.Tenant.ID,
		&invitation.Tenant.Slug,
		&invitation.Tenant.Name,
		&invitation.Tenant.IsPersonal,
		&invitation.Tenant.CreatedAt,
		&invitation.Tenant.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if acceptedAt.Valid {
		invitation.Invitation.AcceptedAt = &acceptedAt.Time
	}
	if revokedAt.Valid {
		invitation.Invitation.RevokedAt = &revokedAt.Time
	}
	return &invitation, nil
}

func (s *IAMStore) MarkInvitationAccepted(ctx context.Context, invitationID string, acceptedAt time.Time) error {
	_, err := s.q.ExecContext(ctx, `
		UPDATE invitations
		SET accepted_at = $2
		WHERE id = $1
	`, invitationID, acceptedAt)
	return err
}

func (s *IAMStore) CreateSession(ctx context.Context, session *domain.Session) error {
	_, err := s.q.ExecContext(ctx, `
		INSERT INTO sessions (id, user_id, token_hash, active_tenant_id, expires_at, last_seen_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, session.ID, session.UserID, session.TokenHash, session.ActiveTenantID, session.ExpiresAt, session.LastSeenAt, session.CreatedAt)
	return mapPQError(err)
}

func (s *IAMStore) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	row := s.q.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, active_tenant_id, expires_at, last_seen_at, created_at
		FROM sessions
		WHERE token_hash = $1
	`, tokenHash)
	var session domain.Session
	if err := row.Scan(&session.ID, &session.UserID, &session.TokenHash, &session.ActiveTenantID, &session.ExpiresAt, &session.LastSeenAt, &session.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (s *IAMStore) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := s.q.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (s *IAMStore) UpdateSessionActiveTenant(ctx context.Context, sessionID, tenantID string) error {
	_, err := s.q.ExecContext(ctx, `
		UPDATE sessions
		SET active_tenant_id = $2,
		    last_seen_at = NOW()
		WHERE id = $1
	`, sessionID, tenantID)
	return err
}

func mapPQError(err error) error {
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return domain.ErrConflict
	}
	return err
}
