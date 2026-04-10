package cli

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"

	appconfig "github.com/kvitrvn/go-ygg/internal/infrastructure/config"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
}

type migrateClient interface {
	Up() error
	Steps(int) error
	Version() (uint, bool, error)
	Close() (error, error)
}

type migrationStatus struct {
	Applied    bool
	HasVersion bool
	Version    uint
	Dirty      bool
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(_ *cobra.Command, _ []string) error {
		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer closeMigrate(m)

		status, err := applyMigrations(m)
		if err != nil {
			return fmt.Errorf("migrate up: %w", err)
		}

		if status.Applied {
			fmt.Println("migrations applied")
		} else {
			fmt.Println("no migrations to apply")
		}
		printMigrationVersion(status)
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down [N]",
	Short: "Revert N migrations (default: 1)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		steps := 1
		if len(args) > 0 {
			if _, err := fmt.Sscanf(args[0], "%d", &steps); err != nil {
				return fmt.Errorf("invalid steps: %w", err)
			}
		}
		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer closeMigrate(m)

		if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate down: %w", err)
		}
		fmt.Printf("%d migration(s) reverted\n", steps)
		status, err := migrationVersion(m)
		if err != nil {
			return fmt.Errorf("get version after migrate down: %w", err)
		}
		printMigrationVersion(status)
		return nil
	},
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print current migration version",
	RunE: func(_ *cobra.Command, _ []string) error {
		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer closeMigrate(m)

		status, err := migrationVersion(m)
		if err != nil {
			return fmt.Errorf("get version: %w", err)
		}
		printMigrationVersion(status)
		return nil
	},
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateVersionCmd)
}

func newMigrate() (*migrate.Migrate, error) {
	cfg, err := appconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	m, err := migrate.New("file://migrations", cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("create migrate instance: %w", err)
	}
	return m, nil
}

func applyMigrations(m migrateClient) (migrationStatus, error) {
	status := migrationStatus{Applied: true}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			status.Applied = false
		} else {
			return migrationStatus{}, err
		}
	}

	versionStatus, err := migrationVersion(m)
	if err != nil {
		return migrationStatus{}, err
	}

	status.HasVersion = versionStatus.HasVersion
	status.Version = versionStatus.Version
	status.Dirty = versionStatus.Dirty

	return status, nil
}

func migrationVersion(m interface{ Version() (uint, bool, error) }) (migrationStatus, error) {
	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return migrationStatus{}, nil
		}
		return migrationStatus{}, err
	}

	return migrationStatus{
		HasVersion: true,
		Version:    version,
		Dirty:      dirty,
	}, nil
}

func logMigrationStatus(status migrationStatus) {
	if status.Applied {
		slog.Info("database migrations applied", "version", migrationVersionValue(status), "dirty", status.Dirty)
		return
	}

	slog.Info("database migrations already up to date", "version", migrationVersionValue(status), "dirty", status.Dirty)
}

func printMigrationVersion(status migrationStatus) {
	if !status.HasVersion {
		fmt.Println("version: none, dirty: false")
		return
	}

	fmt.Printf("version: %d, dirty: %v\n", status.Version, status.Dirty)
}

func migrationVersionValue(status migrationStatus) any {
	if !status.HasVersion {
		return "none"
	}

	return status.Version
}

func closeMigrate(m migrateClient) {
	sourceErr, databaseErr := m.Close()
	if sourceErr != nil {
		slog.Warn("closing migration source failed", "error", sourceErr)
	}
	if databaseErr != nil {
		slog.Warn("closing migration database failed", "error", databaseErr)
	}
}
