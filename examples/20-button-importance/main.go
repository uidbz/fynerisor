package main

import (
	"os"

	"github.com/uidbz/fynerisor/gui"
)

func main() {
	fw := gui.NewApp("Button Importance Example")

	script, err := os.ReadFile("script.risor")
	if err != nil {
		panic(err)
	}

	fw.LoadScript(string(script))
	fw.Execute()
	fw.ShowAndRun()
}
