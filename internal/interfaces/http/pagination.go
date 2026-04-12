package httpapi

import (
	"net/url"
	"strconv"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

// parseFeedPage из query: page (с 1), limit; limit по умолчанию DefaultPageSize, максимум MaxPageSize.
func parseFeedPage(q url.Values) (page, limit int) {
	limit = catalog.DefaultPageSize
	if ls := q.Get("limit"); ls != "" {
		if v, err := strconv.Atoi(ls); err == nil && v > 0 {
			limit = v
		}
	}
	if limit > catalog.MaxPageSize {
		limit = catalog.MaxPageSize
	}
	page = 1
	if ps := q.Get("page"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			page = v
		}
	}
	return page, limit
}

func pageToOffset(page, limit int) int {
	return (page - 1) * limit
}
