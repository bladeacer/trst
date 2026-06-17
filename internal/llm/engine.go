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

func ExtractJSONBlock(input string) string {
	re := regexp.MustCompile(`\{[\s\S]*\}`)
	match := re.FindString(input)
	if match != "" {
		return match
	}
	return input
}
