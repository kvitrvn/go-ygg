package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kvitrvn/go-ygg/internal/version"
)

var versionCmd = &cobra.Command{
	Use:               "version",
	Short:             "Print version information",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error { return nil },
	RunE: func(_ *cobra.Command, _ []string) error {
		info := version.Get()
		fmt.Printf("version:    %s\ncommit:     %s\nbuild date: %s\n", info.Version, info.Commit, info.BuildDate)
		return nil
	},
}
