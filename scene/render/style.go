package render

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	"github.com/uidbz/fynerisor/scene"
)

// styleTheme applies a scene node's Style overrides to its subtree through
// Fyne's ThemeOverride container.
type styleTheme struct {
	base fyne.Theme
	st   scene.Style
}

func newStyleTheme(st *scene.Style) fyne.Theme {
	base := theme.DefaultTheme()
	if app := fyne.CurrentApp(); app != nil && app.Settings() != nil && app.Settings().Theme() != nil {
		base = app.Settings().Theme()
	}
	return &styleTheme{base: base, st: *st}
}

func (t *styleTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.st.Bg != "" {
		switch name {
		case theme.ColorNameBackground, theme.ColorNameInputBackground,
			theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground,
			theme.ColorNameButton:
			return Color(t.st.Bg, t.base.Color(name, variant))
		}
	}
	if t.st.Fg != "" {
		switch name {
		case theme.ColorNameForeground, theme.ColorNamePlaceHolder,
			theme.ColorNameDisabled, theme.ColorNameHyperlink:
			return Color(t.st.Fg, t.base.Color(name, variant))
		}
	}
	return t.base.Color(name, variant)
}

func (t *styleTheme) Font(style fyne.TextStyle) fyne.Resource {
	style.Bold = style.Bold || t.st.Bold
	style.Italic = style.Italic || t.st.Italic
	return t.base.Font(style)
}

func (t *styleTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *styleTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText && t.st.TextSize > 0 {
		return float32(t.st.TextSize)
	}
	return t.base.Size(name)
}

type styledLive struct {
	Live
	obj fyne.CanvasObject
}

func (s *styledLive) Object() fyne.CanvasObject { return s.obj }

type styledContainer struct {
	ContainerLive
	obj fyne.CanvasObject
}

func (s *styledContainer) Object() fyne.CanvasObject { return s.obj }

func (s *styledContainer) FreeLayout() *FreeLayout {
	if f, ok := s.ContainerLive.(interface{ FreeLayout() *FreeLayout }); ok {
		return f.FreeLayout()
	}
	return nil
}

func (s *styledContainer) DirectLayout() bool {
	if d, ok := s.ContainerLive.(interface{ DirectLayout() bool }); ok {
		return d.DirectLayout()
	}
	return false
}

// wrapStyle applies a node's style to its whole subtree. It runs after the
// children are inserted so the ThemeOverride sees the complete tree.
func wrapStyle(l Live, st *scene.Style) Live {
	if st == nil || st.IsZero() {
		return l
	}
	over := container.NewThemeOverride(l.Object(), newStyleTheme(st))
	if c, ok := l.(ContainerLive); ok {
		return &styledContainer{ContainerLive: c, obj: over}
	}
	return &styledLive{Live: l, obj: over}
}
