package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	// Create fynerisor window with os and strings modules enabled
	fyneWindow := gui.NewApp("Image Gallery",
		gui.WithOS(),       // Enable os module for file operations
		gui.WithStrings(),  // Enable strings module for string operations
		gui.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
	)

	// Load script
	script, err := os.ReadFile("image-gallery.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show window and run the application
	fyneWindow.ShowAndRun()
}
