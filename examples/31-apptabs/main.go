package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	// Create fynerisor window
	fyneWindow := gui.NewApp("AppTabs Example",
		gui.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
	)

	// Load script
	script, err := os.ReadFile("apptabs.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show window and run the application
	fyneWindow.ShowAndRun()
}
