package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/config"
)

func bindRuntimeFlags(cmd *cobra.Command, cfg *config.Config) {
	cmd.Flags().StringVar(&cfg.ListenAddr, "listen", cfg.ListenAddr,
		fmt.Sprintf("адрес прослушивания (env %s)", config.EnvListen))
	cmd.Flags().StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL,
		fmt.Sprintf("публичный базовый URL для ссылок в фидах (env %s)", config.EnvBaseURL))
	cmd.Flags().StringVar(&cfg.ArchivePath, "archive", cfg.ArchivePath,
		fmt.Sprintf("путь к архиву или корню хранилища (env %s)", config.EnvArchive))
	cmd.Flags().StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel,
		fmt.Sprintf("уровень slog: debug, info, warn, error (env %s)", config.EnvLogLevel))
	cmd.Flags().BoolVar(&cfg.StrictIndex, "strict-index", cfg.StrictIndex,
		fmt.Sprintf("завершать запуск при ошибке INPX (env %s)", config.EnvStrictIndex))
	cmd.Flags().BoolVar(&cfg.ExposeErrors, "expose-errors", cfg.ExposeErrors,
		fmt.Sprintf("отдавать клиенту текст внутренних ошибок (env %s)", config.EnvExposeErrors))
	cmd.Flags().Float64Var(&cfg.RateLimitRPS, "rate-limit-rps", cfg.RateLimitRPS,
		fmt.Sprintf("лимит запросов/с, 0=выкл (env %s)", config.EnvRateLimitRPS))
	cmd.Flags().BoolVar(&cfg.RateLimitPerIP, "rate-limit-per-ip", cfg.RateLimitPerIP,
		fmt.Sprintf("отдельный лимит на IP-клиент (env %s)", config.EnvRateLimitPerIP))
	cmd.Flags().BoolVar(&cfg.RateLimitTrustForward, "rate-limit-trust-forward", cfg.RateLimitTrustForward,
		fmt.Sprintf("для per-IP: доверять X-Forwarded-For (только за доверенным прокси, env %s)", config.EnvRateLimitTrustForward))
	cmd.Flags().IntVar(&cfg.RateLimitMaxTrackedIPs, "rate-limit-max-tracked-ips", cfg.RateLimitMaxTrackedIPs,
		fmt.Sprintf("макс. число отдельных per-IP лимитеров, 0=без капа (env %s)", config.EnvRateLimitMaxTrackedIPs))
	cmd.Flags().IntVar(&cfg.MaxOpenZipVolumes, "max-open-zip-volumes", cfg.MaxOpenZipVolumes,
		fmt.Sprintf("макс. открытых .zip в кэше (LRU), 0=без лимита (env %s)", config.EnvMaxOpenZipVolumes))
	cmd.Flags().IntVar(&cfg.AnnotationWorkers, "annotation-workers", cfg.AnnotationWorkers,
		fmt.Sprintf("параллельность FB2-аннотаций, мин. 1 (env %s)", config.EnvAnnotationWorkers))
}
