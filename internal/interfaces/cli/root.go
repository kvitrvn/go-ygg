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
	cfgFile string
	v       *viper.Viper
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
	rootCmd.PersistentFlags().String("log-level", "", "log level: debug|info|warn|error (overrides config and GO_YGG_LOG_LEVEL)")
	rootCmd.PersistentFlags().String("log-format", "", "log format: json|text (overrides config and GO_YGG_LOG_FORMAT)")

	_ = v.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = v.BindPFlag("log.format", rootCmd.PersistentFlags().Lookup("log-format"))

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

	setupLogger(v.GetString("log.level"), v.GetString("log.format"))
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
