package main

import (
	"os"

	"github.com/uidbz/fynerisor"
)

func main() {
	fw := fynerisor.NewApp("TextGrid Example")

	script, err := os.ReadFile("script.risor")
	if err != nil {
		panic(err)
	}

	fw.LoadScript(string(script))
	fw.Execute()
	fw.ShowAndRun()
}
