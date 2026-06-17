package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bladeacer/trst/pkg/models"
)

type OllamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func AutoSelectOllamaModel() (string, error) {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return "", fmt.Errorf("could not connect to local Ollama on port 11434: %w", err)
	}
	defer resp.Body.Close()

	var tags OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("failed reading tags: %w", err)
	}

	if len(tags.Models) == 0 {
		fallbackModel := "llama3.2:latest"
		fmt.Printf("--- No models found on your engine. Auto-downloading %s... ---\n", fallbackModel)
		if err := PullOllamaModel(fallbackModel); err != nil {
			return "", fmt.Errorf("failed to pull fallback model: %w", err)
		}
		return fallbackModel, nil
	}

	if len(tags.Models) == 1 {
		return tags.Models[0].Name, nil
	}

	fmt.Println("--- Multiple local Ollama models found. Please select one: ---")
	for i, m := range tags.Models {
		fmt.Printf("[%d] %s\n", i+1, m.Name)
	}

	reader := os.NewFile(uintptr(os.Stdin.Fd()), "/dev/stdin")
	if reader == nil {
		return tags.Models[0].Name, nil
	}
	
	fmt.Print("Enter option number: ")
	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	var choice int
	_, err = fmt.Sscanf(input, "%d", &choice)
	if err != nil || choice < 1 || choice > len(tags.Models) {
		fmt.Println("Invalid pick. Defaulting to first option.")
		return tags.Models[0].Name, nil
	}

	return tags.Models[choice-1].Name, nil
}

func PullOllamaModel(modelName string) error {
	payload := map[string]any{
		"name":   modelName,
		"stream": false,
	}
	body, _ := json.Marshal(payload)
	
	resp, err := http.Post("http://localhost:11434/api/pull", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned bad status: %s", resp.Status)
	}
	return nil
}

func GenerateRoast(backend, model, systemPrompt string, track models.Track, jerkLevel int, allowProfanity bool) (string, error) {
	fsString := StringifyFSProperties(track)

	lyricsContext := "NOT AVAILABLE (Do NOT mention, quote, or hallucinate any lyrics for this song)."
	if track.Lyrics != "" {
		lyricsContext = fmt.Sprintf("ACTUAL VERIFIED LYRICS: %s", track.Lyrics)
	}

	profanityRule := "STRICTLY PROHIBITED. Keep your insults completely clean of curse words, slurs, or explicit vulgarity. Be clever, not dirty."
	if allowProfanity {
		profanityRule = "ALLOWED. You can use explicit language and raw profanities if it makes the roast punchier."
	}

	behaviorDirectives := fmt.Sprintf(`
BEHAVIOR MODIFIERS:
- Jerk Scale Severity: %d out of 5 (1 = playful/light teasing, 5 = absolutely brutal, soul-crushing, and unapologetically mean).
- Profanity Usage: %s
- Formatting: Use **bold** for loud callouts, and *italics* heavily for descriptive, dry sarcastic stabs. Do NOT use '#' headers.`, jerkLevel, profanityRule)

	userPrompt := fmt.Sprintf(
		"Track Title: %s\nArtist: %s\nGenre Context: %s\nBPM Context: %d\nTechnical File Stats: %s\nEmbedded Meta Comment/Description: %s\nLyrics Status: %s\n%s\n\nRoast my musical taste according to your directives.",
		track.Title, track.Artist, track.Genre, track.BPM, fsString, track.Description, lyricsContext, behaviorDirectives,
	)

	var roast string
	var err error

	if backend == "ollama" {
		roast, err = CallOllama(model, systemPrompt, userPrompt, false)
	} else {
		roast = "OpenRouter fallback handler..."
	}

	if err != nil {
		return "", err
	}

	return RenderTerminalMarkdown(roast), nil
}

func CallOllama(model, system, user string, jsonMode bool) (string, error) {
	payload := map[string]any{
		"model":  model,
		"prompt": user,
		"system": system,
		"stream": false,
	}

	if jsonMode {
		payload["format"] = "json"
		payload["options"] = map[string]any{
			"temperature": 0.0,
		}
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
	
	b, _ := io.ReadAll(resp.Body)
	json.Unmarshal(b, &ollamaResp)
	return ollamaResp.Response, nil
}
