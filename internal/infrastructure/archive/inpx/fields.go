package inpx

import (
	"strings"
)

// Разбор строки .inp: чаще всего поля разделены 0x04 (Flibusta/LibRusEc); иначе — табуляция.
func splitFields(line string) []string {
	if strings.ContainsRune(line, '\x04') {
		return strings.Split(line, "\x04")
	}
	return strings.Split(line, "\t")
}
