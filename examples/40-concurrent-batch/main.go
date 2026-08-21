package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/uidbz/fynerisor/core"
)

// Headless parallel batch: compile one Risor script once and run it over many
// CSV files concurrently, each on its own isolated VM, using core.EvalBatch.
func main() {
	files, err := filepath.Glob("data/*.csv")
	if err != nil {
		log.Fatalf("glob: %v", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		log.Fatal("no CSV files found in data/ (run from the example directory)")
	}

	script, err := os.ReadFile("worker.risor")
	if err != nil {
		log.Fatalf("read worker.risor: %v", err)
	}

	// One input per file; each becomes the global `file` for that run.
	inputs := make([]map[string]any, len(files))
	for i, f := range files {
		inputs[i] = map[string]any{"file": f}
	}

	results, err := core.EvalBatch(
		context.Background(),
		string(script),
		inputs,
		core.BatchConfig{Concurrency: 4},
		core.WithCSV(),
	)
	if err != nil {
		log.Fatalf("batch: %v", err) // setup/compile failure only
	}

	// Results are index-aligned with inputs (and therefore with files).
	for i, r := range results {
		if r.Err != nil {
			fmt.Printf("%-16s ERROR: %v\n", files[i], r.Err)
			continue
		}
		fmt.Printf("%-16s %v\n", files[i], r.Value)
	}
}
