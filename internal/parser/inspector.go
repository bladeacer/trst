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
		FSProperties: map[string]string{"Path": fp},
	}

	// Step 1: Deep container query via ffprobe
	parseFFProbe(fp, &track)

	// Step 2: Native tag library fallback
	if track.Title == "" {
		parseNativeTags(fp, &track)
	}

	// Step 3: Structural filename fallback
	if track.Title == "" {
		parseFilenameFallback(fp, ext, &track)
	}

	// Step 4: Lyrics retrieval
	track.Lyrics = readLyrics(fp, ext)

	if track.Genre == "" {
		track.Genre = "Unknown"
	}

	return track, true
}

// --- EXTRACTED COMPONENT HELPERS ---

func parseFFProbe(fp string, track *models.Track) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", fp)
	out, err := cmd.Output()
	if err != nil {
		return
	}

	var probe FFProbeResponse
	if err := json.Unmarshal(out, &probe); err != nil {
		return
	}

	if len(probe.Streams) > 0 {
		stream := probe.Streams[0]
		track.FSProperties["Codec"] = stream.CodecName
		if stream.SampleRate != "" {
			track.FSProperties["Frequency"] = stream.SampleRate + " Hz"
		}
	}

	if probe.Format.BitRate != "" {
		if br, err := strconv.Atoi(probe.Format.BitRate); err == nil && br > 0 {
			track.FSProperties["Bitrate"] = fmt.Sprintf("%d kbps", br/1000)
		}
	}

	if probe.Format.Tags == nil {
		return
	}

	tags := probe.Format.Tags
	track.Title = tags["title"]
	track.Artist = tags["artist"]
	track.Genre = tags["genre"]

	if desc, ok := tags["description"]; ok {
		track.Description = desc
	} else if comment, ok := tags["comment"]; ok {
		track.Description = comment
	}

	if bpmStr, ok := tags["bpm"]; ok {
		if bpm, err := strconv.Atoi(bpmStr); err == nil {
			track.BPM = bpm
		}
	}
}

func parseNativeTags(fp string, track *models.Track) {
	f, err := os.Open(fp)
	if err != nil {
		return
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return
	}

	track.Title = m.Title()
	track.Artist = m.Artist()
	if track.Genre == "" {
		track.Genre = m.Genre()
	}
	if comment := m.Comment(); comment != "" && track.Description == "" {
		track.Description = comment
	}
}

func parseFilenameFallback(fp, ext string, track *models.Track) {
	base := strings.TrimSuffix(filepath.Base(fp), ext)
	parts := strings.Split(base, " - ")

	if len(parts) >= 2 {
		track.Artist = strings.TrimSpace(parts[0])
		track.Title = strings.TrimSpace(parts[1])
		return
	}

	track.Title = base
	track.Artist = "Unknown Artist"
}

func readLyrics(fp, ext string) string {
	lyricPath := strings.TrimSuffix(fp, ext) + ".lrc"
	if _, err := os.Stat(lyricPath); err != nil {
		lyricPath = strings.TrimSuffix(fp, ext) + ".txt"
	}

	lyricBytes, err := os.ReadFile(lyricPath)
	if err != nil {
		return ""
	}

	return cleanLyrics(string(lyricBytes))
}
