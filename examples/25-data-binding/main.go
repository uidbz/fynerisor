package main

import (
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	// Create fynerisor window
	w := gui.NewApp("Data Binding Example")

	// Load Risor script from file
	script, err := os.ReadFile("script.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	w.LoadScript(string(script))
	w.Execute()

	// Show window and run the application
	w.ShowAndRun()
}
