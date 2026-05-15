package main

import (
	"log"
	"os"

	"github.com/uidbz/fynerisor"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := fynerisor.NewApp("Calendar Example",
		fynerisor.WithTime(),
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
