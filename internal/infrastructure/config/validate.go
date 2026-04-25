package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Validate проверяет обязательные и формально корректные поля.
func (c Config) Validate() error {
	if err := ValidateBaseURL(c.BaseURL); err != nil {
		return err
	}
	if c.RateLimitRPS < 0 {
		return fmt.Errorf("rate_limit_rps: не может быть отрицательным")
	}
	if c.RateLimitMaxTrackedIPs < 0 {
		return fmt.Errorf("rate_limit_max_tracked_ips: не может быть отрицательным (0 = без лимита)")
	}
	if c.MaxOpenZipVolumes < 0 {
		return fmt.Errorf("max_open_zip_volumes: не может быть отрицательным (0 = без лимита)")
	}
	if c.AnnotationWorkers < 1 {
		return fmt.Errorf("annotation_workers: минимум 1 (1 = последовательно)")
	}
	if c.AnnotationWorkers > 64 {
		return fmt.Errorf("annotation_workers: не более 64")
	}
	return nil
}

// ValidateBaseURL — для OPDS нужен абсолютный URL со схемой http(s) и хостом.
func ValidateBaseURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("base_url: пусто (нужен полный URL, например http://127.0.0.1:8080)")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("base_url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("base_url: схема должна быть http или https, не %q", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("base_url: укажите хост (например http://192.168.1.2:8080)")
	}
	return nil
}
