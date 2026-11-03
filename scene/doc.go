// Package scene defines the fynerisor scene document: the declarative,
// data-only description of a page that the designer edits and the renderer
// plays.
//
// A scene document is canonical JSON (format "fynerisor.scene/1"). It owns
// structure, placement and properties; behaviour is expressed as declarative
// actions in the document and, optionally, in a sidecar Risor script that
// finds nodes by ID. The package is deliberately free of Fyne, Risor and GUI
// dependencies so it can be used headless (validation, generation, storage)
// and imported from the CLI.
//
// The companion packages are:
//
//   - [github.com/uidbz/fynerisor/scene/render]: the Fyne runtime.
//   - [github.com/uidbz/fynerisor/designer]: the in-place editor.
//   - [github.com/uidbz/fynerisor/gui/scene]: the opt-in Risor binding.
package scene

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
)

// Format is the current scene document format.
const Format = "fynerisor.scene/1"

// Document is a complete scene: metadata plus one root node tree.
type Document struct {
	Format string            `json:"format"`
	ID     string            `json:"id,omitempty"` // stable identity (tie: the scene dir UID)
	Title  string            `json:"title,omitempty"`
	Script string            `json:"script,omitempty"` // optional sidecar script, relative reference
	Meta   map[string]string `json:"meta,omitempty"`
	Root   *Node             `json:"root"`
}

// Node is one element of the scene tree. The meaning of Place and the order
// of Children depend on the *parent's* component kind (see ComponentSpec).
type Node struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Name     string              `json:"name,omitempty"`
	Place    Place               `json:"place,omitzero"`
	Props    map[string]any      `json:"props,omitempty"`
	Style    *Style              `json:"style,omitempty"`
	On       map[string][]Action `json:"on,omitempty"`
	Hidden   bool                `json:"hidden,omitempty"`
	Locked   bool                `json:"locked,omitempty"` // editor only: input passes through
	Children []*Node             `json:"children,omitempty"`
}

// Place is the position of a node inside its parent. Free parents use X, Y, W
// and H (design units) and optionally Anchor; ordered parents use MinW/MinH as
// a size hint; slot parents use Slot; tabs use Title and Icon.
type Place struct {
	X, Y, W, H float64
	Anchor     *Anchor
	Slot       string
	Title      string
	Icon       string
	MinW, MinH float64
}

// Anchor keeps a free node's offsets fixed when its container resizes. It is
// only used when the free container's scale mode is "none". H is one of
// left, right, center, stretch; V is one of top, bottom, center, stretch.
type Anchor struct {
	H string `json:"h,omitempty"`
	V string `json:"v,omitempty"`
}

// Style overrides theme values for a node and its subtree.
type Style struct {
	Fg       string  `json:"fg,omitempty"`
	Bg       string  `json:"bg,omitempty"`
	TextSize float64 `json:"textSize,omitempty"`
	Bold     bool    `json:"bold,omitempty"`
	Italic   bool    `json:"italic,omitempty"`
}

// Background is a free container's backdrop. Src may be a scene asset
// reference (tie:hash, https URL, file: URL or a path relative to the
// document). Fill is stretch, contain, cover or original. Color is used behind
// and around the image (the letterbox area).
type Background struct {
	Src   string `json:"src,omitempty"`
	Fill  string `json:"fill,omitempty"`
	Color string `json:"color,omitempty"`
}

// Action is one declarative behaviour step. Verbs are listed in Verbs.
type Action struct {
	Do      string         `json:"do"`
	To      string         `json:"to,omitempty"`
	Node    string         `json:"node,omitempty"`
	Prop    string         `json:"prop,omitempty"`
	Value   any            `json:"value,omitempty"`
	Handler string         `json:"handler,omitempty"` // call: page.Bind name
	Ref     string         `json:"ref,omitempty"`     // open/play: defaults to the event item
	Args    map[string]any `json:"args,omitempty"`
}

// Verbs is the declarative action vocabulary.
var Verbs = []string{
	"navigate", "back", "show", "hide", "toggle", "set",
	"call", "open", "play", "pause", "select",
}

// Events every component may use. Component specs list the subset they emit.
var Events = []string{
	"tapped", "doubleTapped", "secondaryTapped",
	"hoverIn", "hoverOut", "changed", "submitted",
	"itemTapped", "selected",
}

// IsZero reports whether the place carries no placement information.
func (p Place) IsZero() bool {
	return p.X == 0 && p.Y == 0 && p.W == 0 && p.H == 0 &&
		p.Anchor == nil && p.Slot == "" && p.Title == "" && p.Icon == "" &&
		p.MinW == 0 && p.MinH == 0
}

