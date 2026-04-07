package cli

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	appconfig "github.com/kvitrvn/go-ygg/internal/infrastructure/config"
)

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "go-ygg — your Go application",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		return initConfig(cmd)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig(_ *cobra.Command) error {
	cfg, err := appconfig.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	setupLogger(cfg.Log.Level, cfg.Log.Format)
	return nil
}

func setupLogger(level, format string) {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: l}

	var h slog.Handler
	if strings.ToLower(format) == "text" {
		h = slog.NewTextHandler(os.Stderr, opts)
	} else {
		h = slog.NewJSONHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(h))
}
