package objattrs

import (
	"sort"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// The registry catalog: every attr registry created via Register is
// listed here so tooling (`fynerisor docs`, issue #47) can enumerate
// the complete script attr surface from the compiled binary. Package
// vars initialize before any init(), so the catalog is full by the
// time a command reads it; All() sorts by type name for stable output.

// TypeSpecs names one registered type and its attribute surface.
type TypeSpecs struct {
	TypeName string
	Specs    []object.AttrSpec
}

var catalog []catalogEntry

type catalogEntry struct {
	typeName string
	registry interface{ Specs() []object.AttrSpec }
}

// Register adds r to the catalog under typeName and returns r, so
// wrapper files can declare
//
//	var labelAttrs = objattrs.Register(string(LabelType), object.NewAttrRegistry[*Label](string(LabelType)))
func Register[T any](typeName string, r *object.AttrRegistry[T]) *object.AttrRegistry[T] {
	catalog = append(catalog, catalogEntry{typeName, r})
	return r
}

// All returns every registered type's specs, sorted by type name.
// Types with no attributes are included (their Specs slice is empty).
func All() []TypeSpecs {
	out := make([]TypeSpecs, 0, len(catalog))
	for _, e := range catalog {
		out = append(out, TypeSpecs{TypeName: e.typeName, Specs: e.registry.Specs()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TypeName < out[j].TypeName })
	return out
}
