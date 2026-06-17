package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/bladeacer/trst/internal/llm"
	"github.com/bladeacer/trst/internal/parser"
	"github.com/bladeacer/trst/internal/persona"
)

var CLI struct {
	Path    string `arg:"" help:"Path to local audio/video file or folder." type:"path"`
	Persona string `help:"The roasting persona to use." default:"sarcastic" short:"p"`
	Backend string `help:"LLM provider (ollama or openrouter)." default:"ollama" short:"s"`
	Model   string `help:"Model name to target. Auto-detects if empty." default:"" short:"m"`
}

func main() {
	_ = kong.Parse(&CLI,
		kong.Name("trst"),
		kong.Description("A local CLI LLM that roasts your music taste directly from files."),
		kong.UsageOnError(),
	)

	// 1. Parse track metadata
	tracks, err := parser.ParsePath(CLI.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing path: %v\n", err)
		os.Exit(1)
	}
	if len(tracks) == 0 {
		fmt.Fprintf(os.Stderr, "error: no processable files found in %s\n", CLI.Path)
		os.Exit(1)
	}

	track := tracks[0]

	// 2. Resolve Persona Prompt
	systemPrompt := persona.GetSystemPrompt(CLI.Persona)

	// 3. Handle Auto-detection or pulling fallback model
	targetModel := CLI.Model
	if CLI.Backend == "ollama" && targetModel == "" {
		detectedModel, err := llm.AutoSelectOllamaModel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ollama error: %v\n", err)
			os.Exit(1)
		}
		targetModel = detectedModel
	} else if targetModel == "" {
		targetModel = "google/gemini-2.5-flash"
	}

	// 4. Generate the Roast
	fmt.Printf("\n[ROASTING] '%s' by '%s' [%s | %d BPM] using %s (%s)...\n\n", 
		track.Title, track.Artist, track.Genre, track.BPM, CLI.Backend, targetModel)
	
	roast, err := llm.GenerateRoast(CLI.Backend, targetModel, systemPrompt, track)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(roast)
}
