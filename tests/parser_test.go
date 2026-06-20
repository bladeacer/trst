package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bladeacer/trst/internal/parser"
)

func TestParsePath_MetadataFallbacks(t *testing.T) {
	// 1. Create a temporary directory isolated to this test
	tmpDir := t.TempDir()

	// 2. Set up a mock audio file path (using a supported extension)
	// We'll name it using the "Artist - Title" pattern to test your structural fallback
	mockAudioPath := filepath.Join(tmpDir, "Laserhawk - Grid Racer.mp3")

	// Create the empty mock audio file
	err := os.WriteFile(mockAudioPath, []byte("mock audio data"), 0644)
	if err != nil {
		t.Fatalf("failed to create mock audio file: %v", err)
	}

	// 3. Create a matching mock .txt lyrics file to test the lyrics parser
	mockLyricsPath := filepath.Join(tmpDir, "Laserhawk - Grid Racer.txt")
	mockLyricsContent := "[00:11.22]Neon lights flashing\n[00:15.00]Grid racer"

	err = os.WriteFile(mockLyricsPath, []byte(mockLyricsContent), 0644)
	if err != nil {
		t.Fatalf("failed to create mock lyrics file: %v", err)
	}

	// 4. Run the actual parser package code against your mock environment
	tracks, err := parser.ParsePath(mockAudioPath)
	if err != nil {
		t.Fatalf("ParsePath failed: %v", err)
	}

	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}

	gotTrack := tracks[0]

	// 5. Assert the structural fallback parsed the filename correctly
	if gotTrack.Artist != "Laserhawk" {
		t.Errorf("expected Artist 'Laserhawk', got '%s'", gotTrack.Artist)
	}
	if gotTrack.Title != "Grid Racer" {
		t.Errorf("expected Title 'Grid Racer', got '%s'", gotTrack.Title)
	}

	// 6. Assert the lyrics were found, cleaned of timestamps, and joined by " / "
	expectedLyrics := "Neon lights flashing / Grid racer"
	if gotTrack.Lyrics != expectedLyrics {
		t.Errorf("expected Lyrics '%s', got '%s'", expectedLyrics, gotTrack.Lyrics)
	}

	// 7. Assert default genre fallback works
	if gotTrack.Genre != "Unknown" {
		t.Errorf("expected default Genre 'Unknown', got '%s'", gotTrack.Genre)
	}
}
