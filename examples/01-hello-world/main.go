package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := gui.NewApp("Hello World Example",
		gui.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
	)

	// Load Risor script from file
	script, err := os.ReadFile("hello.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show window and run the application
	fyneWindow.ShowAndRun()
}
