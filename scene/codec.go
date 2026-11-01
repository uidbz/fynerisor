package scene

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// MaxDocumentBytes bounds untrusted scene documents (2 MiB).
	MaxDocumentBytes = 2 << 20
	// MaxNodes bounds the node count of untrusted documents.
	MaxNodes = 5000
	// MaxDepth bounds nesting of untrusted documents.
	MaxDepth = 32
)

// Decode parses and structurally checks a scene document. It does not validate
// component semantics; call Validate with a registry for that.
func Decode(data []byte) (*Document, error) {
	if len(data) > MaxDocumentBytes {
		return nil, fmt.Errorf("scene: document is %d bytes, limit is %d", len(data), MaxDocumentBytes)
	}
	var d Document
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("scene: parse: %w", err)
	}
	if d.Format == "" {
		d.Format = Format // tolerate hand-written drafts; Encode fixes the field
	}
	if d.Format != Format {
		if strings.HasPrefix(d.Format, "fynerisor.scene/") {
			return nil, fmt.Errorf("scene: unsupported format %q (this build supports %q)", d.Format, Format)
		}
		return nil, fmt.Errorf("scene: not a scene document (format %q)", d.Format)
	}
	if d.Root == nil {
		return nil, fmt.Errorf("scene: document has no root node")
	}
	if n := d.Count(); n > MaxNodes {
		return nil, fmt.Errorf("scene: %d nodes, limit is %d", n, MaxNodes)
	}
	if depth := d.Depth(); depth > MaxDepth {
		return nil, fmt.Errorf("scene: depth %d, limit is %d", depth, MaxDepth)
	}
	seen := make(map[string]bool)
	var firstErr error
	d.Walk(func(n *Node) bool {
		if n.ID == "" {
			firstErr = fmt.Errorf("scene: node of type %q has no id", n.Type)
			return false
		}
		if !ValidID(n.ID) {
			firstErr = fmt.Errorf("scene: invalid node id %q (use lowercase letters, digits and dashes)", n.ID)
			return false
		}
		if seen[n.ID] {
			firstErr = fmt.Errorf("scene: duplicate node id %q", n.ID)
			return false
		}
		seen[n.ID] = true
		return true
	})
	if firstErr != nil {
		return nil, firstErr
	}
	return &d, nil
}

// DecodeValidated decodes and validates against a schema.
func DecodeValidated(data []byte, schema Schema) (*Document, []Issue, error) {
	d, err := Decode(data)
	if err != nil {
		return nil, nil, err
	}
	return d, Validate(d, schema), nil
}

// Encode writes the canonical representation: sorted map keys, fixed field
// order, two-space indentation, values rounded to 0.01 and a trailing newline.
// The same document always encodes to the same bytes, which makes diffs stable
// and content-addressed storage meaningful.
func Encode(d *Document) ([]byte, error) {
	if d == nil || d.Root == nil {
		return nil, fmt.Errorf("scene: cannot encode a document without a root")
	}
	c := d.Clone()
	c.Format = Format
	normalizeDocument(c)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		return nil, fmt.Errorf("scene: encode: %w", err)
	}
	return buf.Bytes(), nil
}

func normalizeDocument(d *Document) {
	d.Each(func(n, _ *Node, _ int, _ int) bool {
		n.Place.X = Round(n.Place.X)
		n.Place.Y = Round(n.Place.Y)
		n.Place.W = Round(n.Place.W)
		n.Place.H = Round(n.Place.H)
		n.Place.MinW = Round(n.Place.MinW)
		n.Place.MinH = Round(n.Place.MinH)
		if n.Style != nil {
			n.Style.TextSize = Round(n.Style.TextSize)
			if n.Style.IsZero() {
				n.Style = nil
			}
		}
		if n.Props != nil {
			n.Props = roundMap(n.Props)
		}
		for ev, acts := range n.On {
			for i := range acts {
				if acts[i].Value != nil {
					acts[i].Value = roundValue(acts[i].Value)
				}
				if acts[i].Args != nil {
					acts[i].Args = roundMap(acts[i].Args)
				}
			}
			n.On[ev] = acts
		}
		return true
	})
}

func roundMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = roundValue(v)
	}
	return out
}

func roundValue(v any) any {
	switch x := v.(type) {
	case float64:
		return Round(x)
	case float32:
		return Round(float64(x))
	case map[string]any:
		return roundMap(x)
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = roundValue(e)
		}
		return out
	default:
		return v
	}
}
