package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	appconfig "github.com/kvitrvn/go-ygg/internal/infrastructure/config"
)

var (
	cfgFile  string
	logLevel string
	v        *viper.Viper
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
	v = appconfig.DefaultViper()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default: ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level: debug|info|warn|error")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig(_ *cobra.Command) error {
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("read config: %w", err)
		}
	}

	setupLogger(logLevel)
	return nil
}

func setupLogger(level string) {
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
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: l})
	slog.SetDefault(slog.New(handler))
}
