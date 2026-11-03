package scene_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/builtin"
)

func registry() *scene.Registry {
	r := scene.NewRegistry()
	builtin.Register(r)
	return r
}

func sampleDoc() *scene.Document {
	return &scene.Document{
		Title: "Living room",
		Root: &scene.Node{
			ID: "root", Type: "free",
			Props: map[string]any{"designW": 1600.0, "designH": 900.0, "scale": "fit"},
			Children: []*scene.Node{
				{
					ID: "door", Type: "hotspot", Name: "Door",
					Place: scene.Place{X: 1210, Y: 300, W: 140, H: 380},
					On: map[string][]scene.Action{
						"tapped": {{Do: "navigate", To: "../kitchen"}},
					},
				},
				{
					ID: "wall", Type: "vbox",
					Place: scene.Place{X: 80, Y: 120, W: 520, H: 600},
					Children: []*scene.Node{
						{ID: "title", Type: "label", Props: map[string]any{"text": "Summer 2014"}},
						{ID: "gallery", Type: "slot", Place: scene.Place{MinH: 480}},
					},
				},
				{ID: "escape", Type: "link", Props: map[string]any{"text": "Back", "to": "../"}},
			},
		},
	}
}

func mustEncode(t *testing.T, d *scene.Document) string {
	t.Helper()
	b, err := scene.Encode(d)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return string(b)
}

func TestEncodeCanonical(t *testing.T) {
	d := sampleDoc()
	first := mustEncode(t, d)
	second := mustEncode(t, d)
	if first != second {
		t.Fatalf("encoding is not stable:\n%s\n---\n%s", first, second)
	}
	if !strings.HasSuffix(first, "\n") {
		t.Fatal("canonical encoding must end with a newline")
	}
	if !strings.Contains(first, `"format": "fynerisor.scene/1"`) {
		t.Fatalf("format missing:\n%s", first)
	}
	// Floats are rounded to 0.01.
	d.Root.Children[0].Place.X = 1210.123456
	if got := mustEncode(t, d); !strings.Contains(got, `"x": 1210.12`) {
		t.Fatalf("float not rounded:\n%s", got)
	}
	// Encoding must not mutate the document.
	if d.Root.Children[0].Place.X != 1210.123456 {
		t.Fatal("Encode mutated the document")
	}
}

