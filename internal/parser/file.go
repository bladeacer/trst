package parser

import (
	"os"
	"path/filepath"

	"github.com/bladeacer/trst/pkg/models"
)

var supportedExts = map[string]bool{
	".mp3": true, ".m4a": true, ".flac": true, ".ogg": true,
	".mp4": true, ".mkv": true, ".wav": true, ".opus": true,
}

// ParsePath crawls directories or handles single audio files
func ParsePath(path string) ([]models.Track, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var tracks []models.Track
	if !fi.IsDir() {
		if track, ok := parseFile(path); ok {
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
