package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/bladeacer/trst/pkg/models"
)

type ClassificationPayload struct {
	Genre string `json:"genre"`
	BPM   int    `json:"bpm"`
}

func RefineTrackDetails(backend, model string, track *models.Track) {
	systemPrompt := `You are an expert musicology classifier database. Analyze the provided metadata details and context strings.
Deduce the exact musical subgenre (e.g., "Liquid Drum & Bass", "French House", "J-Pop/Anisong", "Math Rock", "Tech House", "Hardstyle").
Provide a realistic tempo BPM matching that subgenre style standard baseline.

CRITICAL: Return ONLY a valid JSON object matching this schema. Do not write markdown, do not write code blocks, do not add introductory text or chat greetings.
JSON Schema: {"genre": "string", "bpm": integer}`

	// Cleaned user prompt: Removed the instruction to "roast" so the refiner does only classification
	var fsClues []string
	for k, v := range track.FSProperties {
		fsClues = append(fsClues, fmt.Sprintf("%s: %s", k, v))
	}
	fsString := strings.Join(fsClues, " | ")

	userPrompt := fmt.Sprintf(
		"Track Title: %s\nArtist: %s\nInitial Tag Suggestion: %s\nTechnical File Stats: %s\nEmbedded Meta Comment/Description: %s",
		track.Title, track.Artist, track.Genre, fsString, track.Description,
	)

	var jsonRaw string
	var err error

	if backend == "ollama" {
		jsonRaw, err = callOllamaJSON(model, systemPrompt, userPrompt)
	}

	if err != nil || jsonRaw == "" {
		return // Gracefully fall back to local structural defaults
	}

	// Extract JSON using regex in case the model ignored directions and wrapped text around it
	jsonRaw = extractJSONBlock(jsonRaw)

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

// callOllamaJSON uses Ollama's native "format": "json" grammar constraint driver
func callOllamaJSON(model, system, user string) (string, error) {
	payload := map[string]any{
		"model":  model,
		"prompt": user,
		"system": system,
		"stream": false,
		"format": "json", // Forces Ollama's sampler layer to strictly generate valid structural data
		"options": map[string]any{
			"temperature": 0.0, // Eliminate creative randomness for true deterministic parsing
		},
	}
	
	body, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var ollamaResp struct {
		Response string `json:"response"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}
	return ollamaResp.Response, nil
}

// extractJSONBlock strips non-JSON leading/trailing clutter
func extractJSONBlock(input string) string {
	re := regexp.MustCompile(`\{[\s\S]*\}`)
	match := re.FindString(input)
	if match != "" {
		return match
	}
	return input
}
