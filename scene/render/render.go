// Package render turns scene documents into live Fyne object trees. It is the
// "player" half of fynerisor's scene support; the designer edits the document
// that this package renders.
//
// The runtime is deliberately independent of Risor: a scene without a script is
// a fully functional page. gui/scene adds the script bindings on top.
package render

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"github.com/uidbz/fynerisor/internal/colors"
	"github.com/uidbz/fynerisor/scene"
)

// Mode selects how a scene is rendered.
type Mode int

const (
	// ModePlay is the normal interactive page. Hotspots are invisible unless
	// they set showInPlay.
	ModePlay Mode = iota
	// ModeEdit shows edit affordances (hotspot outlines, script slot hints)
	// but does not itself handle input; the designer's overlay does.
	ModeEdit
)

// ErrRebuild asks the View to rebuild the node's live subtree (for properties
// that cannot be updated in place, such as a theme style change).
var ErrRebuild = errors.New("scene/render: rebuild required")

// Live is one rendered scene node.
type Live interface {
	// Object is the Fyne object inserted into the parent container.
	Object() fyne.CanvasObject
	// Update applies changed node properties. Returning ErrRebuild makes the
	// View replace this live subtree.
	Update(n *scene.Node, changed scene.Fields) error
	// Destroy releases resources (media, pending loads).
	Destroy()
}

// ContainerLive is implemented by components with children.
type ContainerLive interface {
	Live
	Insert(index int, child Live, p scene.Place) error
	Remove(child Live) error
	Reorder(child Live, index int) error
	SetPlace(child Live, p scene.Place) error
	// Preview shows a temporary placement during an editor drag. It must not
	// change the document.
	Preview(child Live, p scene.Place)
	// DropSpot maps a position in the container's local coordinate space to an
	// insertion point.
	DropSpot(local fyne.Position) (DropSpot, bool)
	// Transform maps design units to local units inside this container.
	Transform(size fyne.Size) Transform
}

// DropSpot describes where a dragged node would land.
type DropSpot struct {
	Index int
	Slot  string
	Rect  Rect // highlight geometry in the container's local space
}

// Transform maps design units to local pixels.
type Transform struct {
	SX, SY float64
	OX, OY float64
}

// ToDesign converts a local point into design units.
func (t Transform) ToDesign(p fyne.Position) (x, y float64) {
	if t.SX == 0 {
		t.SX = 1
	}
	if t.SY == 0 {
		t.SY = 1
	}
	return (float64(p.X) - t.OX) / t.SX, (float64(p.Y) - t.OY) / t.SY
}

// ToLocal converts design units into a local point.
func (t Transform) ToLocal(x, y float64) fyne.Position {
	if t.SX == 0 {
		t.SX = 1
	}
	if t.SY == 0 {
		t.SY = 1
	}
	return fyne.NewPos(float32(t.OX+x*t.SX), float32(t.OY+y*t.SY))
}

// AssetResolver resolves and loads scene assets. Resolve is cheap and may be
// called on the UI thread; Open does I/O and must be called off the UI thread.
type AssetResolver interface {
	Resolve(base, ref string) (string, error)
	Open(ctx context.Context, ref string) (fyne.Resource, error)
}

// EmbedResolver loads another scene document for an embed node. It runs off
// the UI thread. The returned base is the embedded document's own URL/path,
// used for relative links and assets inside it.
type EmbedResolver interface {
	LoadEmbed(ctx context.Context, base, ref string) (*scene.Document, string, error)
}

// Navigator receives declarative navigation actions.
type Navigator interface {
	Navigate(ref string)
	Back()
}

// Opener receives open/play/pause actions (media, external URLs, documents).
type Opener interface {
	Open(ref string)
	Play(ref string)
	Pause()
}

// Caller invokes a script handler bound with page.Bind. It must not block the
// UI thread: implementations queue the call on the script's VM goroutine.
type Caller interface {
	Call(handler string, evt map[string]any) bool
}

// BuildContext is passed to component factories.
type BuildContext struct {
	Context context.Context
	Mode    Mode
	Base    string
	Assets  AssetResolver
	Embed   EmbedResolver
	Status  func(string)
	View    *View
	emit    func(id, event string, payload map[string]any)
}

// Emit reports a component event. Components call it from Fyne callbacks,
// which run on the UI thread.
func (bc *BuildContext) Emit(id, event string, payload map[string]any) {
	if bc.emit != nil {
		bc.emit(id, event, payload)
	}
}

