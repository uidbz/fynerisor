package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := fynerisor.NewApp("SQL Module Test",
		fynerisor.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
		fynerisor.WithSQL(), // Enable SQL module
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
