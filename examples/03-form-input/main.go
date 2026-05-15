package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := fynerisor.NewApp("Form Input Example",
		fynerisor.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
	)

	// Load Risor script from file
	script, err := os.ReadFile("form.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show window and run the application
	fyneWindow.ShowAndRun()
}