// MarshalJSON writes only the fields the placement actually uses, so
// hand-written and canonical documents stay readable.
func (p Place) MarshalJSON() ([]byte, error) {
	out := make(map[string]any, 9)
	if p.X != 0 {
		out["x"] = p.X
	}
	if p.Y != 0 {
		out["y"] = p.Y
	}
	if p.W != 0 {
		out["w"] = p.W
	}
	if p.H != 0 {
		out["h"] = p.H
	}
	if p.Anchor != nil && (p.Anchor.H != "" || p.Anchor.V != "") {
		out["anchor"] = p.Anchor
	}
	if p.Slot != "" {
		out["slot"] = p.Slot
	}
	if p.Title != "" {
		out["title"] = p.Title
	}
	if p.Icon != "" {
		out["icon"] = p.Icon
	}
	if p.MinW != 0 {
		out["minW"] = p.MinW
	}
	if p.MinH != 0 {
		out["minH"] = p.MinH
	}
	return json.Marshal(out)
}

// UnmarshalJSON reads the sparse placement representation.
func (p *Place) UnmarshalJSON(data []byte) error {
	var in struct {
		X      float64 `json:"x"`
		Y      float64 `json:"y"`
		W      float64 `json:"w"`
		H      float64 `json:"h"`
		Anchor *Anchor `json:"anchor"`
		Slot   string  `json:"slot"`
		Title  string  `json:"title"`
		Icon   string  `json:"icon"`
		MinW   float64 `json:"minW"`
		MinH   float64 `json:"minH"`
	}
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}
	*p = Place{
		X: in.X, Y: in.Y, W: in.W, H: in.H, Anchor: in.Anchor,
		Slot: in.Slot, Title: in.Title, Icon: in.Icon,
		MinW: in.MinW, MinH: in.MinH,
	}
	return nil
}

// IsZero reports whether no style override is set.
func (s Style) IsZero() bool {
	return s.Fg == "" && s.Bg == "" && s.TextSize == 0 && !s.Bold && !s.Italic
}

// Prop returns a property value.
func (n *Node) Prop(name string) (any, bool) {
	if n == nil || n.Props == nil {
		return nil, false
	}
	v, ok := n.Props[name]
	return v, ok
}

// StringProp returns a string property, or def.
func (n *Node) StringProp(name, def string) string {
	if v, ok := n.Prop(name); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

// FloatProp returns a numeric property, or def.
func (n *Node) FloatProp(name string, def float64) float64 {
	v, ok := n.Prop(name)
	if !ok {
		return def
	}
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		if f, err := x.Float64(); err == nil {
			return f
		}
	}
	return def
}

// IntProp returns an integer property, or def.
func (n *Node) IntProp(name string, def int) int {
	return int(math.Round(n.FloatProp(name, float64(def))))
}

