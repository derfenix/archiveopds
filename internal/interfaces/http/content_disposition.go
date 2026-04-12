package httpapi

import (
	"fmt"
	"strings"
)

// contentDispositionAttachment — RFC 6266 / RFC 5987: ASCII fallback + UTF-8 filename*.
func contentDispositionAttachment(filename string) string {
	ascii := asciiFilenameFallback(filename)
	ascii = strings.ReplaceAll(ascii, `\`, `\\`)
	ascii = strings.ReplaceAll(ascii, `"`, `\"`)
	star := percentEncodeRFC5987(filename)
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, ascii, star)
}

func asciiFilenameFallback(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x20 && r < 0x7f && r != '"' && r != '\\' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "download"
	}
	return out
}

func percentEncodeRFC5987(s string) string {
	const hex = "0123456789ABCDEF"
	var buf strings.Builder
	for _, c := range []byte(s) {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '.' || c == '-' || c == '_' || c == '~' {
			buf.WriteByte(c)
		} else {
			buf.WriteByte('%')
			buf.WriteByte(hex[c>>4])
			buf.WriteByte(hex[c&0xf])
		}
	}
	return buf.String()
}
