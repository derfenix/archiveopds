package config

// Config — настройки процесса (флаги, env). Расширять по мере появления фич.
type Config struct {
	ListenAddr  string
	BaseURL     string
	ArchivePath string
	// LogLevel: debug, info, warn, error (slog).
	LogLevel string

	// StrictIndex — при ошибке загрузки INPX завершать запуск, а не подставлять пустой каталог.
	StrictIndex bool
	// ExposeErrors — отдавать клиенту текст внутренних ошибок (только для отладки).
	ExposeErrors bool
	// RateLimitRPS — лимит запросов/сек; 0 — без ограничения. When RateLimitPerIP is set, per client IP; otherwise one bucket for the process.
	RateLimitRPS float64
	// RateLimitPerIP — if true, each client address gets its own limiter; if false, one global limiter.
	RateLimitPerIP bool
	// RateLimitTrustForward — if true, use the first X-Forwarded-For address (only if behind a trusted reverse proxy).
	RateLimitTrustForward bool
	// RateLimitMaxTrackedIPs — upper bound for distinct per-IP limiters when PerIP is true; 0 = unlimited. When at capacity, the least-recently-used entry is evicted.
	RateLimitMaxTrackedIPs int
	// MaxOpenZipVolumes — max cached open .zip files (LRU+refcount; empty slots evicted when idle). 0 = unlimited.
	MaxOpenZipVolumes int
	// AnnotationWorkers — параллельное чтение <annotation> из FB2 при построении страниц OPDS (минимум 1).
	AnnotationWorkers int
}

func Default() Config {
	return Config{
		ListenAddr:             ":8080",
		BaseURL:                "http://127.0.0.1:8080",
		ArchivePath:            "",
		LogLevel:               "info",
		StrictIndex:            false,
		ExposeErrors:           false,
		RateLimitRPS:           0,
		RateLimitPerIP:         true,
		RateLimitTrustForward:  false,
		RateLimitMaxTrackedIPs: 4096,
		MaxOpenZipVolumes:      512,
		AnnotationWorkers:      4,
	}
}
