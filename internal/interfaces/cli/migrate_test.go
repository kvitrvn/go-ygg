package cli

import (
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
)

type fakeMigrator struct {
	upErr      error
	version    uint
	dirty      bool
	versionErr error
}

func (f fakeMigrator) Up() error {
	return f.upErr
}

func (f fakeMigrator) Steps(int) error {
	return nil
}

func (f fakeMigrator) Version() (uint, bool, error) {
	return f.version, f.dirty, f.versionErr
}

func (f fakeMigrator) Close() (error, error) {
	return nil, nil
}

func TestApplyMigrationsApplied(t *testing.T) {
	status, err := applyMigrations(fakeMigrator{
		version: 1,
	})
	if err != nil {
		t.Fatalf("applyMigrations() error = %v", err)
	}

	if !status.Applied {
		t.Fatal("status.Applied = false, want true")
	}
	if !status.HasVersion {
		t.Fatal("status.HasVersion = false, want true")
	}
	if status.Version != 1 {
		t.Fatalf("status.Version = %d, want 1", status.Version)
	}
	if status.Dirty {
		t.Fatal("status.Dirty = true, want false")
	}
}

func TestApplyMigrationsNoChange(t *testing.T) {
	status, err := applyMigrations(fakeMigrator{
		upErr:   migrate.ErrNoChange,
		version: 7,
	})
	if err != nil {
		t.Fatalf("applyMigrations() error = %v", err)
	}

	if status.Applied {
		t.Fatal("status.Applied = true, want false")
	}
	if !status.HasVersion {
		t.Fatal("status.HasVersion = false, want true")
	}
	if status.Version != 7 {
		t.Fatalf("status.Version = %d, want 7", status.Version)
	}
}

func TestApplyMigrationsNoVersionYet(t *testing.T) {
	status, err := applyMigrations(fakeMigrator{
		upErr:      migrate.ErrNoChange,
		versionErr: migrate.ErrNilVersion,
	})
	if err != nil {
		t.Fatalf("applyMigrations() error = %v", err)
	}

	if status.Applied {
		t.Fatal("status.Applied = true, want false")
	}
	if status.HasVersion {
		t.Fatal("status.HasVersion = true, want false")
	}
}

func TestApplyMigrationsReturnsUpError(t *testing.T) {
	wantErr := errors.New("boom")

	_, err := applyMigrations(fakeMigrator{
		upErr: wantErr,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("applyMigrations() error = %v, want %v", err, wantErr)
	}
}

func TestMigrationVersionHandlesNilVersion(t *testing.T) {
	status, err := migrationVersion(fakeMigrator{
		versionErr: migrate.ErrNilVersion,
	})
	if err != nil {
		t.Fatalf("migrationVersion() error = %v", err)
	}
	if status.HasVersion {
		t.Fatal("status.HasVersion = true, want false")
	}
}
