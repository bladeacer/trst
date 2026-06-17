package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/bladeacer/trst/internal/llm"
	"github.com/bladeacer/trst/internal/parser"
	"github.com/bladeacer/trst/internal/persona"
	"github.com/bladeacer/trst/internal/ui"
)

var CLI struct {
	Path    string `arg:"" help:"Path to local audio/video file or folder." type:"path" optional:""`
	Persona string `help:"The roasting persona to use." default:"sarcastic" short:"p"`
	Backend string `help:"LLM provider (ollama or openrouter)." default:"ollama" short:"s"`
	Model   string `help:"Model name to target. Auto-detects if empty." default:"" short:"m"`
}

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "--help")
	}

	_ = kong.Parse(&CLI,
		kong.Name("trst"),
		kong.Description("A local CLI LLM that roasts your music taste directly from files."),
		kong.UsageOnError(),
	)

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
	systemPrompt := persona.GetSystemPrompt(CLI.Persona)

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

	fmt.Printf("\n[ROASTING] '%s' by '%s' [%s | %d BPM] using %s (%s)...\n", 
		track.Title, track.Artist, track.Genre, track.BPM, CLI.Backend, targetModel)
	
	// Start modular UI spinner thread
	spinner := ui.NewSpinner(CLI.Persona)

	roast, err := llm.GenerateRoast(CLI.Backend, targetModel, systemPrompt, track)
	
	// Safely kill loading indicators instantly when payload hits memory buffer
	spinner.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "generation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(roast)
}
