package cli

import (
	"github.com/spf13/cobra"

	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/config"
)

func newConfigCmd() *cobra.Command {
	cfg := config.Load()

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Показать эффективную конфигурацию (по умолчанию + env + флаги)",
		RunE: func(cmd *cobra.Command, args []string) error {
			config.WriteText(cmd.OutOrStdout(), cfg)
			return nil
		},
	}

	bindRuntimeFlags(cmd, &cfg)
	return cmd
}
