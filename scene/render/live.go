package render

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
)

// hintObject carries a minimum-size hint for a node placed inside a layout
// container. Free placement ignores it (the place is authoritative).
type hintObject struct {
	fyne.CanvasObject
	minW, minH float32
}

// MinSize returns the child's minimum size enlarged by the scene hint.
func (h *hintObject) MinSize() fyne.Size {
	m := h.CanvasObject.MinSize()
	if h.minW > m.Width {
		m.Width = h.minW
	}
	if h.minH > m.Height {
		m.Height = h.minH
	}
	return m
}

// SetScale forwards the free-layout scale to objects that care (canvas text).
func (h *hintObject) SetScale(sx, sy float64) {
	if s, ok := h.CanvasObject.(interface{ SetScale(sx, sy float64) }); ok {
		s.SetScale(sx, sy)
	}
}

type hintedLive struct {
	Live
	hint *hintObject
}

func (h *hintedLive) Object() fyne.CanvasObject { return h.hint }

// hintedContainer applies place.minW/minH to a container's object while
// keeping the ContainerLive contract intact.
type hintedContainer struct {
	ContainerLive
	hint *hintObject
}

func (h *hintedContainer) Object() fyne.CanvasObject { return h.hint }

// FreeLayout forwards the layout of a free container through the hint wrapper.
func (h *hintedContainer) FreeLayout() *FreeLayout {
	if f, ok := h.ContainerLive.(interface{ FreeLayout() *FreeLayout }); ok {
		return f.FreeLayout()
	}
	return nil
}

// wrapHint wraps lives so placed min sizes work for leaves and containers.
func wrapHint(l Live) Live {
	if c, ok := l.(ContainerLive); ok {
		return &hintedContainer{ContainerLive: c, hint: &hintObject{CanvasObject: l.Object()}}
	}
	return &hintedLive{Live: l, hint: &hintObject{CanvasObject: l.Object()}}
}

// SetSizeHint updates the minimum size of a node's object. It is a no-op for
// objects without a hint wrapper (containers).
func SetSizeHint(o fyne.CanvasObject, minW, minH float64) {
	if hint, ok := o.(*hintObject); ok {
		hint.minW, hint.minH = float32(minW), float32(minH)
	}
}

// Prop returns a property value, falling back to the component spec's default.
func (bc *BuildContext) Prop(n *scene.Node, key string) any {
	if v, ok := n.Prop(key); ok {
		return v
	}
	if bc != nil && bc.View != nil && bc.View.Registry() != nil {
		if spec, ok := bc.View.Registry().Spec(n.Type); ok {
			if p, ok := spec.Prop(key); ok {
				return p.Default
			}
		}
	}
	return nil
}

// Str returns a string property.
func (bc *BuildContext) Str(n *scene.Node, key, def string) string {
	if s, ok := bc.Prop(n, key).(string); ok {
		return s
	}
	return def
}

// Num returns a numeric property.
func (bc *BuildContext) Num(n *scene.Node, key string, def float64) float64 {
	return float64Of(bc.Prop(n, key), def)
}

// Bool returns a boolean property.
func (bc *BuildContext) Bool(n *scene.Node, key string, def bool) bool {
	if b, ok := bc.Prop(n, key).(bool); ok {
		return b
	}
	return def
}

// List returns a list-of-strings property.
func (bc *BuildContext) List(n *scene.Node, key string) []string {
	return listOf(bc.Prop(n, key))
}

// ColorProp returns a color property.
func (bc *BuildContext) ColorProp(n *scene.Node, key string, def color.Color) color.Color {
	if s, ok := bc.Prop(n, key).(string); ok {
		return Color(s, def)
	}
	return def
}

func float64Of(v any, def float64) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	}
	return def
}

