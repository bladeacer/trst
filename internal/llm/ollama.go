package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bladeacer/trst/internal/persona"
	"github.com/bladeacer/trst/internal/ansi"
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

func GenerateRoast(backend, model, systemPromptName string, track models.Track, jerkLevel int, allowProfanity bool) (string, error) {
	// 1. Map dynamic lyrics profile parameters
	lyricsContext := "NOT AVAILABLE"
	if track.Lyrics != "" {
		lyricsContext = fmt.Sprintf("Lyrics: %s", track.Lyrics)
	}

	profanityRule := "STRICTLY PROHIBITED. Do not use curse words."
	if allowProfanity {
		profanityRule = "ALLOWED. You can use profanity."
	}

	// 2. Fetch structural core persona using the dynamic jerk scale selector architecture
	baseSystemPrompt := persona.GetSystemPrompt(systemPromptName, jerkLevel)

	// 3. Assemble Spacing Rules cleanly
	extraDirectives := ""
	if strings.ToLower(strings.TrimSpace(systemPromptName)) == "spitter" {
		extraDirectives = "\n- Spacing Rule: End every single bar/sentence with a raw newline character instantly. Do NOT compile sentences into paragraphs."
	}

	rigorousSystemPrompt := fmt.Sprintf("%s\n\nCRITICAL SYSTEM DIRECTIVES:\n- Profanity: %s\n- Guardrails: Do NOT use markdown headers (#), lists, or bullet points.%s", 
		baseSystemPrompt, profanityRule, extraDirectives)

	fsString := StringifyFSProperties(track)

	// 4. Construct user payload matrix
	userPrompt := fmt.Sprintf(`Track Title: %s
Artist: %s
Genre: %s
BPM: %d
File Stats: %s
Description: %s
Lyrics: %s

FORMATTING RULES:
- Use _italics_ for overly sarcastic adjectives.
- Use __underlines__ to call out direct tech numbers and stats (e.g., __120 BPM__).
- Use **bold** exclusively for extreme emphasis or frustration.
- Do NOT use highlights (==).

CRITICAL FORMAT ATTACK:
- If the File Stats or Codec indicate the track is using the OPUS (.opus) format, you must viciously roast the file container format itself, mocking its lossy compression, discord-tier stream quality, or budget storage profile.

Execute the roast matching these style rules now.`,
		track.Title, 
		track.Artist, 
		track.Genre, 
		track.BPM, 
		fsString,
		track.Description, 
		lyricsContext,
	)

	var roast string
	var err error

	if backend == "ollama" {
		roast, err = CallOllama(model, rigorousSystemPrompt, userPrompt)
	} else {
		return "", fmt.Errorf("unsupported backend: %s", backend)
	}

	if err != nil {
		return "", err
	}

	return ansi.RenderTerminalMarkdown(roast), nil
}

func CallOllama(model, system, user string) (string, error) {
	payload := map[string]any{
		"model":  model,
		"prompt": user,
		"system": system,
		"stream": false,
		"options": map[string]any{
			"temperature": 0.7, // Balances creative insult variance with formatting compliance
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
	b, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(b, &ollamaResp)

	return ollamaResp.Response, nil
}
