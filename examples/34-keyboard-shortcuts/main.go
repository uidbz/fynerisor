package main

import (
	"flag"
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	// Parse command line flag for which script to run
	withMenu := flag.Bool("menu", false, "Run with menu example")
	flag.Parse()

	// Determine which script to load
	scriptFile := "main.risor"
	title := "Keyboard Shortcuts Example"
	if *withMenu {
		scriptFile = "with-menu.risor"
		title = "Keyboard Shortcuts with Menu"
	}

	// Read the Risor script
	script, err := os.ReadFile(scriptFile)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new fynerisor application
	fw := gui.NewApp(title)
	fw.Resize(600, 300)

	// Load and execute the script
	fw.LoadScript(string(script))
	fw.Execute()

	// Show the window and run the app
	fw.ShowAndRun()
}
