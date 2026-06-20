package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bladeacer/trst/pkg/models"
)

func RefineTrackDetails(backend, model string, track *models.Track) {
	// Added explicit constraint directly into the system message context
	systemPrompt := `You are an expert musicology classifier database. Analyze the provided metadata details and context strings.
Deduce the exact musical subgenre (e.g., "Liquid Drum & Bass", "French House", "J-Pop", "Math Rock", "Tech House", "Hardstyle").
Provide a realistic tempo BPM matching that subgenre style standard baseline.

CRITICAL: Return ONLY a raw JSON object matching this schema. Do not write markdown, do not write code blocks, do not add chat greetings.
JSON Schema: {"genre": "string", "bpm": integer}`

	fsString := StringifyFSProperties(*track)

	userPrompt := fmt.Sprintf(
		"Track Title: %s\nArtist: %s\nInitial Tag Suggestion: %s\nTechnical File Stats: %s\nEmbedded Meta Comment/Description: %s",
		track.Title, track.Artist, track.Genre, fsString, track.Description,
	)

	var jsonRaw string
	var err error

	if backend == "ollama" {
		// FIX: Dropped the 4th parameter 'true' to match CallOllama(model, system, user)
		jsonRaw, err = CallOllama(model, systemPrompt, userPrompt)
	}

	if err != nil || jsonRaw == "" {
		return
	}

	jsonRaw = ExtractJSONBlock(jsonRaw)

	var result ClassificationPayload
	if err := json.Unmarshal([]byte(jsonRaw), &result); err == nil {
		if result.Genre != "" && !strings.EqualFold(result.Genre, "unknown") {
			track.Genre = result.Genre
		}
		if result.BPM > 0 {
			track.BPM = result.BPM
		}
	}
}
