package scene

import (
	"fmt"
	"math"
	"strings"
)

// Severity classifies a validation issue.
type Severity string

const (
	// Error issues make the document invalid: the designer refuses to save.
	Error Severity = "error"
	// Warning issues are shown in the problems panel but do not block saving.
	Warning Severity = "warning"
)

// Issue is one validation finding.
type Issue struct {
	Severity Severity
	NodeID   string
	Path     string
	Message  string
}

func (i Issue) String() string {
	where := i.Path
	if i.NodeID != "" {
		where = i.NodeID
		if i.Path != "" && !strings.HasPrefix(i.Path, "root") {
			where += ": " + i.Path
		}
	}
	if where == "" {
		return fmt.Sprintf("%s: %s", i.Severity, i.Message)
	}
	return fmt.Sprintf("%s: %s: %s", i.Severity, where, i.Message)
}

// Schema supplies component semantics for validation. *Registry implements it.
type Schema interface {
	Spec(typ string) (ComponentSpec, bool)
}

// Validate checks a document against a component schema. A nil schema performs
// only the structural checks that Decode already does for parsed documents.
func Validate(d *Document, schema Schema) []Issue {
	if d == nil || d.Root == nil {
		return []Issue{{Severity: Error, Message: "document has no root node"}}
	}
	var issues []Issue
	add := func(sev Severity, id, path, format string, args ...any) {
		issues = append(issues, Issue{Severity: sev, NodeID: id, Path: path, Message: fmt.Sprintf(format, args...)})
	}

	if d.Root.Hidden {
		add(Warning, d.Root.ID, "hidden", "the root node is hidden; the page will be blank")
	}
	if schema == nil {
		return issues
	}
	rootSpec, ok := schema.Spec(d.Root.Type)
	if !ok {
		add(Warning, d.Root.ID, "type", "unknown component type %q; it renders as a placeholder", d.Root.Type)
	} else if rootSpec.Kind == Leaf {
		add(Error, d.Root.ID, "type", "root component %q cannot have children", d.Root.Type)
	}

	d.Each(func(n, parent *Node, index, depth int) bool {
		spec, known := schema.Spec(n.Type)
		if !known {
			add(Warning, n.ID, "type", "unknown component type %q; it renders as a placeholder", n.Type)
			// Still check children structurally: a future composite may be valid.
			return true
		}
		path := "root"
		if parent != nil {
			path = fmt.Sprintf("root/children[%d]", index)
		}
		if n.Type == "" {
			add(Error, n.ID, path, "node has no type")
		}
		switch spec.Kind {
		case Leaf:
			if len(n.Children) > 0 {
				add(Error, n.ID, path, "component %q is a leaf but has %d children", n.Type, len(n.Children))
			}
		case Free:
			// The root free container derives its size from designW/H; only
			// placed free canvases use place.w/h.
			if parent != nil {
				if n.Place.W < 0 || n.Place.H < 0 {
					add(Error, n.ID, "place", "negative size %.1fx%.1f", n.Place.W, n.Place.H)
				}
				if n.Place.W == 0 || n.Place.H == 0 {
					add(Warning, n.ID, "place", "zero size %.1fx%.1f; the node will not be visible", n.Place.W, n.Place.H)
				}
			}
		case Ordered:
			if n.Place.Slot != "" {
				add(Warning, n.ID, "place", "slot %q is ignored inside an ordered container", n.Place.Slot)
			}
		}
		if spec.MaxChildren > 0 && len(n.Children) > spec.MaxChildren {
			add(Error, n.ID, path, "component %q allows at most %d children, has %d", n.Type, spec.MaxChildren, len(n.Children))
		}

		// Parent-kind rules apply to this node's place.
		if parent != nil {
			if pspec, ok := schema.Spec(parent.Type); ok {
				switch pspec.Kind {
				case Free:
					if n.Place.Slot != "" {
						add(Warning, n.ID, "place", "slot %q is ignored inside a free container", n.Place.Slot)
					}
				case Ordered, Tabs:
					if n.Place.Slot != "" {
						add(Warning, n.ID, "place", "slot %q is ignored inside %q", n.Place.Slot, parent.Type)
					}
				case Slots:
					if n.Place.Slot == "" {
						add(Error, n.ID, path, "component %q needs a slot (%s)", parent.Type, strings.Join(pspec.Slots, ", "))
					} else if !contains(pspec.Slots, n.Place.Slot) {
						add(Error, n.ID, "place", "unknown slot %q (want %s)", n.Place.Slot, strings.Join(pspec.Slots, ", "))
					}
				}
			}
		}

		if n.StringProp("shape", "") == "polygon" {
			if raw, ok := n.Prop("points"); ok {
				if pts, ok := raw.([]any); ok && len(pts) < 3 {
					add(Warning, n.ID, "props.points", "a polygon hotspot needs at least 3 points")
				}
			} else {
				add(Warning, n.ID, "props.points", "shape is polygon but points is empty")
			}
		}
		checkProps(n, spec, d, add)
		checkEvents(n, spec, d, add)
		checkSlots(n, spec, add)
		return true
	})
	return issues
}

