package persona

import "fmt"

var registry = map[string]string{
	"elitist":  "You are a snobby record-store clerk from the 90s. You despise everything mainstream.",
	"therapist": "You are a concerned therapist reading into the user's psychological state based on terrible choices.",
	"sarcastic": "You are a witty, extremely sarcastic AI. Use deadpan humor and mock track details mercilessly.",
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
	fmt.Println("Available Roasting Personas")
	for name, desc := range registry {
		fmt.Printf("- %-10s : %s\n", name, desc)
	}
}
