package main

import (
	"fmt"
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

	ctx := core.NewContext(core.WithTie())
	result, err := ctx.Eval(string(script))
	if err != nil {
		log.Fatalf("script error: %v", err)
	}

	fmt.Println("Result:", result)
}
