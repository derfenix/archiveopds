package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"git.derfenix.pro/derfenix/archiveopds/internal/bootstrap"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/config"
)

func newServe() *cobra.Command {
	cfg := config.Load()

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Запустить HTTP-сервер OPDS",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := bootstrap.New(cfg)
			if err != nil {
				return err
			}
			cmd.Printf("OPDS listening on %s (base URL %s)\n", cfg.ListenAddr, cfg.BaseURL)

			errCh := make(chan error, 1)
			go func() { errCh <- app.HTTP.ListenAndServe() }()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

			select {
			case err := <-errCh:
				return err
			case <-sigCh:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := app.HTTP.Shutdown(ctx); err != nil {
					return err
				}
				return <-errCh
			}
		},
	}

	bindRuntimeFlags(cmd, &cfg)
	return cmd
}
