package ui

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Spinner manages a pure ASCII background loading feedback thread
type Spinner struct {
	done chan bool
}

// NewSpinner initializes and kicks off the visual worker loop
func NewSpinner(persona string) *Spinner {
	fmt.Printf("[PERSONA]  %s\n\n", strings.ToTitle(persona))
	
	s := &Spinner{
		done: make(chan bool),
	}

	go func() {
		var seconds float64
		barLength := 20
		pos := 0
		for {
			select {
			case <-s.done:
				return
			default:
				// Construct a pure alnum / basic extended ASCII progress slider
				bar := make([]byte, barLength)
				for i := range bar {
					if i == pos {
						bar[i] = '='
					} else {
						bar[i] = ' '
					}
				}
				
				// Carriage return (\r) allows us to paint cleanly over the same terminal row
				fmt.Printf("\rProcessing: [%s] (%.1fs)", string(bar), seconds)
				os.Stdout.Sync()
				
				time.Sleep(100 * time.Millisecond)
				seconds += 0.1
				pos = (pos + 1) % barLength
			}
		}
	}()

	return s
}

// Stop terminates the background printer and sweeps the row clean
func (s *Spinner) Stop() {
	s.done <- true
	// Wipe out the loading line using simple trailing spaces
	fmt.Print("\r" + strings.Repeat(" ", 60) + "\r")
	os.Stdout.Sync()
}
