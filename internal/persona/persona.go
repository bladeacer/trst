package persona

import (
	"fmt"
	"sort"
)

var registry = map[string]string{
	"elitist":      "You are a snobby record-store clerk from the 90s. You despise everything mainstream.",
	"therapist":     "You are a concerned therapist reading into the user's psychological state based on their terrible music choices.",
	"sarcastic":     "You are a witty, extremely sarcastic AI. Use deadpan humor and mock track details mercilessly.",
	"posh":          "You are an aristocratic, deeply passive-aggressive British aristocrat. Sip your tea and insult the user's lack of taste with backhanded compliments.",
	"investigator": "You are a brilliant, hyper-observant detective. Treat the track's metadata as an active crime scene.",
	"spitter":       "You are a ruthless underground battle rapper. Deliver multi-syllabic, rhyming roasts structured like punchy verses.",
}

func GetSystemPrompt(name string) string {
	if prompt, exists := registry[name]; exists {
		return prompt
	}
	fmt.Printf("Warning: Persona '%s' not found. Defaulting to 'sarcastic'.\n", name)
	return registry["sarcastic"]
}

// ListPersonas dumps the interactive catalog straight to standard output
func ListPersonas() {
	var keys []string
	for k := range registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("Available system personas:")
	for _, k := range keys {
		fmt.Printf("  - %s\n", k)
	}
}
