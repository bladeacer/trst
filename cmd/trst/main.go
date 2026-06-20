package main

import (
	"os"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Path           string `arg:"" help:"Path to local audio/video file or folder." type:"path" optional:""`
	Persona        string `help:"The roasting persona to use." default:"sarcastic" short:"p"`
	Backend        string `help:"LLM provider (ollama or openrouter)." default:"ollama" short:"s"`
	Model          string `help:"Model name to target. Auto-detects if empty." default:"" short:"m"`
	List           bool   `help:"List all available roasting personas and exit." short:"l"`
	Jerk           int    `help:"Scale of how much of a jerk the persona is (1-5)." default:"3" short:"j"`
	AllowProfanity bool   `help:"Allow the use of profanities in the roast." default:"false"`

	// Cache Management Administrative Flags
	ClearCache   bool   `help:"Purge and clear all cached roasts inside the database." name:"clear-cache" short:"c"`
	DeleteCached string `help:"Evict a single specific entry using format 'Title::Artist'." name:"delete-cached" type:"string" short:"d"`
	ViewCached   bool   `help:"Inspect and list all active keys and metadata currently stored in the cache." name:"view-cached" short:"v"`
}

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "--help")
	}

	_ = kong.Parse(&CLI,
		kong.Name("trst"),
		kong.Description("CLI for local LLMs to roast your music taste."),
		kong.UsageOnError(),
	)

	// Map the Kong CLI struct fields directly into the AppConfig type expected by Execute()
	config := AppConfig{
		Path:           CLI.Path,
		Persona:        CLI.Persona,
		Backend:        CLI.Backend,
		Model:          CLI.Model,
		List:           CLI.List,
		Jerk:           CLI.Jerk,
		AllowProfanity: CLI.AllowProfanity,

		// Wire up the database control assignments
		ClearCache:   CLI.ClearCache,
		DeleteCached: CLI.DeleteCached,
		ViewCached:   CLI.ViewCached,
	}

	// Pass the address of our newly populated config
	Execute(&config)
}
