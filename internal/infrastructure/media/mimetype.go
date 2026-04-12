package media

import (
	"path"
	"strings"
)

// ContentTypeByExt — MIME по расширению файла (нижний регистр, с точкой).
func ContentTypeByExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".fb2":
		return "application/fb2+xml"
	case ".epub":
		return "application/epub+zip"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	case ".mobi", ".azw", ".azw3":
		return "application/x-mobipocket-ebook"
	case ".djvu", ".djv":
		return "image/vnd.djvu"
	default:
		return "application/octet-stream"
	}
}

// ContentTypeByFilename — MIME по имени файла.
func ContentTypeByFilename(name string) string {
	return ContentTypeByExt(path.Ext(name))
}

// RefineContentType — при octet-stream или пустом типе подставить тип по расширению имени.
func RefineContentType(filename, contentType string) string {
	ext := strings.ToLower(path.Ext(filename))
	if contentType != "" && contentType != "application/octet-stream" {
		return contentType
	}
	if ext != "" {
		ct := ContentTypeByExt(ext)
		if ct != "application/octet-stream" {
			return ct
		}
	}
	if contentType != "" {
		return contentType
	}
	if ext != "" {
		return "application/octet-stream"
	}
	return ""
}
