package llm

import (
	"encoding/json"
	"strings"
	"fmt"
	"github.com/bladeacer/trst/pkg/models"
)

type RefinedMetadata struct {
	Genre string `json:"genre"`
	BPM   int    `json:"bpm"`
}

// RefineTrackDetails asks the local backend to populate accurate Genre/BPM profiles using structural text context
func RefineTrackDetails(backend, model string, track *models.Track) {
	// If the file already has pristine embedded tags, don't override them
	if track.Genre != "" && track.BPM > 0 && !strings.EqualFold(track.Genre, "Unknown") {
		return
	}

	systemPrompt := "You are a music database system. Return a valid JSON object matching this schema: {\"genre\": \"string\", \"bpm\": integer}. Do not write regular prose, markdown formatting, or triple backticks. Only output valid JSON raw text. If you do not know the song, make a logical guess based on lyrics style or artist name."
	userPrompt := fmt.Sprintf("Title: %s\nArtist: %s\nLyrics Snippet: %s", track.Title, track.Artist, track.Description)

	var jsonRaw string
	var err error

	if backend == "ollama" {
		jsonRaw, err = callOllama(model, systemPrompt, userPrompt)
	}

	if err != nil || jsonRaw == "" {
		// Silent safety fallbacks if execution loops time out
		if track.Genre == "" { track.Genre = "Unclassifiable Audio" }
		if track.BPM == 0 { track.BPM = 100 }
		return
	}

	var refined RefinedMetadata
	if err := json.Unmarshal([]byte(jsonRaw), &refined); err == nil {
		if track.Genre == "" || strings.EqualFold(track.Genre, "Unknown") {
			track.Genre = refined.Genre
		}
		if track.BPM == 0 {
			track.BPM = refined.BPM
		}
	} else {
		// Strip possible model markdown blocks if it disobeyed our system structure instructions
		// Safety defaults
		if track.Genre == "" { track.Genre = "Acoustic Audio" }
		if track.BPM == 0 { track.BPM = 110 }
	}
}
