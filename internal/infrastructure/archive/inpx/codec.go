package inpx

import (
	"bytes"
	"encoding/base64"
	"strings"

	"git.derfenix.pro/derfenix/archiveopds/internal/domain/book"
)

const bookIDPrefix = "x1"

func encodeBookRef(zipStem, innerPath string) book.ID {
	payload := zipStem + "\x1e" + innerPath
	return book.ID(bookIDPrefix + base64.RawURLEncoding.EncodeToString([]byte(payload)))
}

func decodeBookRef(id book.ID) (zipStem, innerPath string, ok bool) {
	s := string(id)
	if !strings.HasPrefix(s, bookIDPrefix) {
		return "", "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(s, bookIDPrefix))
	if err != nil {
		return "", "", false
	}
	i := bytes.IndexByte(raw, 0x1e)
	if i <= 0 || i >= len(raw)-1 {
		return "", "", false
	}
	return string(raw[:i]), string(raw[i+1:]), true
}
