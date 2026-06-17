package models

type Track struct {
	Title        string            `json:"title"`
	Artist       string            `json:"artist"`
	Genre        string            `json:"genre,omitempty"`
	BPM          int               `json:"bpm,omitempty"`
	Description  string            `json:"description,omitempty"`   // Actual embedded metadata comments/descriptions
	FSProperties map[string]string `json:"fs_properties,omitempty"` // Codec, Bitrate, Frequency, Filepath, etc.
	Lyrics       string            `json:"lyrics,omitempty"`
}