// Statusf reports a non-fatal runtime message.
func (bc *BuildContext) Statusf(format string, args ...any) {
	if bc.Status != nil {
		bc.Status(fmt.Sprintf(format, args...))
	}
}

// Factory builds a component's live object. Children are inserted by View.
type Factory func(bc *BuildContext, n *scene.Node) (Live, error)

// Registry maps scene component types to factories. Specs are shared with the
// scene package so validation and rendering cannot drift.
type Registry struct {
	*scene.Registry
	factories map[string]Factory
}

// NewRegistry returns an empty render registry backed by base. A nil base
// creates an empty scene registry.
func NewRegistry(base *scene.Registry) *Registry {
	if base == nil {
		base = scene.NewRegistry()
	}
	return &Registry{Registry: base, factories: make(map[string]Factory)}
}

// Register adds a factory for a spec.
func (r *Registry) Register(spec scene.ComponentSpec, f Factory) {
	r.Registry.Register(spec)
	r.factories[spec.Type] = f
}

// Factory adds a factory for a type whose spec is already registered. It is
// how scene/render/builtin binds Fyne implementations to the pure-data specs
// without registering them twice.
func (r *Registry) Factory(typ string, f Factory) {
	if _, ok := r.Registry.Spec(typ); !ok {
		panic(fmt.Sprintf("scene/render: factory for unregistered component %q", typ))
	}
	r.factories[typ] = f
}

// Build creates a live component for n. Unknown types become a visible
// placeholder so forward-compatible documents degrade instead of failing.
func (r *Registry) Build(bc *BuildContext, n *scene.Node) (Live, error) {
	if r == nil {
		return nil, fmt.Errorf("scene/render: nil registry")
	}
	if f, ok := r.factories[n.Type]; ok {
		l, err := f(bc, n)
		if err != nil {
			return nil, fmt.Errorf("%s(%s): %w", n.Type, n.ID, err)
		}
		return l, nil
	}
	if _, known := r.Registry.Spec(n.Type); known {
		return nil, fmt.Errorf("scene/render: component %q has no factory", n.Type)
	}
	return newPlaceholder(n), nil
}

// Color parses a scene color: a name, a hex value, or theme:<name>.
func Color(s string, fallback color.Color) color.Color {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	if name, ok := strings.CutPrefix(s, "theme:"); ok {
		if c := themeColor(name); c != nil {
			return c
		}
		return fallback
	}
	c, err := colors.Parse(s)
	if err != nil {
		return fallback
	}
	return c
}

var themeColors = map[string]fyne.ThemeColorName{
	"background":          theme.ColorNameBackground,
	"foreground":          theme.ColorNameForeground,
	"primary":             theme.ColorNamePrimary,
	"button":              theme.ColorNameButton,
	"disabled":            theme.ColorNameDisabled,
	"disabledbutton":      theme.ColorNameDisabledButton,
	"error":               theme.ColorNameError,
	"focus":               theme.ColorNameFocus,
	"hover":               theme.ColorNameHover,
	"inputbackground":     theme.ColorNameInputBackground,
	"inputborder":         theme.ColorNameInputBorder,
	"menubackground":      theme.ColorNameMenuBackground,
	"overlaybackground":   theme.ColorNameOverlayBackground,
	"placeholder":         theme.ColorNamePlaceHolder,
	"pressed":             theme.ColorNamePressed,
	"scrollbar":           theme.ColorNameScrollBar,
	"separator":           theme.ColorNameSeparator,
	"shadow":              theme.ColorNameShadow,
	"success":             theme.ColorNameSuccess,
	"warning":             theme.ColorNameWarning,
	"hyperlink":           theme.ColorNameHyperlink,
	"selection":           theme.ColorNameSelection,
	"header":              theme.ColorNameHeaderBackground,
	"foregroundonprimary": theme.ColorNameForegroundOnPrimary,
}

func themeColor(name string) color.Color {
	app := fyne.CurrentApp()
	if app == nil || app.Settings() == nil || app.Settings().Theme() == nil {
		return nil
	}
	cn, ok := themeColors[strings.ToLower(name)]
	if !ok {
		return nil
	}
	return app.Settings().Theme().Color(cn, app.Settings().ThemeVariant())
}

// ColorHex renders a color for editors and documents.
func ColorHex(c color.Color) string {
	if c == nil {
		return ""
	}
	return colors.Hex(c)
}
