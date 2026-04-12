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
	// RateLimitRPS — глобальный лимит запросов/сек на процесс; 0 — без ограничения.
	RateLimitRPS float64
	// AnnotationWorkers — параллельное чтение <annotation> из FB2 при построении страниц OPDS (минимум 1).
	AnnotationWorkers int
}

func Default() Config {
	return Config{
		ListenAddr:   ":8080",
		BaseURL:      "http://127.0.0.1:8080",
		ArchivePath:  "",
		LogLevel:     "info",
		StrictIndex:  false,
		ExposeErrors:      false,
		RateLimitRPS:      0,
		AnnotationWorkers: 4,
	}
}
