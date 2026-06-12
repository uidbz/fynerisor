package main

import (
	"log"
	"os"

	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/gui"
)

func main() {
	core.SetAppVersion("0.6.0")

	// Create fynerisor window
	fw := gui.NewApp("Module Imports Example",
		gui.WithStatusCallback(func(status string) {
			log.Println("Status:", status)
		}),
	)

	// Load main script
	script, err := os.ReadFile("main.risor")
	if err != nil {
		log.Fatal(err)
	}

	fw.LoadScript(string(script))
	fw.Execute()
	fw.ShowAndRun()
}