// Errors counts error-severity issues.
func Errors(issues []Issue) int {
	n := 0
	for _, i := range issues {
		if i.Severity == Error {
			n++
		}
	}
	return n
}

func checkProps(n *Node, spec ComponentSpec, d *Document, add func(Severity, string, string, string, ...any)) {
	for name, v := range n.Props {
		p, ok := spec.Prop(name)
		if !ok {
			add(Warning, n.ID, "props."+name, "unknown property %q for component %q", name, spec.Type)
			continue
		}
		if msg := checkPropValue(p, v); msg != "" {
			add(Error, n.ID, "props."+name, "%s", msg)
			continue
		}
		if p.Kind == PropNodeRef {
			if s, _ := v.(string); s != "" {
				if _, ok := d.Find(s); !ok {
					add(Error, n.ID, "props."+name, "node %q does not exist", s)
				}
			}
		}
	}
	if n.Style != nil {
		for _, c := range []struct{ name, value string }{
			{"fg", n.Style.Fg}, {"bg", n.Style.Bg},
		} {
			if c.value != "" && !ValidColor(c.value) {
				add(Error, n.ID, "style."+c.name, "invalid color %q", c.value)
			}
		}
	}
}

func checkPropValue(p PropSpec, v any) string {
	switch p.Kind {
	case PropString, PropText, PropAsset, PropURL, PropSceneRef, PropIcon:
		if _, ok := v.(string); !ok {
			return fmt.Sprintf("property %q must be a string, got %T", p.Name, v)
		}
	case PropColor:
		s, ok := v.(string)
		if !ok {
			return fmt.Sprintf("property %q must be a color string, got %T", p.Name, v)
		}
		if !ValidColor(s) {
			return fmt.Sprintf("property %q has invalid color %q", p.Name, s)
		}
	case PropEnum:
		s, ok := v.(string)
		if !ok {
			return fmt.Sprintf("property %q must be a string, got %T", p.Name, v)
		}
		if len(p.Enum) > 0 && !contains(p.Enum, s) {
			return fmt.Sprintf("property %q must be one of %s, got %q", p.Name, strings.Join(p.Enum, ", "), s)
		}
	case PropNumber, PropInt:
		f, ok := toFloat(v)
		if !ok {
			return fmt.Sprintf("property %q must be a number, got %T", p.Name, v)
		}
		if p.Kind == PropInt && f != math.Trunc(f) {
			return fmt.Sprintf("property %q must be an integer, got %v", p.Name, f)
		}
		if p.HasMin && f < p.Min {
			return fmt.Sprintf("property %q must be >= %v, got %v", p.Name, p.Min, f)
		}
		if p.HasMax && f > p.Max {
			return fmt.Sprintf("property %q must be <= %v, got %v", p.Name, p.Max, f)
		}
	case PropBool:
		if _, ok := v.(bool); !ok {
			return fmt.Sprintf("property %q must be a bool, got %T", p.Name, v)
		}
	case PropPoints:
		pts, ok := v.([]any)
		if !ok {
			return fmt.Sprintf("property %q must be a list of points", p.Name)
		}
		if len(pts) > 16 {
			return fmt.Sprintf("property %q has %d points; the renderer supports at most 16", p.Name, len(pts))
		}
		for i, item := range pts {
			m, ok := item.(map[string]any)
			if !ok {
				return fmt.Sprintf("property %q point %d must be a {x,y} map", p.Name, i)
			}
			x, okx := toFloat(m["x"])
			y, oky := toFloat(m["y"])
			if !okx || !oky {
				return fmt.Sprintf("property %q point %d needs numeric x and y", p.Name, i)
			}
			if x < 0 || x > 1 || y < 0 || y > 1 {
				return fmt.Sprintf("property %q point %d is outside 0..1", p.Name, i)
			}
		}
	case PropOptions, PropTags:
		for _, s := range toStrings(v) {
			if strings.TrimSpace(s) == "" {
				return fmt.Sprintf("property %q contains an empty entry", p.Name)
			}
		}
		if _, ok := v.([]any); !ok {
			if _, ok := v.([]string); !ok {
				return fmt.Sprintf("property %q must be a list of strings, got %T", p.Name, v)
			}
		}
	}
	return ""
}

