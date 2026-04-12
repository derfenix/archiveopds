package archive

import (
	"fmt"
	"log/slog"
	"strings"

	"git.derfenix.pro/derfenix/archiveopds/internal/application/port/outbound"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/archive/inpx"
	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/config"
)

// NewNavigator возвращает навигатор по INPX в корне ARCHIVEOPDS_ARCHIVE, если есть *.inpx;
// иначе — заглушку. При StrictIndex и ошибке индекса возвращает ошибку.
func NewNavigator(cfg config.Config) (outbound.ArchiveNavigator, error) {
	root := strings.TrimSpace(cfg.ArchivePath)
	if root == "" {
		return NewNavigatorStub(), nil
	}

	paths, err := inpx.FindInRoot(root)
	if err != nil {
		if cfg.StrictIndex {
			return nil, fmt.Errorf("inpx в %q: поиск каталога: %w", root, err)
		}
		slog.Warn("inpx find in root failed", "root", root, "err", err)
		return NewNavigatorStub(), nil
	}
	if len(paths) == 0 {
		if cfg.StrictIndex {
			return nil, fmt.Errorf("inpx: в %q не найдено файлов *.inpx", root)
		}
		return NewNavigatorStub(), nil
	}

	nav, err := inpx.Load(root, paths, cfg.AnnotationWorkers)
	if err != nil {
		if cfg.StrictIndex {
			return nil, fmt.Errorf("inpx load %q: %w", root, err)
		}
		slog.Warn("inpx load failed", "root", root, "err", err)
		return NewNavigatorStub(), nil
	}
	return nav, nil
}
