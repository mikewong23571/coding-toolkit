package util

import (
	"encoding/json"
	"strings"
)

func TSVEscape(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

func JSONEscape(s string) string {
	b, _ := json.Marshal(s)
	if len(b) >= 2 {
		return string(b[1 : len(b)-1])
	}
	return ""
}

func ShellQuote(path string) string {
	if path == "" {
		return "''"
	}
	if !strings.ContainsAny(path, " \t\n'\"\\$") {
		return path
	}
	return "'" + strings.ReplaceAll(path, "'", "'\\''") + "'"
}

func RenderTemplate(template string, values map[string]string) string {
	out := template
	for key, val := range values {
		placeholder := "{{" + key + "}}"
		out = strings.ReplaceAll(out, placeholder, val)
	}
	return out
}