// BoolProp returns a boolean property, or def.
func (n *Node) BoolProp(name string, def bool) bool {
	if v, ok := n.Prop(name); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

// StringListProp returns a list-of-strings property, or nil.
func (n *Node) StringListProp(name string) []string {
	v, ok := n.Prop(name)
	if !ok {
		return nil
	}
	switch x := v.(type) {
	case []string:
		return x
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// SetProp sets a property, creating the map when needed.
func (n *Node) SetProp(name string, value any) {
	if n.Props == nil {
		n.Props = make(map[string]any)
	}
	n.Props[name] = value
}

// Clone deep-copies the node tree.
func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}
	out := *n
	out.Place = n.Place
	if n.Place.Anchor != nil {
		a := *n.Place.Anchor
		out.Place.Anchor = &a
	}
	if n.Style != nil {
		s := *n.Style
		out.Style = &s
	}
	if n.Props != nil {
		out.Props = cloneMap(n.Props)
	}
	if n.On != nil {
		out.On = make(map[string][]Action, len(n.On))
		for ev, acts := range n.On {
			cp := make([]Action, len(acts))
			for i, a := range acts {
				cp[i] = a
				if a.Value != nil {
					cp[i].Value = cloneValue(a.Value)
				}
				if a.Args != nil {
					cp[i].Args = cloneMap(a.Args)
				}
			}
			out.On[ev] = cp
		}
	}
	if n.Children != nil {
		out.Children = make([]*Node, len(n.Children))
		for i, c := range n.Children {
			out.Children[i] = c.Clone()
		}
	}
	return &out
}

// Clone deep-copies the document.
func (d *Document) Clone() *Document {
	if d == nil {
		return nil
	}
	out := *d
	if d.Meta != nil {
		out.Meta = make(map[string]string, len(d.Meta))
		for k, v := range d.Meta {
			out.Meta[k] = v
		}
	}
	out.Root = d.Root.Clone()
	return &out
}

// Find returns the node with the given ID.
func (d *Document) Find(id string) (*Node, bool) {
	var found *Node
	d.Each(func(n, _ *Node, _ int, _ int) bool {
		if n.ID == id {
			found = n
			return false
		}
		return true
	})
	return found, found != nil
}

// Locate returns a node together with its parent and child index.
func (d *Document) Locate(id string) (node, parent *Node, index int, ok bool) {
	d.Each(func(n, p *Node, i int, _ int) bool {
		if n.ID == id {
			node, parent, index, ok = n, p, i, true
			return false
		}
		return true
	})
	return
}

// Each walks the tree depth-first in document order. fn receives the node, its
// parent (nil for the root), its child index (0 for the root) and its depth.
// Returning false stops the walk.
func (d *Document) Each(fn func(n, parent *Node, index, depth int) bool) {
	if d == nil || d.Root == nil {
		return
	}
	var walk func(n, parent *Node, index, depth int) bool
	walk = func(n, parent *Node, index, depth int) bool {
		if !fn(n, parent, index, depth) {
			return false
		}
		for i, c := range n.Children {
			if !walk(c, n, i, depth+1) {
				return false
			}
		}
		return true
	}
	walk(d.Root, nil, 0, 0)
}

// Each is Each with a simpler callback signature, kept for convenience in
// UI code that does not need parent or depth.
func (d *Document) Walk(fn func(*Node) bool) {
	d.Each(func(n, _ *Node, _ int, _ int) bool { return fn(n) })
}

// IDs returns every node indexed by ID.
func (d *Document) IDs() map[string]*Node {
	out := make(map[string]*Node)
	d.Walk(func(n *Node) bool {
		out[n.ID] = n
		return true
	})
	return out
}

// Count returns the number of nodes in the document.
func (d *Document) Count() int {
	n := 0
	d.Walk(func(*Node) bool { n++; return true })
	return n
}

// Count returns the number of nodes in the subtree (including n).
func (n *Node) Count() int {
	count := 0
	n.Walk(func(*Node) bool { count++; return true })
	return count
}

// Depth returns the maximum nesting depth (root = 1).
func (d *Document) Depth() int {
	max := 0
	d.Each(func(_ *Node, _ *Node, _ int, depth int) bool {
		if depth+1 > max {
			max = depth + 1
		}
		return true
	})
	return max
}

// ParentOf returns the parent of id and the child's index.
func (d *Document) ParentOf(id string) (*Node, int, bool) {
	_, p, i, ok := d.Locate(id)
	return p, i, ok
}

// IsAncestor reports whether a is b or an ancestor of b.
func (d *Document) IsAncestor(a, b string) bool {
	an, ok := d.Find(a)
	if !ok {
		return false
	}
	found := false
	an.Walk(func(n *Node) bool {
		if n.ID == b {
			found = true
			return false
		}
		return true
	})
	return found
}

// Walk walks the node subtree.
func (n *Node) Walk(fn func(*Node) bool) {
	if n == nil {
		return
	}
	if !fn(n) {
		return
	}
	for _, c := range n.Children {
		c.Walk(fn)
	}
}

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// ValidID reports whether id is a legal scene node ID.
func ValidID(id string) bool { return idPattern.MatchString(id) }

// Slug converts arbitrary text into an ID-safe stem.
func Slug(s string) string {
	var b strings.Builder
	lastDash := true
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "node"
	}
	if len(out) > 40 {
		out = strings.Trim(out[:40], "-")
	}
	return out
}

// ReID gives every node in the subtree a fresh ID and returns the mapping from
// old to new IDs. Actions that target IDs inside the copied subtree are
// rewritten; callers pass the IDs already in use in the destination document.
func (n *Node) ReID(taken map[string]bool) map[string]string {
	m := make(map[string]string)
	n.Walk(func(x *Node) bool {
		base := baseID(x.ID)
		id := base + "-copy"
		for i := 2; taken[id]; i++ {
			id = fmt.Sprintf("%s-copy-%d", base, i)
		}
		taken[id] = true
		m[x.ID] = id
		return true
	})
	n.Walk(func(x *Node) bool {
		x.ID = m[x.ID]
		for ev, acts := range x.On {
			for i := range acts {
				if _, ok := m[acts[i].Node]; ok {
					acts[i].Node = m[acts[i].Node]
				}
			}
			x.On[ev] = acts
		}
		return true
	})
	return m
}

func baseID(id string) string {
	for {
		if strings.HasSuffix(id, "-copy") {
			id = strings.TrimSuffix(id, "-copy")
			continue
		}
		if i := strings.LastIndex(id, "-copy-"); i > 0 {
			id = id[:i]
			continue
		}
		return Slug(id)
	}
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = cloneValue(v)
	}
	return out
}

func cloneValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		return cloneMap(x)
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = cloneValue(e)
		}
		return out
	case []string:
		return append([]string(nil), x...)
	default:
		return v
	}
}

// Round rounds a design-unit value to the canonical precision (0.01).
func Round(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return math.Round(v*100) / 100
}
