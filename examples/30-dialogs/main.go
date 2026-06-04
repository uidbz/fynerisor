package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	// Create fynerisor window
	fyneWindow := gui.NewApp("Dialog Examples",
		gui.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
	)

	// Load basic or advanced example
	scriptFile := "dialogs.risor"
	if len(os.Args) > 1 && os.Args[1] == "advanced" {
		scriptFile = "advanced-dialogs.risor"
	}

	script, err := os.ReadFile(scriptFile)
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show window and run the application
	fyneWindow.ShowAndRun()
}
