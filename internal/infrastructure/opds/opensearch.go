package opds

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/catalog"
)

type openSearchRoot struct {
	XMLName     xml.Name        `xml:"http://a9.com/-/spec/opensearch/1.1/ OpenSearchDescription"`
	ShortName   string          `xml:"ShortName"`
	Description string          `xml:"Description"`
	InputEnc    string          `xml:"InputEncoding"`
	OutputEnc   string          `xml:"OutputEncoding"`
	URLs        []openSearchURL `xml:"Url"`
}

type openSearchURL struct {
	Type     string `xml:"type,attr"`
	Method   string `xml:"method,attr"`
	Template string `xml:"template,attr"`
}

// OpenSearchDescription — XML для rel="search" в OPDS (шаблон с {searchTerms}).
func OpenSearchDescription(baseURL string) ([]byte, error) {
	base := strings.TrimRight(baseURL, "/")
	tpl := fmt.Sprintf("%s/opds/search?q={searchTerms}&page=1&limit=%d", base, catalog.DefaultPageSize)
	root := openSearchRoot{
		ShortName:   "archiveopds",
		Description: "Поиск (q, author, title, genre, series, year; page и limit — постранично, см. RFC 5005 в фиде)",
		InputEnc:    "UTF-8",
		OutputEnc:   "UTF-8",
		// OPDS 1.2 требует kind=acquisition для шаблона поиска; остальные — для старых клиентов.
		URLs: []openSearchURL{
			{Type: "application/atom+xml;profile=opds-catalog;kind=acquisition", Method: "get", Template: tpl},
			{Type: "application/atom+xml;profile=opds-catalog", Method: "get", Template: tpl},
			{Type: "application/atom+xml", Method: "get", Template: tpl},
		},
	}
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(root); err != nil {
		return nil, fmt.Errorf("opensearch xml: %w", err)
	}
	return buf.Bytes(), nil
}
