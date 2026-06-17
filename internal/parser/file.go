package parser

import (
	"encoding/json"
	"fmt"
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
		CodecName  string `json:"codec_name"`
		SampleRate string `json:"sample_rate"`
		Channels   int    `json:"channels"`
	} `json:"streams"`
	Format struct {
		BitRate string            `json:"bit_rate"`
		Tags    map[string]string `json:"tags"`
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
		if track, ok := parseFile(path); ok { tracks = append(tracks, track) }
		return tracks, nil
	}

	err = filepath.Walk(path, func(fp string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !info.IsDir() {
			if track, ok := parseFile(fp); ok { tracks = append(tracks, track) }
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

	track := models.Track{}
	var qualityClues []string

	// 1. Run deep container query
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", fp)
	if out, err := cmd.Output(); err == nil {
		var probe FFProbeResponse
		if json.Unmarshal(out, &probe) == nil {
			if len(probe.Streams) > 0 {
				stream := probe.Streams[0]
				qualityClues = append(qualityClues, "Codec: "+stream.CodecName)
				if stream.SampleRate != "" {
					qualityClues = append(qualityClues, "Frequency: "+stream.SampleRate+" Hz")
				}
			}
			if probe.Format.BitRate != "" {
				var br int
				fmt.Sscanf(probe.Format.BitRate, "%d", &br)
				if br > 0 {
					qualityClues = append(qualityClues, fmt.Sprintf("Bitrate: %d kbps", br/1000))
				}
			}
			if probe.Format.Tags != nil {
				track.Title = probe.Format.Tags["title"]
				track.Artist = probe.Format.Tags["artist"]
				track.Genre = probe.Format.Tags["genre"]
				if bpmStr, ok := probe.Format.Tags["bpm"]; ok {
					fmt.Sscanf(bpmStr, "%d", &track.BPM)
				}
			}
		}
	}

	// 2. Tag library fallback
	if track.Title == "" {
		if f, err := os.Open(fp); err == nil {
			if m, err := tag.ReadFrom(f); err == nil {
				track.Title = m.Title()
				track.Artist = m.Artist()
				if track.Genre == "" {
					track.Genre = m.Genre()
				}
			}
			f.Close()
		}
	}

	// 3. Structural fallback naming conversions
	if track.Title == "" {
		base := strings.TrimSuffix(filepath.Base(fp), ext)
		if parts := strings.Split(base, " - "); len(parts) >= 2 {
			track.Artist = strings.TrimSpace(parts[0])
			track.Title = strings.TrimSpace(parts[1])
		} else {
			track.Title = base
			track.Artist = "Unknown Artist"
		}
	}

	// 4. Clean lyrics collection
	var lyricsText string
	lyricPath := strings.TrimSuffix(fp, ext) + ".lrc"
	if _, err := os.Stat(lyricPath); err != nil {
		lyricPath = strings.TrimSuffix(fp, ext) + ".txt"
	}
	if lyricBytes, err := os.ReadFile(lyricPath); err == nil {
		lyricsText = cleanLyrics(string(lyricBytes))
	}

	// Construct comprehensive metadata clues for downstream LLM processing
	qualityString := strings.Join(qualityClues, " | ")
	track.Description = "File Info: " + qualityString + " | Path: " + fp
	if lyricsText != "" {
		track.Description += " | Lyrics: " + lyricsText
	}

	// Set soft native fallbacks if they are completely unparsed before refiner calls
	if track.Genre == "" {
		track.Genre = "Unknown"
	}

	return track, true
}

func cleanLyrics(raw string) string {
	// Strip standard timestamp syntax patterns like [00:11.22] or [01:10:05] safely
	re := regexp.MustCompile(`\[\d+:\d+[\.\:]?\d*\]`)
	var lines []string
	
	for _, line := range strings.Split(raw, "\n") {
		cleanedLine := re.ReplaceAllString(line, "")
		trimmed := strings.TrimSpace(cleanedLine)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	if len(lines) > 30 {
		lines = lines[:30]
	}
	return strings.Join(lines, " / ")
}
