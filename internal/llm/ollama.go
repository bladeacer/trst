package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
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
		if err := pullOllamaModel(fallbackModel); err != nil {
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

func pullOllamaModel(modelName string) error {
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

func GenerateRoast(backend, model, systemPrompt string, track models.Track) (string, error) {
	// Format a structured payload ensuring lyrics and properties are front and center
	userPrompt := fmt.Sprintf(
		"Track Title: %s\nArtist: %s\nGenre Context: %s\nBPM Context: %d\nFull Audio Clues & Lyrics: %s\n\nRoast my musical taste mercilessly.",
		track.Title, track.Artist, track.Genre, track.BPM, track.Description,
	)

	var roast string
	var err error

	if backend == "ollama" {
		roast, err = callOllama(model, systemPrompt, userPrompt)
	} else {
		roast = "OpenRouter fallback handler..."
	}

	if err != nil {
		return "", err
	}

	// Format terminal display text
	return renderTerminalMarkdown(roast), nil
}

func callOllama(model, system, user string) (string, error) {
	payload := map[string]any{
		"model":  model,
		"prompt": user,
		"system": system,
		"stream": false,
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

// renderTerminalMarkdown dynamically swaps out Markdown bold characters with functional ANSI styles
func renderTerminalMarkdown(input string) string {
	// Replaces **text** with ANSI Intense Bold (\033[1m) and resets (\033[0m)
	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	output := reBold.ReplaceAllString(input, "\033[1m$1\033[0m")

	// Clean up any loose hashes (#) left behind by model headers
	reHeaders := regexp.MustCompile(`(?m)^#+\s+(.*)$`)
	output = reHeaders.ReplaceAllString(output, "\033[1m\033[4m$1\033[0m") // Underline + Bold headers

	return output
}