func checkEvents(n *Node, spec ComponentSpec, d *Document, add func(Severity, string, string, string, ...any)) {
	for ev, acts := range n.On {
		if !contains(Events, ev) {
			add(Warning, n.ID, "on."+ev, "unknown event %q", ev)
		} else if len(spec.Events) > 0 && !spec.Event(ev) {
			add(Warning, n.ID, "on."+ev, "component %q does not emit %q", spec.Type, ev)
		}
		for i, a := range acts {
			path := fmt.Sprintf("on.%s[%d]", ev, i)
			if a.Do == "" {
				add(Error, n.ID, path, "action has no verb")
				continue
			}
			if !contains(Verbs, a.Do) {
				add(Error, n.ID, path, "unknown action verb %q", a.Do)
				continue
			}
			switch a.Do {
			case "navigate":
				if strings.TrimSpace(a.To) == "" {
					add(Error, n.ID, path, "navigate needs a target URL")
				}
			case "call":
				if strings.TrimSpace(a.Handler) == "" {
					add(Error, n.ID, path, "call needs a handler name")
				}
			case "set":
				if strings.TrimSpace(a.Prop) == "" {
					add(Error, n.ID, path, "set needs a property name")
				}
			case "show", "hide", "toggle", "select":
				if a.Node != "" && a.Node != "self" {
					if _, ok := d.Find(a.Node); !ok {
						add(Error, n.ID, path, "target node %q does not exist", a.Node)
					}
				}
			}
		}
	}
}

func checkSlots(n *Node, spec ComponentSpec, add func(Severity, string, string, string, ...any)) {
	if spec.Kind != Slots {
		return
	}
	seen := make(map[string]bool)
	for i, c := range n.Children {
		s := c.Place.Slot
		if s == "" {
			continue // reported by the parent-kind check
		}
		if seen[s] {
			add(Error, c.ID, fmt.Sprintf("root/children[%d]", i), "slot %q is used by more than one child", s)
		}
		seen[s] = true
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	}
	return 0, false
}

func toStrings(v any) []string {
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

var namedSceneColors = map[string]bool{
	"black": true, "white": true, "red": true, "green": true, "blue": true,
	"yellow": true, "cyan": true, "magenta": true, "gray": true, "grey": true,
	"orange": true, "purple": true, "transparent": true,
}

// ValidColor accepts the palette shared with gui/canvas (names, #RGB,
// #RRGGBB, #RRGGBBAA) plus theme: references such as theme:foreground.
func ValidColor(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "theme:") {
		return len(strings.TrimPrefix(s, "theme:")) > 0
	}
	if namedSceneColors[s] {
		return true
	}
	hex := strings.TrimPrefix(s, "#")
	if len(hex) != 3 && len(hex) != 6 && len(hex) != 8 {
		return false
	}
	for _, r := range hex {
		if !(r >= '0' && r <= '9') && !(r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
