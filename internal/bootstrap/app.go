package bootstrap

import (
	"log/slog"
	"net/http"

	appcatalog "git.derfenix.pro/derfenix/archiveopds/internal/application/catalog"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/archive"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/config"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/logging"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/opds"
	httpapi "git.derfenix.pro/derfenix/archiveopds/internal/interfaces/http"
)

// App — composition root: связывает домен, сценарии и адаптеры.
type App struct {
	HTTP *httpapi.Server
}

// New собирает приложение. Возвращает ошибку при невалидном конфиге или StrictIndex + сбой INPX.
func New(cfg config.Config) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log := logging.New(cfg.LogLevel)
	slog.SetDefault(log)

	nav, err := archive.NewNavigator(cfg)
	if err != nil {
		return nil, err
	}

	listSections := &appcatalog.ListSections{Nav: nav}
	listBooks := &appcatalog.ListBooks{Nav: nav}
	searchBooks := &appcatalog.SearchBooks{Nav: nav}
	streamBook := &appcatalog.StreamBook{Nav: nav}

	feed := &opds.FeedBuilder{BaseURL: cfg.BaseURL}
	h := &httpapi.OPDSHandler{
		Log:          log,
		ExposeErrors: cfg.ExposeErrors,
		ListSections: listSections,
		ListBooks:    listBooks,
		SearchBooks:  searchBooks,
		StreamBook:   streamBook,
		Feed:         feed,
	}

	mux := http.NewServeMux()
	h.Register(mux)
	httpapi.RegisterHealth(mux, cfg.ArchivePath, nav)

	inner := httpapi.WithServerHeader(mux)
	if cfg.RateLimitRPS > 0 {
		burst := int(cfg.RateLimitRPS * 2)
		if burst < 10 {
			burst = 10
		}
		if burst > 256 {
			burst = 256
		}
		inner = httpapi.WithRateLimit(httpapi.RateLimitConfig{
			RPS:            cfg.RateLimitRPS,
			Burst:          burst,
			PerIP:          cfg.RateLimitPerIP,
			TrustForwarded: cfg.RateLimitTrustForward,
			MaxTrackedIPs:  cfg.RateLimitMaxTrackedIPs,
		}, inner)
	}

	chain := httpapi.WithRequestLogger(log, inner)
	srv := httpapi.NewServer(cfg.ListenAddr, chain)
	return &App{HTTP: srv}, nil
}
