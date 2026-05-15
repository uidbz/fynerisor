package main

import (
	"log"
	"os"

	"github.com/uidbz/fynerisor"
)

func main() {
	// Read script from file
	scriptBytes, err := os.ReadFile("main.risor")
	if err != nil {
		log.Fatalf("Failed to read script: %v", err)
	}

	// Create fynerisor window using NewApp convenience function
	window := fynerisor.NewApp("HTTP Fetch Example",
		fynerisor.WithHTTP(),
		fynerisor.WithStatusCallback(func(status string) {
			log.Printf("[STATUS] %s", status)
		}),
	)
	window.Resize(500, 400)

	// Load and execute the script
	window.LoadScript(string(scriptBytes))
	window.Execute()

	// Show and run
	window.ShowAndRun()
}
