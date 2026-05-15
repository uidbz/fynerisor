package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := fynerisor.NewApp("Button Callback Example",
		fynerisor.WithStatusCallback(func(status string) {
			fmt.Printf("Status: %s\n", status)
		}),
		fynerisor.WithResultCallback(func(result string) {
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
