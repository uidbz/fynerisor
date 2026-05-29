package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := gui.NewApp("Button Callback Example",
		gui.WithStatusCallback(func(status string) {
			fmt.Printf("Status: %s\n", status)
		}),
		gui.WithResultCallback(func(result string) {
			fmt.Printf("Result: %s\n", result)
		}),
	)

	// Load Risor script from file
	script, err := os.ReadFile("button.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show window and run the application
	fyneWindow.ShowAndRun()
}
