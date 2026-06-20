package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bladeacer/trst/pkg/models"
	"github.com/dhowden/tag"
)

type FFProbeResponse struct {
	Streams []struct {
		CodecName  string `json:"codec_name"`
		SampleRate string `json:"sample_rate"`
	} `json:"streams"`
	Format struct {
		BitRate string            `json:"bit_rate"`
		Tags    map[string]string `json:"tags"`
	} `json:"format"`
}

func parseFile(fp string) (models.Track, bool) {
	ext := strings.ToLower(filepath.Ext(fp))
	if !supportedExts[ext] {
		return models.Track{}, false
	}

	track := models.Track{
		FSProperties: make(map[string]string),
	}
	track.FSProperties["Path"] = fp

	// 1. Run deep container query
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", fp)
	if out, err := cmd.Output(); err == nil {
		var probe FFProbeResponse
		if json.Unmarshal(out, &probe) == nil {
			if len(probe.Streams) > 0 {
				stream := probe.Streams[0]
				track.FSProperties["Codec"] = stream.CodecName
				if stream.SampleRate != "" {
					track.FSProperties["Frequency"] = stream.SampleRate + " Hz"
				}
			}
			if probe.Format.BitRate != "" {
				// strconv.Atoi is faster and more idiomatic for simple strings to ints
				if br, err := strconv.Atoi(probe.Format.BitRate); err == nil && br > 0 {
					track.FSProperties["Bitrate"] = fmt.Sprintf("%d kbps", br/1000)
				}
			}
			if probe.Format.Tags != nil {
				track.Title = probe.Format.Tags["title"]
				track.Artist = probe.Format.Tags["artist"]
				track.Genre = probe.Format.Tags["genre"]

				// Map actual file metadata comments/descriptions specifically to Description
				if desc, ok := probe.Format.Tags["description"]; ok {
					track.Description = desc
				} else if comment, ok := probe.Format.Tags["comment"]; ok {
					track.Description = comment
				}

				if bpmStr, ok := probe.Format.Tags["bpm"]; ok {
					if bpm, err := strconv.Atoi(bpmStr); err == nil {
						track.BPM = bpm
					}
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
				if comment := m.Comment(); comment != "" && track.Description == "" {
					track.Description = comment
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

	// 4. Lyrics retrieval
	var lyricsText string
	lyricPath := strings.TrimSuffix(fp, ext) + ".lrc"
	if _, err := os.Stat(lyricPath); err != nil {
		lyricPath = strings.TrimSuffix(fp, ext) + ".txt"
	}
	if lyricBytes, err := os.ReadFile(lyricPath); err == nil {
		lyricsText = cleanLyrics(string(lyricBytes))
	}

	track.Lyrics = lyricsText

	if track.Genre == "" {
		track.Genre = "Unknown"
	}

	return track, true
}
