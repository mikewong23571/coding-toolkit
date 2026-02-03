package status

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FormatLine(layout, category, repo, branch, intent string, rightLen int) string {
	left := strings.Join([]string{layout, category, repo, branch}, " | ")
	cleanIntent := strings.ReplaceAll(intent, "\n", " ")
	cleanIntent = strings.TrimSpace(cleanIntent)
	if rightLen > 0 && len(cleanIntent) > rightLen {
		cleanIntent = cleanIntent[:rightLen]
	}
	if cleanIntent == "" {
		return left
	}
	return fmt.Sprintf("%s || %s", left, cleanIntent)
}

func WriteFile(path, line string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(line+"\n"), 0o644)
}
