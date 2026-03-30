package cli

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	appconfig "github.com/kvitrvn/go-ygg/internal/infrastructure/config"
	apphttp "github.com/kvitrvn/go-ygg/internal/interfaces/http"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := appconfig.Load(v)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		slog.Info("starting server", "addr", addr)

		srv := apphttp.NewServer(addr)
		return srv.Start()
	},
}
