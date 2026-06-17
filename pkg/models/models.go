package models

// Track represents the normalized music metadata passed to the LLM
type Track struct {
	Title       string   `json:"title"`
	Artist      string   `json:"artist"`
	Genre       string   `json:"genre,omitempty"`
	BPM         int      `json:"bpm,omitempty"`
	DateAllowed string   `json:"date_released,omitempty"`
	Description string   `json:"description,omitempty"`
}
