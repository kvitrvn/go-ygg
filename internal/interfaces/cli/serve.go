package cli

import (
	"fmt"
	"log/slog"
	"time"

	appiam "github.com/kvitrvn/go-ygg/internal/application/iam"
	"github.com/spf13/cobra"

	appconfig "github.com/kvitrvn/go-ygg/internal/infrastructure/config"
	"github.com/kvitrvn/go-ygg/internal/infrastructure/persistence"
	apphttp "github.com/kvitrvn/go-ygg/internal/interfaces/http"
	"github.com/kvitrvn/go-ygg/internal/interfaces/http/web"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := appconfig.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		db, err := persistence.OpenPostgres(cfg.Database.DSN)
		if err != nil {
			return fmt.Errorf("open postgres: %w", err)
		}
		defer func() {
			if closeErr := db.Close(); closeErr != nil {
				slog.Warn("closing postgres failed", "error", closeErr)
			}
		}()
		if err := db.Ping(); err != nil {
			return fmt.Errorf("ping postgres: %w", err)
		}

		m, err := newMigrate()
		if err != nil {
			return err
		}

		slog.Info("applying database migrations")
		migrationStatus, err := applyMigrations(m)
		closeMigrate(m)
		if err != nil {
			return fmt.Errorf("apply database migrations: %w", err)
		}
		logMigrationStatus(migrationStatus)

		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		slog.Info("starting server", "addr", addr)

		sessionTTL, err := time.ParseDuration(cfg.Auth.SessionTTL)
		if err != nil {
			return fmt.Errorf("parse session ttl: %w", err)
		}
		invitationTTL, err := time.ParseDuration(cfg.Auth.InvitationTTL)
		if err != nil {
			return fmt.Errorf("parse invitation ttl: %w", err)
		}

		iamService := appiam.NewService(
			persistence.NewIAMStore(db),
			appiam.BcryptHasher{},
			appiam.SHA256TokenManager{},
			sessionTTL,
			invitationTTL,
			cfg.App.BaseURL,
		)

		srv := apphttp.NewServer(
			addr,
			iamService,
			web.CookieConfig{Name: cfg.Auth.CookieName, Secure: cfg.Auth.CookieSecure},
			cfg.App.BaseURL,
			sessionTTL,
		)
		return srv.Start()
	},
}
