package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/bladeacer/trst/internal/llm"
	"github.com/bladeacer/trst/internal/parser"
	"github.com/bladeacer/trst/internal/persona"
)

type Globals struct {
	Persona string `help:"The roasting persona to use." default:"sarcastic" short:"p"`
	Backend string `help:"LLM provider (ollama or openrouter)." default:"ollama" short:"s"`
	Model   string `help:"Model name to target. Auto-detects if empty." default:"" short:"m"`
}

type LocalCmd struct {
	Path string `arg:"" help:"Path to local audio/video file or folder." type:"path"`
}

func (l *LocalCmd) Run(g *Globals) error {
	// 1. Parse tracks from path
	tracks, err := parser.ParsePath(l.Path)
	if err != nil {
		return fmt.Errorf("failed parsing path: %w", err)
	}
	if len(tracks) == 0 {
		return fmt.Errorf("no processable files found in: %s", l.Path)
	}

	// Roast the first found track
	track := tracks[0]

	// 2. Resolve Persona Prompt
	systemPrompt := persona.GetSystemPrompt(g.Persona)

	// 3. Resolve Model via Auto-detection if empty
	targetModel := g.Model
	if g.Backend == "ollama" && targetModel == "" {
		detectedModel, err := llm.AutoSelectOllamaModel()
		if err != nil {
			return fmt.Errorf("ollama error: %w", err)
		}
		targetModel = detectedModel
	} else if targetModel == "" {
		targetModel = "google/gemini-2.5-flash"
	}

	// 4. Generate the Roast
	fmt.Printf("\n[ROASTING] '%s' by '%s' using %s (%s)...\n\n", track.Title, track.Artist, g.Backend, targetModel)
	
	roast, err := llm.GenerateRoast(g.Backend, targetModel, systemPrompt, track)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Println(roast)
	return nil
}

type PlayerctlCmd struct{}
func (p *PlayerctlCmd) Run(g *Globals) error { return nil }

type YoutubeCmd struct{ URL string `arg:""` }
func (y *YoutubeCmd) Run(g *Globals) error { return nil }

var CLI struct {
	Globals
	Local     LocalCmd     `cmd:"" help:"Roast music from local directory metadata."`
	Playerctl PlayerctlCmd `cmd:"" help:"Roast the song currently playing on your system."`
	Youtube   YoutubeCmd   `cmd:"" help:"Roast a YouTube / YT Music playlist or video."`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name("trst"),
		kong.Description("A local CLI LLM that roasts your music taste."),
		kong.UsageOnError(),
	)
	if err := ctx.Run(&CLI.Globals); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
