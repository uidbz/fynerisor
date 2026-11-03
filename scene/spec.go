package scene

import (
	"fmt"
	"sort"
	"sync"
)

// ContainerKind describes how a component's children are placed.
type ContainerKind string

const (
	// Leaf components have no children.
	Leaf ContainerKind = "leaf"
	// Free components place children at x, y, w, h in design units.
	Free ContainerKind = "free"
	// Ordered components lay children out in document order (VBox, Grid, ...).
	Ordered ContainerKind = "ordered"
	// Slots components assign children to named slots (Border, Split, ...).
	Slots ContainerKind = "slots"
	// Tabs components show one child per tab (AppTabs, Accordion).
	Tabs ContainerKind = "tabs"
)

// PropKind drives validation and the generated inspector editor.
type PropKind string

const (
	PropString   PropKind = "string"
	PropText     PropKind = "text"
	PropNumber   PropKind = "number"
	PropInt      PropKind = "int"
	PropBool     PropKind = "bool"
	PropColor    PropKind = "color"
	PropEnum     PropKind = "enum"
	PropAsset    PropKind = "asset"
	PropURL      PropKind = "url"
	PropSceneRef PropKind = "sceneRef"
	PropNodeRef  PropKind = "nodeRef"
	PropIcon     PropKind = "icon"
	PropOptions  PropKind = "options" // list of strings
	PropTags     PropKind = "tags"    // list of strings, tag-search semantics
	PropPoints   PropKind = "points"  // future: polygon hotspots
)

// PropSpec describes one script- and inspector-visible property.
type PropSpec struct {
	Name     string
	Title    string
	Group    string // inspector group, e.g. "Content", "Layout", "Look", "Data"
	Kind     PropKind
	Default  any
	Enum     []string
	Min, Max float64
	HasMin   bool
	HasMax   bool
	Help     string
	Advanced bool
}

// EventSpec documents an event a component can emit.
type EventSpec struct {
	Name string
	Doc  string
}

// ComponentSpec is the plain-data contract shared by validation, the renderer,
// the inspector, documentation and tests. It carries no Fyne or Risor types.
type ComponentSpec struct {
	Type         string
	Title        string
	Category     string
	Icon         string // theme icon name, resolved by the renderer
	Doc          string
	Kind         ContainerKind
	MaxChildren  int // 0 = unlimited
	DefaultPlace Place
	DefaultProps map[string]any // optional seed for new nodes
	Props        []PropSpec
	Events       []EventSpec
	Slots        []string
}

// Prop returns the spec of a named property.
func (s ComponentSpec) Prop(name string) (PropSpec, bool) {
	for _, p := range s.Props {
		if p.Name == name {
			return p, true
		}
	}
	return PropSpec{}, false
}

// Event reports whether the component documents the event.
func (s ComponentSpec) Event(name string) bool {
	for _, e := range s.Events {
		if e.Name == name {
			return true
		}
	}
	return false
}

// Default returns a property's default value, or nil.
func (s ComponentSpec) Default(name string) any {
	if p, ok := s.Prop(name); ok {
		return p.Default
	}
	return nil
}

// Registry is an ordered set of component specs. It is safe for concurrent
// reads after construction. The scene package registers no specs itself;
// scene/builtin provides the standard set and scene/render adds factories.
type Registry struct {
	mu    sync.RWMutex
	types []string
	specs map[string]ComponentSpec
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{specs: make(map[string]ComponentSpec)}
}

// Register adds a spec. Registering the same type twice panics: component
// types are process-wide identifiers and a silent override hides bugs.
func (r *Registry) Register(spec ComponentSpec) {
	if spec.Type == "" {
		panic("scene: component spec with empty type")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.specs[spec.Type]; ok {
		panic(fmt.Sprintf("scene: component %q registered twice", spec.Type))
	}
	r.specs[spec.Type] = spec
	r.types = append(r.types, spec.Type)
}

// Spec returns a component spec.
func (r *Registry) Spec(typ string) (ComponentSpec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.specs[typ]
	return s, ok
}

// Types returns the registered types in registration order.
func (r *Registry) Types() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]string(nil), r.types...)
}

// Len returns the number of registered components.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.types)
}

// SortedTypes returns the registered types sorted alphabetically.
func (r *Registry) SortedTypes() []string {
	out := r.Types()
	sort.Strings(out)
	return out
}

// Kind returns a component's container kind and whether it is known.
func (r *Registry) Kind(typ string) (ContainerKind, bool) {
	s, ok := r.Spec(typ)
	if !ok {
		return Leaf, false
	}
	return s.Kind, true
}

// Props returns a component's property specs.
func (r *Registry) Props(typ string) []PropSpec {
	s, ok := r.Spec(typ)
	if !ok {
		return nil
	}
	return s.Props
}

// Events returns a component's event names.
func (r *Registry) Events(typ string) []string {
	s, ok := r.Spec(typ)
	if !ok {
		return nil
	}
	out := make([]string, len(s.Events))
	for i, e := range s.Events {
		out[i] = e.Name
	}
	return out
}

// Slots returns a component's slot names.
func (r *Registry) Slots(typ string) []string {
	s, ok := r.Spec(typ)
	if !ok {
		return nil
	}
	return s.Slots
}
