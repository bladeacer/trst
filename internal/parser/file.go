package parser

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bladeacer/trst/pkg/models"
	"github.com/dhowden/tag"
)

var supportedExts = map[string]bool{
	".mp3": true, ".m4a": true, ".flac": true, ".ogg": true, 
	".mp4": true, ".mkv": true, ".wav": true, ".opus": true,
}

type FFProbeResponse struct {
	Format struct {
		Tags map[string]string `json:"tags"`
	} `json:"format"`
}

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

	track := models.Track{
		// Leaving explicit 'Source' field out of the equation entirely
		Description: "",
	}

	var rawMetadataGenre string

	// 1. Read metadata via ffprobe
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", fp)
	if out, err := cmd.Output(); err == nil {
		var probe FFProbeResponse
		if json.Unmarshal(out, &probe) == nil {
			if probe.Format.Tags != nil {
				track.Title = probe.Format.Tags["title"]
				track.Artist = probe.Format.Tags["artist"]
				rawMetadataGenre = probe.Format.Tags["genre"]
			}
		}
	}

	// 2. Tag library fallback
	if track.Title == "" {
		f, err := os.Open(fp)
		if err == nil {
			m, err := tag.ReadFrom(f)
			if err == nil {
				track.Title = m.Title()
				track.Artist = m.Artist()
				rawMetadataGenre = m.Genre()
			}
			f.Close()
		}
	}

	// 3. Filename splitting fallback
	if track.Title == "" {
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

	// 4. Extract lyrics if present
	var lyricsText string
	lyricPath := strings.TrimSuffix(fp, ext) + ".lrc"
	if _, err := os.Stat(lyricPath); err != nil {
		lyricPath = strings.TrimSuffix(fp, ext) + ".txt"
	}
	if lyricBytes, err := os.ReadFile(lyricPath); err == nil {
		lyricsText = cleanLyrics(string(lyricBytes))
	}

	// Semantic compilation: Combine metadata clues, path layouts, and lyrics into a profile description
	track.Description = CompileSemanticProfile(fp, track.Title, track.Artist, rawMetadataGenre, lyricsText)

	// Local hard fallbacks for execution safety before LLM pipeline handles it
	track.Genre = InferGenreFromProfile(track.Description)
	track.BPM = InferBPMFromProfile(track.Title, track.Genre)

	return track, true
}

func CompileSemanticProfile(fp, title, artist, metaGenre, lyrics string) string {
	var parts []string
	if artist != "" { parts = append(parts, "Artist Context: "+artist) }
	if title != "" { parts = append(parts, "Title Context: "+title) }
	if metaGenre != "" && !strings.EqualFold(metaGenre, "Unknown") { parts = append(parts, "Raw Clue: "+metaGenre) }
	
	// Add structural filepath directory clues (e.g., /music/Heavy-Metal/track.mp3)
	dir := filepath.Base(filepath.Dir(fp))
	if dir != "." && dir != "" {
		parts = append(parts, "Folder Context: "+dir)
	}
	if lyrics != "" { parts = append(parts, "Lyrical Content: "+lyrics) }

	return strings.Join(parts, " | ")
}

func InferGenreFromProfile(description string) string {
	searchString := strings.ToLower(description)
	genreMatrix := map[string][]string{
		"Classical":  {"orchestral", "zimmer", "symphony", "piano", "sonata", "classical", "opera", "ost"},
		"Metal":      {"metal", "core", "death", "thrash", "slayer", "riff", "djent"},
		"Synthwave":  {"synthwave", "retrowave", "outrun", "cyberpunk", "neon"},
		"Lofi":       {"lofi", "chillhop", "study", "relaxing", "bedroom"},
		"Techno/EDM":{"techno", "house", "edm", "dance", "remix", "club", "trance"},
		"Hip-Hop":    {"rap", "hiphop", "trap", "beats", "freestyle"},
		"Rock":       {"rock", "grunge", "punk", "indie rock", "guitar"},
		"Pop":        {"pop", "hits", "top40", "radio"},
	}

	for genre, keywords := range genreMatrix {
		for _, kw := range keywords {
			if strings.Contains(searchString, kw) {
				return genre
			}
		}
	}
	return "Unclassifiable Noise"
}

func InferBPMFromProfile(title, genre string) int {
	titleLower := strings.ToLower(title)
	if strings.Contains(titleLower, "speed up") || strings.Contains(titleLower, "nightcore") { return 165 }
	if strings.Contains(titleLower, "slowed") { return 68 }

	switch genre {
	case "Classical": return 80
	case "Lofi":      return 72
	case "Hip-Hop":   return 92
	case "Rock":      return 115
	case "Synthwave": return 118
	case "Techno/EDM": return 128
	case "Metal":     return 145
	case "Pop":       return 120
	default:          return 100
	}
}

func cleanLyrics(raw string) string {
	re := regexp.MustCompile(`\[\d+:\d+[\.\:]\d+\]`)
	cleaned := re.ReplaceAllString(raw, "")
	lines := strings.Split(cleaned, "\n")
	var finalLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" { finalLines = append(finalLines, trimmed) }
	}
	if len(finalLines) > 30 { finalLines = finalLines[:30] }
	return strings.Join(finalLines, " / ")
}
