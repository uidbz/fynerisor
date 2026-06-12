package main

import (
	"log"
	"os"

	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/gui"
)

func main() {
	core.SetAppVersion("1.0.0")

	// Create the custom user database object
	userDB := NewUserDatabaseObject()

	// Create fynerisor window with custom global and status callback
	w := gui.NewApp("Custom Struct Example",
		gui.WithGlobal("users", userDB),
		gui.WithStatusCallback(func(status string) {
			log.Printf("Status: %s", status)
		}),
	)

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
