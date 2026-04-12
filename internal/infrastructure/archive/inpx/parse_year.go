package inpx

import (
	"regexp"
	"strconv"
	"strings"
)

var yearWordRE = regexp.MustCompile(`\b(19|20)\d{2}\b`)

// yearFromINPFields извлекает год из типичных полей даты Flibusta/LibRusEc (поля 9–11).
func yearFromINPFields(fields []string) int {
	for i := 9; i < len(fields) && i < 12; i++ {
		if y := parseYearString(fields[i]); y > 0 {
			return y
		}
	}
	return 0
}

func parseYearString(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// YYYYMMDD
	if len(s) >= 8 && isAllASCIIDigits(s[:8]) {
		if y, err := strconv.Atoi(s[:4]); err == nil && y >= 1000 && y <= 2100 {
			return y
		}
	}
	// Подстрока YYYY в датах вида DD-MM-YYYY
	if m := yearWordRE.FindString(s); m != "" {
		if y, err := strconv.Atoi(m); err == nil {
			return y
		}
	}
	return 0
}

func isAllASCIIDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
