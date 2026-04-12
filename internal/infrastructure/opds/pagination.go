package opds

import (
	"net/url"
	"strconv"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

// SectionEntryQuery — первая страница раздела для ссылок из навигационного фида.
func SectionEntryQuery() string {
	return "page=1&limit=" + strconv.Itoa(catalog.DefaultPageSize)
}

func (f *FeedBuilder) paginationLinks(path string, total, page, limit int, rawQuery string) []atomLink {
	if limit <= 0 {
		limit = catalog.DefaultPageSize
	}
	v, err := url.ParseQuery(rawQuery)
	if err != nil {
		v = url.Values{}
	}
	v.Set("limit", strconv.Itoa(limit))
	v.Set("page", strconv.Itoa(page))

	self := f.join(path + "?" + v.Encode())
	var links []atomLink
	links = append(links, atomLink{Rel: "self", Href: self, Type: AcquisitionFeedMediaType})

	lastPage := (total + limit - 1) / limit
	if lastPage < 1 {
		lastPage = 1
	}

	v.Set("page", "1")
	links = append(links, atomLink{Rel: "first", Href: f.join(path + "?" + v.Encode()), Type: AcquisitionFeedMediaType})

	if page > 1 {
		v.Set("page", strconv.Itoa(page-1))
		links = append(links, atomLink{Rel: "previous", Href: f.join(path + "?" + v.Encode()), Type: AcquisitionFeedMediaType})
	}
	if total > 0 && page < lastPage {
		v.Set("page", strconv.Itoa(page+1))
		links = append(links, atomLink{Rel: "next", Href: f.join(path + "?" + v.Encode()), Type: AcquisitionFeedMediaType})
	}
	if lastPage > 1 {
		v.Set("page", strconv.Itoa(lastPage))
		links = append(links, atomLink{Rel: "last", Href: f.join(path + "?" + v.Encode()), Type: AcquisitionFeedMediaType})
	}
	return links
}
