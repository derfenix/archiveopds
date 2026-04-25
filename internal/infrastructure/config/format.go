package config

import (
	"fmt"
	"io"
	"strings"
)

// WriteText печатает эффективную конфигурацию и подсказку по env.
func WriteText(w io.Writer, c Config) {
	_, _ = fmt.Fprintf(w, "listen:      %s\n", showEmpty(c.ListenAddr))
	_, _ = fmt.Fprintf(w, "base_url:    %s\n", showEmpty(c.BaseURL))
	_, _ = fmt.Fprintf(w, "archive:     %s\n", showEmpty(c.ArchivePath))
	_, _ = fmt.Fprintf(w, "log_level:   %s\n", showEmpty(c.LogLevel))
	_, _ = fmt.Fprintf(w, "strict_index: %v\n", c.StrictIndex)
	_, _ = fmt.Fprintf(w, "expose_errors: %v\n", c.ExposeErrors)
	_, _ = fmt.Fprintf(w, "rate_limit_rps: %v\n", c.RateLimitRPS)
	_, _ = fmt.Fprintf(w, "rate_limit_per_ip: %v\n", c.RateLimitPerIP)
	_, _ = fmt.Fprintf(w, "rate_limit_trust_forward: %v\n", c.RateLimitTrustForward)
	_, _ = fmt.Fprintf(w, "rate_limit_max_tracked_ips: %d\n", c.RateLimitMaxTrackedIPs)
	_, _ = fmt.Fprintf(w, "max_open_zip_volumes: %d\n", c.MaxOpenZipVolumes)
	_, _ = fmt.Fprintf(w, "annotation_workers: %d\n", c.AnnotationWorkers)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Переменные окружения:")
	_, _ = fmt.Fprintf(w, "  %s   — адрес прослушивания\n", EnvListen)
	_, _ = fmt.Fprintf(w, "  %s — публичный базовый URL для OPDS\n", EnvBaseURL)
	_, _ = fmt.Fprintf(w, "  %s  — путь к архиву / корню хранилища\n", EnvArchive)
	_, _ = fmt.Fprintf(w, "  %s — уровень логов slog: debug, info, warn, error\n", EnvLogLevel)
	_, _ = fmt.Fprintf(w, "  %s — при ошибке INPX завершать процесс (1/true)\n", EnvStrictIndex)
	_, _ = fmt.Fprintf(w, "  %s — отдавать клиенту текст внутренних ошибок (1/true)\n", EnvExposeErrors)
	_, _ = fmt.Fprintf(w, "  %s — лимит запросов/с (0 = выкл)\n", EnvRateLimitRPS)
	_, _ = fmt.Fprintf(w, "  %s — отдельный лимит на клиентский IP (1/true) или один на процесс (0/false)\n", EnvRateLimitPerIP)
	_, _ = fmt.Fprintf(w, "  %s — учитывать X-Forwarded-For для per-IP (1/true; только за доверенным прокси)\n", EnvRateLimitTrustForward)
	_, _ = fmt.Fprintf(w, "  %s — макс. отдельных per-IP лимитеров; при переполнении вытесняется старый (0=без капа)\n", EnvRateLimitMaxTrackedIPs)
	_, _ = fmt.Fprintf(w, "  %s — макс. кэшированных открытых .zip (0=без капа)\n", EnvMaxOpenZipVolumes)
	_, _ = fmt.Fprintf(w, "  %s — параллельность чтения FB2-аннотаций (мин. 1)\n", EnvAnnotationWorkers)
}

func showEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(пусто)"
	}
	return s
}
