package config

import (
	"os"
	"strconv"
	"strings"
)

// Имена переменных окружения (префикс ARCHIVEOPDS_).
const (
	EnvListen            = "ARCHIVEOPDS_LISTEN"
	EnvBaseURL           = "ARCHIVEOPDS_BASE_URL"
	EnvArchive           = "ARCHIVEOPDS_ARCHIVE"
	EnvLogLevel          = "ARCHIVEOPDS_LOG_LEVEL"
	EnvStrictIndex       = "ARCHIVEOPDS_STRICT_INDEX"
	EnvExposeErrors      = "ARCHIVEOPDS_EXPOSE_ERRORS"
	EnvRateLimitRPS      = "ARCHIVEOPDS_RATE_LIMIT_RPS"
	EnvAnnotationWorkers = "ARCHIVEOPDS_ANNOTATION_WORKERS"
)

// ApplyEnv перезаписывает поля cfg непустыми значениями из окружения.
func ApplyEnv(cfg *Config) {
	if v := strings.TrimSpace(os.Getenv(EnvListen)); v != "" {
		cfg.ListenAddr = v
	}
	if v := strings.TrimSpace(os.Getenv(EnvBaseURL)); v != "" {
		cfg.BaseURL = v
	}
	if v := strings.TrimSpace(os.Getenv(EnvArchive)); v != "" {
		cfg.ArchivePath = v
	}
	if v := strings.TrimSpace(os.Getenv(EnvLogLevel)); v != "" {
		cfg.LogLevel = v
	}
	if envBool(EnvStrictIndex) {
		cfg.StrictIndex = true
	}
	if envBool(EnvExposeErrors) {
		cfg.ExposeErrors = true
	}
	if v := strings.TrimSpace(os.Getenv(EnvRateLimitRPS)); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.RateLimitRPS = f
		}
	}
	if v := strings.TrimSpace(os.Getenv(EnvAnnotationWorkers)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.AnnotationWorkers = n
		}
	}
}

func envBool(key string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// Load возвращает конфигурацию: значения по умолчанию, затем env.
func Load() Config {
	c := Default()
	ApplyEnv(&c)
	return c
}
