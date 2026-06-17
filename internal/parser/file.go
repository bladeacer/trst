package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bladeacer/trst/pkg/models"
	"github.com/dhowden/tag"
)

var supportedExts = map[string]bool{
	".mp3": true, ".m4a": true, ".flac": true, ".ogg": true, 
	".mp4": true, ".mkv": true, ".wav": true, ".opus": true,
}

// ParsePath reads a file or folder and produces uniform Track models
func ParsePath(path string) ([]models.Track, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var tracks []models.Track

	if !fi.IsDir() {
		track, ok := parseFile(path)
		if ok {
			tracks = append(tracks, track)
		}
		return tracks, nil
	}

	err = filepath.Walk(path, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if track, ok := parseFile(fp); ok {
				tracks = append(tracks, track)
			}
		}
		return nil
	})

	return tracks, err
}

func parseFile(fp string) (models.Track, bool) {
	ext := strings.ToLower(filepath.Ext(fp))
	if !supportedExts[ext] {
		return models.Track{}, false
	}

	f, err := os.Open(fp)
	if err != nil {
		return models.Track{}, false
	}
	defer f.Close()

	track := models.Track{
		Source:      "local",
		Description: filepath.Base(fp),
	}

	// Try reading embedded tags
	m, err := tag.ReadFrom(f)
	if err == nil {
		track.Title = m.Title()
		track.Artist = m.Artist()
		track.Genre = m.Genre()
	}

	// Fallbacks if metadata tags are blank
	if track.Title == "" {
		// e.g., "01. Artist - Song Title.mp3" -> Title: "Song Title", Artist: "Artist"
		base := strings.TrimSuffix(filepath.Base(fp), ext)
		parts := strings.Split(base, " - ")
		if len(parts) >= 2 {
			track.Artist = strings.TrimSpace(parts[0])
			track.Title = strings.TrimSpace(parts[1])
		} else {
			track.Title = base
			track.Artist = "Unknown Artist"
		}
	}

	// Infer configurations based on names or paths
	track.Genre = inferGenre(fp, track.Genre)
	track.BPM = inferBPM(track.Title, track.Genre)

	return track, true
}

func inferGenre(fp, existingGenre string) string {
	if existingGenre != "" {
		return existingGenre
	}
	// Fallback to directory names (e.g. music/Synthwave/track.mp3)
	dir := filepath.Base(filepath.Dir(fp))
	lowerDir := strings.ToLower(dir)
	for _, g := range []string{"rock", "pop", "rap", "hiphop", "techno", "lofi", "jazz", "synthwave", "metal"} {
		if strings.Contains(lowerDir, g) {
			return g
		}
	}
	return "Mystery Sound"
}

func inferBPM(title, genre string) int {
	// A fun heuristics fallback generator since native file tags rarely contain raw structural BPM
	titleLower := strings.ToLower(title)
	if strings.Contains(titleLower, "remix") || strings.Contains(titleLower, "club") {
		return 128
	}
	switch strings.ToLower(genre) {
	case "lofi", "jazz":
		return 75
	case "techno", "synthwave":
		return 125
	case "rap", "hiphop":
		return 90
	default:
		return 100 // Safe mid-tempo default
	}
}
