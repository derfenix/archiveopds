package inpx

import (
	"os"
	"path/filepath"
	"strings"
)

// FindInRoot возвращает пути ко всем *.inpx в каталоге root (не рекурсивно).
func FindInRoot(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".inpx") {
			out = append(out, filepath.Join(root, e.Name()))
		}
	}
	return out, nil
}
