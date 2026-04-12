package catalog

import (
	"context"

	"git.derfenix.pro/derfenix/archiveopds/internal/application/port/outbound"
	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

// ListSections — корневая навигация OPDS.
type ListSections struct {
	Nav outbound.ArchiveNavigator
}

func (u *ListSections) Execute(ctx context.Context) ([]catalog.Section, error) {
	return u.Nav.RootSections(ctx)
}
