package core

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file covers issue #41 (end-to-end): io owns the streams, json and
// csv are pure transforms, and they compose like shell filters —
// csv.parse(io.read(p)), io.write(p, json.format(v)), io.lines.

func TestPipelineComposition(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "data.csv")
	outPath := filepath.Join(dir, "out.json")
	if err := os.WriteFile(inPath, []byte("name,age\nAda,36\nGrace,45\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	cb := NewContext(WithIO(), WithCSV(), WithJSON(), WithStdout(&out),
		WithGlobal("inPath", inPath), WithGlobal("outPath", outPath))

	if _, err := cb.Eval(`
		let rows = csv.parse(io.read(inPath))
		io.write(io.stdout, json.format(rows))
	`); err != nil {
		t.Fatalf("eval: %v", err)
	}
	if !strings.Contains(out.String(), `"Ada"`) || !strings.Contains(out.String(), "36") {
		t.Errorf("stdout = %q, want the rows as JSON with ints kept integral", out.String())
	}

	if _, err := cb.Eval(`
		let lines = io.lines(inPath)
		let csv_out = csv.format([["h1","h2"],["v1","v2"]])
		io.write(outPath, csv_out)
		[len(lines), io.read(outPath)]
	`); err != nil {
		t.Fatalf("eval: %v", err)
	}
	written, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "v1,v2") {
		t.Errorf("out.json = %q, want the csv.format output", written)
	}
}
