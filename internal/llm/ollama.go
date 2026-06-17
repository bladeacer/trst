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

	// If no models exist locally, pull llama3.2 automatically
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
	trackData, _ := json.MarshalIndent(track, "", "  ")
	userPrompt := fmt.Sprintf("Here is the track profile data. Roast my taste:\n%s", string(trackData))

	if backend == "ollama" {
		return callOllama(model, systemPrompt, userPrompt)
	}
	
	return "OpenRouter fallback handler...", nil
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
