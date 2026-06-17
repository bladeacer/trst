package ansi

import (
	"regexp"
)

var (
	// Using strict, non-greedy match boundaries
	reUnderline = regexp.MustCompile(`__([^_]+?)__`)
	reHighlight = regexp.MustCompile(`==([^=]+?)==`)
	reBold      = regexp.MustCompile(`\*\*([^\*]+?)\*\*`)
	reItalic    = regexp.MustCompile(`_([^_]+?)_`)
	
	// Captures single asterisk expressions like *sigh*, *facepalm*, *rolls eyes*
	reAction    = regexp.MustCompile(`\*([^\*]+?)\*`)
)

// RenderTerminalMarkdown safely converts syntax blocks into ANSI escape sequences.
func RenderTerminalMarkdown(input string) string {
	output := input

	// 1. Underline -> ANSI Underline (\033[4m)
	output = reUnderline.ReplaceAllString(output, "\033[4m$1\033[0m")

	// 2. Highlights -> ANSI Inverted Color (\033[7m)
	output = reHighlight.ReplaceAllString(output, "\033[7m$1\033[0m")

	// 3. Bold -> ANSI Intense Bold (\033[1m)
	output = reBold.ReplaceAllString(output, "\033[1m$1\033[0m")

	// 4. Italics -> ANSI Italic (\033[3m)
	output = reItalic.ReplaceAllString(output, "\033[3m$1\033[0m")

	// 5. Single Asterisk Actions -> ANSI Dim + Italic (\033[2;3m) for theatrical sighs
	output = reAction.ReplaceAllString(output, "\033[2;3m$1\033[0m")

	return output
}
