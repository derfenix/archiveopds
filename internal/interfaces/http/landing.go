package httpapi

import (
	"html"
	"net/http"
	"strings"
	"text/template"

	"git.derfenix.pro/derfenix/archiveopds/internal/infrastructure/opds"
)

// landingPage — корень сервера для браузера: человекочитаемая страница с ссылкой на OPDS.
// Клиенты, которые ищут каталог через HTML-документ, находят фид по <link rel="alternate">.
const landingPage = `<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>archiveopds</title>
<link rel="alternate" type="{{.FeedType}}" title="OPDS каталог" href="{{.Catalog}}">
<link rel="search" type="application/opensearchdescription+xml" href="{{.Search}}">
<link rel="icon" href="/favicon.ico" type="image/svg+xml">
<style>
  body { margin: 0; min-height: 100vh; display: flex; align-items: center; justify-content: center;
         font: 16px/1.5 system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
         background: #f6f7f9; color: #1a202c; }
  main { max-width: 34rem; padding: 2rem; }
  h1 { margin: 0 0 .25rem; font-size: 1.6rem; }
  p { margin: .5rem 0; }
  code { background: #e2e8f0; border-radius: 4px; padding: .1rem .35rem; }
  a { color: #2b6cb0; }
  ul { padding-left: 1.2rem; }
</style>
</head>
<body>
<main>
  <h1>archiveopds</h1>
  <p>OPDS-каталог: <a href="{{.Catalog}}">{{.Catalog}}</a></p>
  <p>Этот адрес укажите в читалке с поддержкой OPDS (Moon+ Reader, KOReader, Calibre и др.).</p>
  <ul>
    <li><a href="{{.Catalog}}">Открыть каталог как фид</a></li>
    <li><a href="{{.Search}}">Описание поиска (OpenSearch)</a></li>
  </ul>
  <p><small>HTTP API: <code>/opds</code>, <code>/opds/section/&lt;id&gt;</code>,
  <code>/opds/search</code>, <code>/opds/acquire/&lt;id&gt;</code>, <code>/healthz</code>, <code>/readyz</code></small></p>
</main>
</body>
</html>
`

var landingTemplate = template.Must(template.New("landing").Parse(landingPage))

// favicon — маленький SVG (браузеры принимают его для /favicon.ico, если указан type).
const faviconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">` +
	`<rect width="32" height="32" rx="7" fill="#2b6cb0"/>` +
	`<path d="M8 8h7.5v16H10a2 2 0 0 1-2-2V8zm8.5 0H24v14a2 2 0 0 1-2 2h-5.5V8z" fill="#fff" opacity=".92"/>` +
	`</svg>`

// handleLanding отдаёт страницу на / вместо 404.
func (h *OPDSHandler) handleLanding(w http.ResponseWriter, r *http.Request) {
	base := strings.TrimRight(h.Feed.BaseURL, "/")
	if base == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		base = scheme + "://" + r.Host
	}
	data := struct {
		Catalog  string
		Search   string
		FeedType string
	}{
		// text/template без автоэкранирования: html/template превращает «+» в MIME-типе
		// в &#43;, и поиск каталога по сырому HTML становится ненадёжным.
		Catalog:  html.EscapeString(base + opds.CatalogRootPath),
		Search:   html.EscapeString(base + "/opds/opensearch.xml"),
		FeedType: opds.NavigationCatalogMediaType,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := landingTemplate.Execute(w, data); err != nil {
		h.log().Debug("landing: template error", "err", err)
	}
}

// handleFavicon отдаёт иконку, чтобы браузер не сыпал 404 в лог.
func (h *OPDSHandler) handleFavicon(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write([]byte(faviconSVG))
}
