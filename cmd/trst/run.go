package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bladeacer/trst/internal/cache"
	"github.com/bladeacer/trst/internal/llm"
	"github.com/bladeacer/trst/internal/parser"
	"github.com/bladeacer/trst/internal/persona"
	"github.com/bladeacer/trst/internal/ui" // Import your ui engine package safely
	"github.com/bladeacer/trst/pkg/models"
)

type AppConfig struct {
	Path           string
	Persona        string
	Backend        string
	Model          string
	List           bool
	Jerk           int
	AllowProfanity bool

	ClearCache   bool
	DeleteCached string 
	ViewCached   bool
	DisableCache bool   
}

func Execute(cfg *AppConfig) {
	store := initCacheStore(cfg.DisableCache)
	defer store.Close()

	if interceptAdminCommands(cfg, store) {
		return
	}

	track := parseInputTrack(cfg.Path)
	targetModel := resolveTargetModel(cfg.Model)

	// Step 1: Auto-select local engine parameters
	if cfg.Backend == "ollama" && targetModel == "" {
		localModel, err := llm.AutoSelectOllamaModel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ollama connection error: %v\n", err)
			fmt.Fprintln(os.Stderr, "Ensure Ollama is running locally ('ollama serve').")
			os.Exit(1)
		}
		targetModel = localModel
	}

	if targetModel == "" {
		fmt.Fprintln(os.Stderr, "error: no LLM model specified or detected.")
		os.Exit(1)
	}

	// Step 2: Check database for structural metadata cache hits
	trackKey := fmt.Sprintf("%s::%s", track.Title, track.Artist)
	cachedMeta, found := store.GetTrackMeta(trackKey)

	if found {
		// Cache Hit! Apply stored metrics immediately
		track.Genre = cachedMeta.Genre
		track.BPM = cachedMeta.BPM
	} else {
		// Cache Miss. Boot up the spinner and run musicology refinement via LLM
		spinner := ui.NewSpinner(cfg.Persona)
		llm.RefineTrackDetails(cfg.Backend, targetModel, &track)
		spinner.Stop()

		// Save the newly discovered values so we never have to run refinement on this track again
		_ = store.SetTrackMeta(trackKey, track.Genre, track.BPM)
	}

	// Step 3: Always print the execution header with up-to-date metrics
	printExecutionHeader(track, cfg, targetModel) 

	// Step 4: Run the live generation pipeline using the specified persona config parameters
	spinner := ui.NewSpinner(cfg.Persona)
	roast, err := llm.GenerateRoast(cfg.Backend, targetModel, cfg.Persona, track, cfg.Jerk, cfg.AllowProfanity)
	spinner.Stop()

	if err != nil {
		fmt.Fprintf(os.Stderr, "generation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(roast)
}

// --- Isolated Business Logic Helper Engines ---

func initCacheStore(disabled bool) cache.Service {
	if disabled {
		return &cache.NopCache{}
	}

	dbPath, err := cache.GetDatabasePath()
	if err != nil {
		return &cache.NopCache{} 
	}

	storage, err := cache.NewTrackCache(dbPath)
	if err != nil {
		return &cache.NopCache{}
	}
	return storage
}

func interceptAdminCommands(cfg *AppConfig, store cache.Service) bool {
	if cfg.ClearCache {
		if err := store.ClearAll(); err != nil {
			fmt.Fprintf(os.Stderr, "error flushing cache db: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully purged all cached entries from the database.")
		return true
	}

	if cfg.DeleteCached != "" {
		if err := store.DeleteEntry(cfg.DeleteCached); err != nil {
			fmt.Fprintf(os.Stderr, "error deleting target key: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Evicted key '%s' from local database state safely.\n", cfg.DeleteCached)
		return true
	}

	if cfg.ViewCached {
		entries, err := store.ListAllEntries()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to inspect keys: %v\n", err)
			os.Exit(1)
		}
		if len(entries) == 0 {
			fmt.Println("Cache is completely empty.")
			return true
		}

		fmt.Println("--- CURRENT CACHED ROASTS ---")
		for _, entry := range entries {
			// Convert the saved timestamp into a readable date format
			savedTime := time.Unix(entry.Meta.CreatedAt, 0).Format("2006-01-02 15:04")

			// Print structured track details inline
			fmt.Printf("- %-40s | Genre: %-12s | BPM: %-3d | Saved: %s\n", 
			entry.TrackKey, 
			entry.Meta.Genre, 
			entry.Meta.BPM, 
			savedTime,
		)
	}
	return true
}

if cfg.List {
		persona.ListPersonas()
		return true
	}

	return false
}

func parseInputTrack(path string) models.Track {
	if path == "" {
		fmt.Fprintln(os.Stderr, "error: positional argument <path> is required unless operating management commands")
		os.Exit(1)
	}

	tracks, err := parser.ParsePath(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing path: %v\n", err)
		os.Exit(1)
	}
	if len(tracks) == 0 {
		fmt.Fprintf(os.Stderr, "error: no processable files found\n")
		os.Exit(1)
	}
	return tracks[0]
}

func printExecutionHeader(track models.Track, cfg *AppConfig, model string) {
	displayPersona := cfg.Persona
	if len(cfg.Persona) > 0 {
		displayPersona = strings.ToUpper(string(cfg.Persona[0])) + cfg.Persona[1:]
	}

	fmt.Printf("\n[ROASTING] '%s' by '%s' [%s | %d BPM]\n", track.Title, track.Artist, track.Genre, track.BPM)
	fmt.Printf("[PERSONA]  %s (Jerk Level: %d/5 | Backend: %s via %s)\n\n", displayPersona, cfg.Jerk, cfg.Backend, model)
}

func resolveTargetModel(modelFlag string) string {
	if modelFlag != "" {
		return modelFlag
	}
	return "" 
}