func listOf(v any) []string {
	switch x := v.(type) {
	case []string:
		return append([]string(nil), x...)
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

// TextAlign maps scene alignment names to Fyne.
func TextAlign(name string) fyne.TextAlign {
	switch strings.ToLower(name) {
	case "center":
		return fyne.TextAlignCenter
	case "right":
		return fyne.TextAlignTrailing
	default:
		return fyne.TextAlignLeading
	}
}

// TextWrap maps scene wrapping names to Fyne.
func TextWrap(name string) fyne.TextWrap {
	switch strings.ToLower(name) {
	case "off":
		return fyne.TextWrapOff
	case "break":
		return fyne.TextWrapBreak
	default:
		return fyne.TextWrapWord
	}
}

// TextTruncate maps scene truncation names to Fyne.
func TextTruncate(name string) fyne.TextTruncation {
	if strings.EqualFold(name, "ellipsis") {
		return fyne.TextTruncateEllipsis
	}
	return fyne.TextTruncateOff
}

// FillMode maps scene image fill names to Fyne.
func FillMode(name string) canvas.ImageFill {
	switch strings.ToLower(name) {
	case "stretch":
		return canvas.ImageFillStretch
	case "contain":
		return canvas.ImageFillContain
	case "original":
		return canvas.ImageFillOriginal
	default:
		return canvas.ImageFillCover
	}
}

// Importance maps scene button importance names to Fyne.
func Importance(name string) widget.Importance {
	switch strings.ToLower(name) {
	case "low":
		return widget.LowImportance
	case "high":
		return widget.HighImportance
	case "danger":
		return widget.DangerImportance
	case "warning":
		return widget.WarningImportance
	case "success":
		return widget.SuccessImportance
	default:
		return widget.MediumImportance
	}
}

// IconResource resolves a scene icon name to a theme resource.
func IconResource(name string) fyne.Resource {
	switch strings.ToLower(name) {
	case "home":
		return theme.HomeIcon()
	case "folder":
		return theme.FolderIcon()
	case "settings":
		return theme.SettingsIcon()
	case "search":
		return theme.SearchIcon()
	case "image", "photo":
		return theme.FileImageIcon()
	case "audio", "music":
		return theme.FileAudioIcon()
	case "video":
		return theme.FileVideoIcon()
	case "document", "file":
		return theme.FileTextIcon()
	case "delete":
		return theme.DeleteIcon()
	case "add":
		return theme.ContentAddIcon()
	case "back":
		return theme.NavigateBackIcon()
	case "next":
		return theme.NavigateNextIcon()
	case "menu":
		return theme.MenuIcon()
	case "warning":
		return theme.WarningIcon()
	case "error":
		return theme.ErrorIcon()
	case "help":
		return theme.HelpIcon()
	case "mail":
		return theme.MailComposeIcon()
	case "account":
		return theme.AccountIcon()
	case "calendar":
		return theme.CalendarIcon()
	case "star":
		return theme.InfoIcon()
	case "play":
		return theme.MediaPlayIcon()
	case "pause":
		return theme.MediaPauseIcon()
	default:
		return theme.InfoIcon()
	}
}

// placeholder renders unknown component types without losing their data. A
// newer page opened in an older browser degrades to this box instead of
// failing to load.
type placeholder struct {
	box   *fyne.Container
	label *widget.Label
	node  string
}

func newPlaceholder(n *scene.Node) *placeholder {
	l := widget.NewLabel("")
	p := &placeholder{label: l, node: n.ID}
	rect := canvas.NewRectangle(color.NRGBA{R: 160, G: 90, B: 220, A: 48})
	rect.StrokeColor = color.NRGBA{R: 160, G: 90, B: 220, A: 200}
	rect.StrokeWidth = 2
	p.box = container.NewStack(rect, container.NewCenter(l))
	p.set(n)
	return p
}

func (p *placeholder) set(n *scene.Node) {
	p.label.SetText(fmt.Sprintf("%s\n%s", n.Type, n.ID))
}

func (p *placeholder) Object() fyne.CanvasObject { return p.box }

func (p *placeholder) Update(n *scene.Node, _ scene.Fields) error {
	p.set(n)
	return nil
}

func (p *placeholder) Destroy() {}
