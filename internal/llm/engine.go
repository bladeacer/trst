package llm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bladeacer/trst/pkg/models"
)

type ClassificationPayload struct {
	Genre string `json:"genre"`
	BPM   int    `json:"bpm"`
}

// StringifyFSProperties converts map elements to a unified token block
func StringifyFSProperties(track models.Track) string {
	var fsClues []string
	for k, v := range track.FSProperties {
		fsClues = append(fsClues, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(fsClues, " | ")
}

// RenderTerminalMarkdown dynamically maps symbols to terminal-friendly color channels
func RenderTerminalMarkdown(input string) string {
	// 1. Bold: Matches ** bold ** or **bold**
	reBold := regexp.MustCompile(`\*\*([^\*]+)\*\*`)
	output := reBold.ReplaceAllString(input, "\033[1m$1\033[0m")

	// 2. Italics via Asterisk: Matches * italic * or *italic*
	reItalicAst := regexp.MustCompile(`\*([^\*]+)\*`)
	output = reItalicAst.ReplaceAllString(output, "\033[3m$1\033[0m")

	// 3. Italics via Underscore: Matches _italic_ or _ italic _
	reItalicUnd := regexp.MustCompile(`_([^_,]+)_`)
	output = reItalicUnd.ReplaceAllString(output, "\033[3m$1\033[0m")

	return output
}

// ExtractJSONBlock strips out surrounding model conversational filler text
func ExtractJSONBlock(input string) string {
	re := regexp.MustCompile(`\{[\s\S]*\}`)
	match := re.FindString(input)
	if match != "" {
		return match
	}
	return input
}
