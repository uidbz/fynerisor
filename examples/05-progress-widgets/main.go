package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := gui.NewApp("Progress & Slider Widgets",
		gui.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
	)

	// Load Risor script from file
	script, err := os.ReadFile("progress.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show window and run the application
	fyneWindow.ShowAndRun()
}
