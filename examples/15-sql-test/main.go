package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
	guidb "github.com/uidbz/fynerisor/gui/db"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := gui.NewApp("SQL Module Test",
		gui.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
		guidb.WithSQL(), // Enable SQL module (opt-in subpackage, issue #40)
	)

	// Load Risor script from file
	script, err := os.ReadFile("app.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show window and run the application
	fyneWindow.ShowAndRun()
}
