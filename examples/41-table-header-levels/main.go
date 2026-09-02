package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	fyneWindow := gui.NewApp("Table Header Levels",
		gui.WithTie(),
		gui.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
	)

	script, err := os.ReadFile("app.risor")
	if err != nil {
		log.Fatal(err)
	}

	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	fyneWindow.ShowAndRun()
}
