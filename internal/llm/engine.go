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

func StringifyFSProperties(track models.Track) string {
	var fsClues []string
	for k, v := range track.FSProperties {
		fsClues = append(fsClues, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(fsClues, " | ")
}

// RenderTerminalMarkdown processes formatting with a preference for subtle styles
func RenderTerminalMarkdown(input string) string {
	output := input

	// 1. Underline: Matches __text__ or __ text __ -> ANSI Underline (\033[4m)
	reUnderline := regexp.MustCompile(`__([^_]+)__`)
	output = reUnderline.ReplaceAllString(output, "\033[4m$1\033[0m")

	// 2. Highlights: Matches ==text== or == text == -> ANSI Inverted/Background color (\033[7m)
	reHighlight := regexp.MustCompile(`==([^=]+)==`)
	output = reHighlight.ReplaceAllString(output, "\033[7m$1\033[0m")

	// 3. Bold: Matches **text** -> ANSI Intense Bold (\033[1m)
	reBold := regexp.MustCompile(`\*\*([^\*]+)\*\*`)
	output = reBold.ReplaceAllString(output, "\033[1m$1\033[0m")

	// 4. Italics via Asterisk: Matches *text* -> ANSI Italic (\033[3m)
	reItalicAst := regexp.MustCompile(`\*([^\*]+)\*`)
	output = reItalicAst.ReplaceAllString(output, "\033[3m$1\033[0m")

	// 5. Italics via Single Underscore: Matches _text_ -> ANSI Italic (\033[3m)
	reItalicUnd := regexp.MustCompile(`_([^_]+)_`)
	output = reItalicUnd.ReplaceAllString(output, "\033[3m$1\033[0m")

	return output
}

func ExtractJSONBlock(input string) string {
	re := regexp.MustCompile(`\{[\s\S]*\}`)
	match := re.FindString(input)
	if match != "" {
		return match
	}
	return input
}
