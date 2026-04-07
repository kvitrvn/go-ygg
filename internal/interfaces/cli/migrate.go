package cli

import (
	"fmt"

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

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(_ *cobra.Command, _ []string) error {
		m, err := newMigrate()
		if err != nil {
			return err
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate up: %w", err)
		}
		fmt.Println("migrations applied")
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
		if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate down: %w", err)
		}
		fmt.Printf("%d migration(s) reverted\n", steps)
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
		ver, dirty, err := m.Version()
		if err != nil {
			return fmt.Errorf("get version: %w", err)
		}
		fmt.Printf("version: %d, dirty: %v\n", ver, dirty)
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
