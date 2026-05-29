package main

import (
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := gui.NewApp("Calendar Example",
		gui.WithTime(),
	)
	fyneWindow.Resize(600, 500)

	// Load script
	script, err := os.ReadFile("app.risor")
	if err != nil {
		log.Fatal(err)
	}

	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show and run
	fyneWindow.ShowAndRun()
}
