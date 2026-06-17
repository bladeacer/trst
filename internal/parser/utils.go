package parser

import (
	"regexp"
	"strings"
)

func cleanLyrics(raw string) string {
	// Strip standard timestamp syntax patterns like [00:11.22] safely
	re := regexp.MustCompile(`\[\d+:\d+[\.\:]?\d*\]`)
	var lines []string

	for _, line := range strings.Split(raw, "\n") {
		cleanedLine := re.ReplaceAllString(line, "")
		trimmed := strings.TrimSpace(cleanedLine)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	if len(lines) > 30 {
		lines = lines[:30]
	}
	return strings.Join(lines, " / ")
}
