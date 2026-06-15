package main

import (
	"log"
	"os"

	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/gui"
)

const AppVersion = "1.2.3"

func main() {
	// Set the application version
	// Scripts can now use require(["v1.2"]) to check against this version
	core.SetAppVersion(AppVersion)

	// Create fynerisor window
	w := gui.NewApp("App Versioning Example v" + AppVersion)

	// Load Risor script from file
	script, err := os.ReadFile("script.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	w.LoadScript(string(script))
	w.Execute()

	// Show window and run the application
	w.ShowAndRun()
}
