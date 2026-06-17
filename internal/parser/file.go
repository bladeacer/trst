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

type FFProbeResponse struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
		Duration  string `json:"duration"`
	} `json:"streams"`
	Format struct {
		Tags map[string]string `json:"tags"`
	} `json:"format"`
}

var supportedExts = map[string]bool{
	".mp3": true, ".m4a": true, ".flac": true, ".ogg": true, 
	".mp4": true, ".mkv": true, ".wav": true, ".opus": true,
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
		Source:      "local",
		Description: "",
	}

	// 1. Read metadata via ffprobe
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", fp)
	if out, err := cmd.Output(); err == nil {
		var probe FFProbeResponse
		if json.Unmarshal(out, &probe) == nil {
			if probe.Format.Tags != nil {
				track.Title = probe.Format.Tags["title"]
				track.Artist = probe.Format.Tags["artist"]
				track.Genre = probe.Format.Tags["genre"]
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
				if track.Genre == "" {
					track.Genre = m.Genre()
				}
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

	// 4. Look for an associated LRC/lyrics file
	lyricPath := strings.TrimSuffix(fp, ext) + ".lrc"
	if _, err := os.Stat(lyricPath); err != nil {
		// Try .txt variant just in case
		lyricPath = strings.TrimSuffix(fp, ext) + ".txt"
	}

	if lyricBytes, err := os.ReadFile(lyricPath); err == nil {
		track.Description = cleanLyrics(string(lyricBytes))
	}

	return track, true
}

// cleanLyrics strips timestamps like [00:12.34] out of standard sync profiles
func cleanLyrics(raw string) string {
	re := regexp.MustCompile(`\[\d+:\d+[\.\:]\d+\]`)
	cleaned := re.ReplaceAllString(raw, "")
	
	// Group lines and trim whitespace
	lines := strings.Split(cleaned, "\n")
	var finalLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			finalLines = append(finalLines, trimmed)
		}
	}
	
	// Return up to the first 40 lines so we don't blow past context bounds
	if len(finalLines) > 40 {
		finalLines = finalLines[:40]
	}
	return strings.Join(finalLines, " / ")
}
