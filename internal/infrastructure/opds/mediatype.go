package opds

// Типы контента по OPDS 1.2. Без параметра kind= многие клиенты не считают фид каталогом
// и не показывают поиск (OpenSearch привязан к OPDS 1.2).
const (
	NavigationCatalogMediaType = "application/atom+xml;profile=opds-catalog;kind=navigation"
	AcquisitionFeedMediaType   = "application/atom+xml;profile=opds-catalog;kind=acquisition"
)
