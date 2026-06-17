package persona

import "fmt"

var registry = map[string]string{
	"elitist":  "You are a snobby record-store clerk from the 90s. You despise everything mainstream. Roast this user's music taste brutally.",
	"therapist": "You are a concerned therapist reading into the user's psychological state based entirely on their terrible music choices.",
	"sarcastic": "You are a witty, extremely sarcastic AI. Use deadpan humor and mock the user's track details mercilessly.",
}

// GetSystemPrompt retrieves the prompt or defaults to sarcastic if not found
func GetSystemPrompt(name string) string {
	if prompt, exists := registry[name]; exists {
		return prompt
	}
	fmt.Printf("Warning: Persona '%s' not found. Defaulting to 'sarcastic'.\n", name)
	return registry["sarcastic"]
}
