package main

import (
	"log"
	"os"

	"github.com/uidbz/fynerisor"
)

func main() {
	// Create fynerisor window using NewApp convenience function
	fyneWindow := fynerisor.NewApp("Simple Imports Example")
	fyneWindow.Resize(600, 400)

	// Import utility functions first
	err := fyneWindow.ImportScript("simple-utils.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load main script
	script, err := os.ReadFile("simple-app.risor")
	if err != nil {
		log.Fatal(err)
	}

	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show and run
	fyneWindow.ShowAndRun()
}
