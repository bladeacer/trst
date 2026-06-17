package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bladeacer/trst/pkg/models"
	"github.com/dhowden/tag"
)

var supportedExts = map[string]bool{
	".mp3": true, ".m4a": true, ".flac": true, ".ogg": true, 
	".mp4": true, ".mkv": true, ".wav": true, ".opus": true,
}

type FFProbeResponse struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
		Duration  string `json:"duration"`
	} `json:"streams"`
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
		Source:      "local",
		Description: filepath.Base(fp),
	}

	// 1. Structural parse attempt using container profiles
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", fp)
	if out, err := cmd.Output(); err == nil {
		var probe FFProbeResponse
		if json.Unmarshal(out, &probe) == nil {
			if probe.Format.Tags != nil {
				track.Title = probe.Format.Tags["title"]
				track.Artist = probe.Format.Tags["artist"]
				track.Genre = probe.Format.Tags["genre"]
				
				if bpmStr, ok := probe.Format.Tags["bpm"]; ok {
					var b int
					fmt.Sscanf(bpmStr, "%d", &b)
					track.BPM = b
				}
			}
		}
	}

	// 2. Fallback to basic tag extraction library
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

	// 3. Fallback string mapping based on naming formats
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

	track.Genre = inferGenre(fp, track.Title, track.Artist, track.Genre)
	if track.BPM == 0 {
		track.BPM = inferBPM(track.Title, track.Genre)
	}

	return track, true
}

func inferGenre(fp, title, artist, existing string) string {
	if existing != "" && !strings.EqualFold(existing, "Unknown") {
		return existing
	}

	searchString := strings.ToLower(fp + " " + title + " " + artist)

	genreMatrix := map[string][]string{
		"Classical":  {"orchestral", "zimmer", "symphony", "piano", "sonata", "classical", "opera", "ost"},
		"Metal":      {"metal", "core", "death", "thrash", "slayer", "riff", "djent"},
		"Synthwave":  {"synthwave", "retrowave", "outrun", "cyberpunk", "1984", "neon"},
		"Lofi":       {"lofi", "chillhop", "study", "relaxing", "bedroom"},
		"Techno/EDM":{"techno", "house", "edm", "dance", "remix", "club", "trance", "dubstep"},
		"Hip-Hop":    {"rap", "hiphop", "trap", "beats", "freestyle", "underground rap"},
		"Rock":       {"rock", "grunge", "punk", "indie rock", "guitar", "psychedelic"},
		"Pop":        {"pop", "hits", "top40", "radio", "commercial"},
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

func inferBPM(title, genre string) int {
	titleLower := strings.ToLower(title)
	
	if strings.Contains(titleLower, "speed up") || strings.Contains(titleLower, "nightcore") {
		return 165
	}
	if strings.Contains(titleLower, "slowed") || strings.Contains(titleLower, "reverb") {
		return 68
	}

	switch genre {
	case "Classical":
		if strings.Contains(titleLower, "battle") || strings.Contains(titleLower, "chase") {
			return 140
		}
		return 80
	case "Lofi":
		return 72
	case "Hip-Hop":
		return 92
	case "Rock":
		return 115
	case "Synthwave":
		return 118
	case "Techno/EDM":
		return 128
	case "Metal":
		return 145
	case "Pop":
		return 120
	default:
		return 100
	}
}
