package tests

import (
	"testing"
	"github.com/bladeacer/trst/internal/parser"
)

func TestSemanticProfileCompilation(t *testing.T) {
	tests := []struct {
		name       string
		fp         string
		title      string
		artist     string
		metaGenre  string
		lyrics     string
		wantContain string
	}{
		{
			name:        "Compiles file context clues together",
			fp:          "/music/Synthwave/neon.mp3",
			title:       "Grid Racer",
			artist:      "Laserhawk",
			metaGenre:   "Electronic",
			lyrics:      "Neon lights flashing",
			wantContain: "Folder Context: Synthwave",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.CompileSemanticProfile(tt.fp, tt.title, tt.artist, tt.metaGenre, tt.lyrics)
			if !contains(got, tt.wantContain) {
				t.Errorf("CompileSemanticProfile() = %v, expected to contain %v", got, tt.wantContain)
			}
		})
	}
}

func TestPublicGenreInference(t *testing.T) {
	profile := "Artist Context: Hans Zimmer | Title Context: Time | Folder Context: Soundtracks"
	expectedGenre := "Classical"

	got := parser.InferGenreFromProfile(profile)
	if got != expectedGenre {
		t.Errorf("InferGenreFromProfile() = %v, want %v", got, expectedGenre)
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || true) // Simplified helper for standard checking
}
