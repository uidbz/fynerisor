package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/uidbz/fynerisor/internal/sceneedit"
	"github.com/uidbz/fynerisor/scene"
)

// runHandlers rewrites a scene sidecar script for the editor workflow:
// standalone on_<node>_<event> handlers, one page.Bind section at the bottom,
// and a stub for every call action that has no handler yet.
func runHandlers(argv []string) int {
	fs := flag.NewFlagSet("handlers", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: fynerisor handlers [-write] [-json] <page.scene.json>")
		fmt.Fprintln(os.Stderr, "  rewrites the sidecar script and, with -write, the scene document")
	}
	write := fs.Bool("write", false, "write the rewritten script and document back")
	asJSON := fs.Bool("json", false, "print the report as JSON")
	if err := fs.Parse(argv); err != nil {
		return 2
	}
	args := fs.Args()
	if len(args) != 1 {
		fs.Usage()
		return 2
	}
	path := args[0]

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fynerisor: %v\n", err)
		return 1
	}
	doc, err := scene.Decode(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		return 1
	}
	if doc.Script == "" {
		fmt.Fprintf(os.Stderr, "%s: no sidecar script (set \"script\" in the document)\n", path)
		return 1
	}
	ref, err := scene.ResolveRef(path, doc.Script)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: script: %v\n", path, err)
		return 1
	}
	scriptPath, err := scene.FilePath(ref)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: script: %v\n", path, err)
		return 1
	}
	src, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fynerisor: %v\n", err)
		return 1
	}

	res, err := sceneedit.Apply(doc, string(src))
	if err != nil {
		fmt.Fprintf(os.Stderr, "fynerisor: %v\n", err)
		return 1
	}
	for _, w := range res.Report.Warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}
	if *asJSON {
		printReportJSON(res.Report)
		if !*write {
			return 0
		}
	} else {
		printReport(res.Report)
		if !*write {
			fmt.Print(res.Source)
			return 0
		}
	}
	if res.Source != string(src) {
		if err := os.WriteFile(scriptPath, []byte(res.Source), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "fynerisor: write %s: %v\n", scriptPath, err)
			return 1
		}
	}
	encoded, err := scene.Encode(doc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fynerisor: encode %s: %v\n", path, err)
		return 1
	}
	if !bytes.Equal(encoded, data) {
		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "fynerisor: write %s: %v\n", path, err)
			return 1
		}
	}
	return 0
}

func printReport(r sceneedit.Report) {
	if len(r.Created) > 0 {
		fmt.Fprintf(os.Stderr, "created: %v\n", r.Created)
	}
	if len(r.Converted) > 0 {
		fmt.Fprintf(os.Stderr, "converted inline handlers: %v\n", r.Converted)
	}
	if len(r.Bound) > 0 {
		fmt.Fprintf(os.Stderr, "bound: %v\n", r.Bound)
	}
}

func printReportJSON(r sceneedit.Report) {
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
		"created":   r.Created,
		"converted": r.Converted,
		"bound":     r.Bound,
		"warnings":  r.Warnings,
	})
}
