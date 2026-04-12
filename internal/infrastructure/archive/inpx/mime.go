package inpx

import (
	"path"
	"strings"

	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/media"
)

// MIMEForIndexInner — тип для записи из .inp: без расширения считаем .fb2 (типично для Flibusta).
func MIMEForIndexInner(inner string) string {
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return "application/octet-stream"
	}
	if path.Ext(inner) != "" {
		return media.ContentTypeByFilename(inner)
	}
	return media.ContentTypeByFilename(inner + ".fb2")
}

func mimeFromName(name string) string {
	return media.ContentTypeByFilename(name)
}
