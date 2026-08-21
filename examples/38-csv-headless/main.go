package main

import (
	"log"
	"os"

	"github.com/uidbz/fynerisor/core"
)

func main() {
	scriptPath := "app.risor"
	if len(os.Args) > 1 {
		scriptPath = os.Args[1]
	}

	script, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Fatalf("failed to read script: %v", err)
	}

	ctx := core.NewContext(
		core.WithCSV(),
	)
	if _, err := ctx.Eval(string(script)); err != nil {
		log.Fatalf("script error: %v", err)
	}
}