func TestDecodeRoundTrip(t *testing.T) {
	d := sampleDoc()
	data := []byte(mustEncode(t, d))
	back, err := scene.Decode(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got, want := mustEncode(t, back), string(data); got != want {
		t.Fatalf("round trip changed bytes:\n%s\n---\n%s", got, want)
	}
}

func TestDecodeRejects(t *testing.T) {
	good := mustEncode(t, sampleDoc())
	tests := []struct {
		name string
		data string
		want string
	}{
		{"not-json", `{`, "parse"},
		{"wrong-format", strings.Replace(good, scene.Format, "other/1", 1), "not a scene document"},
		{"future-format", strings.Replace(good, scene.Format, "fynerisor.scene/9", 1), "unsupported format"},
		{"no-root", `{"format":"fynerisor.scene/1"}`, "no root"},
		{"dup-id", strings.Replace(good, `"id": "escape"`, `"id": "door"`, 1), "duplicate"},
		{"bad-id", strings.Replace(good, `"id": "escape"`, `"id": "Escape Me"`, 1), "invalid node id"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := scene.Decode([]byte(tc.data))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
}

func TestDecodeLimits(t *testing.T) {
	big := strings.Repeat("x", scene.MaxDocumentBytes+1)
	if _, err := scene.Decode([]byte(big)); err == nil {
		t.Fatal("expected size limit error")
	}
	deep := &scene.Node{ID: "root", Type: "free"}
	cur := deep
	for i := 0; i < 40; i++ {
		child := &scene.Node{ID: "n" + strings.Repeat("a", i+1), Type: "free"}
		cur.Children = []*scene.Node{child}
		cur = child
	}
	data, err := json.Marshal(&scene.Document{Format: scene.Format, Root: deep})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scene.Decode(data); err == nil || !strings.Contains(err.Error(), "depth") {
		t.Fatalf("expected depth error, got %v", err)
	}
}

func TestValidate(t *testing.T) {
	reg := registry()
	issues := scene.Validate(sampleDoc(), reg)
	for _, i := range issues {
		if i.Severity == scene.Error {
			t.Errorf("unexpected error: %s", i)
		}
	}

	bad := sampleDoc()
	bad.Root.Children[1].Children[0].Children = []*scene.Node{{ID: "orphan", Type: "label"}}
	bad.Root.Children[0].On["tapped"] = []scene.Action{{Do: "teleport"}}
	bad.Root.Children[0].Props = map[string]any{"nope": true}
	bad.Root.Children[2].Props = map[string]any{"text": 42}
	issues = scene.Validate(bad, reg)
	var messages []string
	for _, i := range issues {
		messages = append(messages, i.String())
	}
	joined := strings.Join(messages, "\n")
	for _, want := range []string{"leaf but has", "unknown action verb", "unknown property", "must be a string"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
	if scene.Errors(issues) < 3 {
		t.Fatalf("expected at least 3 errors, got %d", scene.Errors(issues))
	}
}

func TestValidateSlots(t *testing.T) {
	reg := registry()
	d := &scene.Document{
		Root: &scene.Node{ID: "root", Type: "border", Children: []*scene.Node{
			{ID: "a", Type: "label", Place: scene.Place{Slot: "center"}},
			{ID: "b", Type: "label", Place: scene.Place{Slot: "center"}},
			{ID: "c", Type: "label", Place: scene.Place{Slot: "middle"}},
			{ID: "d", Type: "label"},
		}},
	}
	issues := scene.Validate(d, reg)
	joined := ""
	for _, i := range issues {
		joined += i.String() + "\n"
	}
	for _, want := range []string{"more than one child", "unknown slot", "needs a slot"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
}

func TestOpsUndoRoundTrip(t *testing.T) {
	d := sampleDoc()
	before := mustEncode(t, d)

	door, _ := d.Find("door")
	oldProp, _ := door.Prop("showInPlay")
	container := &scene.Node{ID: "tools", Type: "hbox", Place: scene.Place{W: 300, H: 60}}
	rm, err := scene.RemoveOp(d, "escape")
	if err != nil {
		t.Fatal(err)
	}
	mv, err := scene.MoveOp(d, "door", "root", 2)
	if err != nil {
		t.Fatal(err)
	}
	tx := scene.Tx{Label: "edit", Ops: []scene.Op{
		scene.SetPropOp("door", "showInPlay", oldProp, true),
		scene.SetPlaceOp("door", door.Place, scene.Place{X: 10, Y: 20, W: 100, H: 50}),
		scene.SetFieldOp("door", "name", door.Name, "Front door"),
		scene.InsertOp("root", 1, container),
		mv,
		rm,
	}}
	changes, err := d.Apply(tx)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(changes) != len(tx.Ops) {
		t.Fatalf("got %d changes, want %d", len(changes), len(tx.Ops))
	}
	if _, ok := d.Find("escape"); ok {
		t.Fatal("escape was not removed")
	}
	tools, ok := d.Find("tools")
	if !ok {
		t.Fatal("tools was not inserted")
	}
	if tools.Place.W != 300 {
		t.Fatalf("tool place lost: %+v", tools.Place)
	}
	moved, _ := d.Find("door")
	if moved.Place.X != 10 || moved.Place.W != 100 {
		t.Fatalf("place not applied: %+v", moved.Place)
	}
	if moved.Name != "Front door" {
		t.Fatalf("name not applied: %q", moved.Name)
	}

	if _, err := d.Apply(tx.Invert()); err != nil {
		t.Fatalf("invert: %v", err)
	}
	if got := mustEncode(t, d); got != before {
		t.Fatalf("undo did not restore the document:\n%s\n---\n%s", before, got)
	}
}

func TestOpsRejections(t *testing.T) {
	d := sampleDoc()
	if _, err := d.Apply(scene.Tx{Ops: []scene.Op{{Kind: scene.OpRemove, Node: "root"}}}); err == nil {
		t.Fatal("removing the root must fail")
	}
	if _, err := d.Apply(scene.Tx{Ops: []scene.Op{{Kind: scene.OpMove, Node: "wall", Parent: "gallery"}}}); err == nil {
		t.Fatal("moving into a descendant must fail")
	}
	dup := &scene.Node{ID: "door", Type: "label"}
	if _, err := d.Apply(scene.Tx{Ops: []scene.Op{scene.InsertOp("root", 0, dup)}}); err == nil {
		t.Fatal("duplicate ids must fail")
	}
	// A failed transaction leaves the document untouched.
	before := mustEncode(t, d)
	if _, err := d.Apply(scene.Tx{Ops: []scene.Op{
		scene.SetPropOp("door", "showInPlay", nil, true),
		{Kind: scene.OpSetProp, Node: "missing", Key: "x", New: 1},
	}}); err == nil {
		t.Fatal("expected failure")
	}
	if got := mustEncode(t, d); got != before {
		t.Fatal("failed transaction partially applied")
	}
}

func TestHistory(t *testing.T) {
	d := sampleDoc()
	h := scene.NewHistory(10)
	if h.Dirty() {
		t.Fatal("fresh history must be clean")
	}
	door, _ := d.Find("door")
	tx1 := scene.Tx{Label: "move", MergeKey: "drag:door", Ops: []scene.Op{
		scene.SetPlaceOp("door", door.Place, scene.Place{X: 1, Y: 1, W: 10, H: 10}),
	}}
	if _, err := d.Apply(tx1); err != nil {
		t.Fatal(err)
	}
	h.Push(tx1)
	door, _ = d.Find("door")
	tx2 := scene.Tx{Label: "move", MergeKey: "drag:door", Ops: []scene.Op{
		scene.SetPlaceOp("door", door.Place, scene.Place{X: 2, Y: 2, W: 10, H: 10}),
	}}
	if _, err := d.Apply(tx2); err != nil {
		t.Fatal(err)
	}
	h.Push(tx2)
	if !h.Dirty() {
		t.Fatal("history must be dirty after an edit")
	}
	h.MarkSaved()
	if h.Dirty() {
		t.Fatal("history must be clean after MarkSaved")
	}
	// Two placements with the same merge key are one undo step.
	changes, ok, err := h.Undo(d)
	if err != nil || !ok || len(changes) == 0 {
		t.Fatalf("undo: ok=%v err=%v", ok, err)
	}
	door, _ = d.Find("door")
	if door.Place.X != 1210 {
		t.Fatalf("undo did not restore the original place: %+v", door.Place)
	}
	if !h.Dirty() {
		t.Fatal("undo below the saved position must be dirty")
	}
	if _, ok, _ := h.Redo(d); !ok {
		t.Fatal("redo failed")
	}
	if h.Dirty() {
		t.Fatal("redo back to the saved position must be clean")
	}
}

func TestStores(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	ref := filepath.Join(dir, "page.scene.json")
	fs := scene.FileStore{}
	data := []byte(mustEncode(t, sampleDoc()))

	snap, err := fs.Save(ctx, ref, data, "")
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if snap.Version != scene.VersionOf(data) {
		t.Fatalf("version %q, want %q", snap.Version, scene.VersionOf(data))
	}
	loaded, err := fs.Load(ctx, ref)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if string(loaded.Data) != string(data) || loaded.Version != snap.Version {
		t.Fatal("load did not return the saved bytes")
	}
	if _, err := fs.Save(ctx, ref, data, "stale"); !errors.Is(err, scene.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if _, err := fs.Save(ctx, ref, []byte("{}"), snap.Version); err != nil {
		t.Fatalf("save with the right base: %v", err)
	}
	if _, err := fs.Load(ctx, filepath.Join(dir, "missing.scene.json")); !errors.Is(err, scene.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if _, err := os.Stat(ref); err != nil {
		t.Fatal(err)
	}

	mem := scene.NewMemoryStore()
	mem.Put("mem", data)
	got, err := mem.Load(ctx, "mem")
	if err != nil || string(got.Data) != string(data) {
		t.Fatalf("memory load: %v", err)
	}
	if _, err := mem.Save(ctx, "mem", data, "wrong"); !errors.Is(err, scene.ErrConflict) {
		t.Fatalf("memory conflict: %v", err)
	}
}

func TestResolveRef(t *testing.T) {
	tests := []struct{ base, ref, want string }{
		{"", "a.scene.json", "a.scene.json"},
		{"/srv/scenes/house.json", "kitchen.json", "/srv/scenes/kitchen.json"},
		{"/srv/scenes/house.json", "../start.json", "/srv/start.json"},
		{"file:///srv/scenes/house.json", "kitchen.json", "file:///srv/scenes/kitchen.json"},
		{"https://example.com/app/house.scene.json", "../kitchen.scene.json", "https://example.com/kitchen.scene.json"},
		{"tie:/scenes/house/living-room/scene.json", "../kitchen", "tie:/scenes/house/kitchen"},
		{"https://example.com/a.json", "https://other.example/b.json", "https://other.example/b.json"},
		{"https://example.com/a.json", "tie:abc123", "tie:abc123"},
	}
	for _, tc := range tests {
		got, err := scene.ResolveRef(tc.base, tc.ref)
		if err != nil {
			t.Errorf("%q + %q: %v", tc.base, tc.ref, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%q + %q = %q, want %q", tc.base, tc.ref, got, tc.want)
		}
	}
}

func TestReID(t *testing.T) {
	n := &scene.Node{ID: "group", Type: "vbox", Children: []*scene.Node{
		{ID: "a", Type: "button", On: map[string][]scene.Action{
			"tapped": {{Do: "show", Node: "b"}},
		}},
		{ID: "b", Type: "label"},
	}}
	m := n.ReID(map[string]bool{"group": true})
	if m["group"] == "group" || m["a"] == "a" {
		t.Fatalf("ids were not replaced: %#v", m)
	}
	if n.Children[0].On["tapped"][0].Node != m["b"] {
		t.Fatalf("action target was not rewritten: %#v", n.Children[0].On)
	}
	if n.ID == n.Children[0].ID || n.Children[0].ID == n.Children[1].ID {
		t.Fatal("ids are not unique")
	}
}

func TestBuiltinRegistry(t *testing.T) {
	reg := registry()
	for _, typ := range []string{
		"free", "vbox", "hbox", "grid", "border", "tabs", "label", "image",
		"hotspot", "button", "entry", "slot", "embed", "link",
	} {
		if _, ok := reg.Spec(typ); !ok {
			t.Errorf("missing builtin %q", typ)
		}
	}
	if reg.Len() < 30 {
		t.Fatalf("only %d builtin specs", reg.Len())
	}
	for _, typ := range reg.Types() {
		spec, _ := reg.Spec(typ)
		for _, prop := range spec.Props {
			if prop.Title == "" {
				t.Errorf("%s.%s has no title", typ, prop.Name)
			}
			if prop.Group == "" {
				t.Errorf("%s.%s has no group", typ, prop.Name)
			}
		}
	}
}

func TestPolygonValidation(t *testing.T) {
	reg := registry()
	d := &scene.Document{Root: &scene.Node{ID: "root", Type: "free", Children: []*scene.Node{
		{ID: "poly", Type: "hotspot", Place: scene.Place{W: 100, H: 100},
			Props: map[string]any{"shape": "polygon", "points": []any{
				map[string]any{"x": 0.0, "y": 0.0},
				map[string]any{"x": 1.0, "y": 0.0},
				map[string]any{"x": 0.5, "y": 1.0},
			}}},
	}}}
	if issues := scene.Validate(d, reg); scene.Errors(issues) != 0 {
		t.Fatalf("valid polygon: %v", issues)
	}
	d.Root.Children[0].SetProp("points", []any{map[string]any{"x": 2.0, "y": 0.0}})
	if issues := scene.Validate(d, reg); scene.Errors(issues) == 0 {
		t.Fatal("out-of-range point was accepted")
	}
}
