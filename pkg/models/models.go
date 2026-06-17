package models

type Track struct {
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Genre       string `json:"genre,omitempty"`
	BPM         int    `json:"bpm,omitempty"`
	Description string `json:"description,omitempty"`
	Lyrics      string `json:"lyrics,omitempty"`
}
