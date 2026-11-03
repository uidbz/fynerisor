package designer

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// zoomTheme scales text sizes for device/zoom previews. It is applied around
// the view only, so the editor chrome stays at its normal size.
type zoomTheme struct {
	base  fyne.Theme
	scale float64
}

func newZoomTheme(scale float64) fyne.Theme {
	base := theme.DefaultTheme()
	if app := fyne.CurrentApp(); app != nil && app.Settings() != nil && app.Settings().Theme() != nil {
		base = app.Settings().Theme()
	}
	return &zoomTheme{base: base, scale: scale}
}

func (t *zoomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}
func (t *zoomTheme) Font(style fyne.TextStyle) fyne.Resource    { return t.base.Font(style) }
func (t *zoomTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }
func (t *zoomTheme) Size(name fyne.ThemeSizeName) float32 {
	size := t.base.Size(name)
	if t.scale <= 0 {
		return size
	}
	switch name {
	case theme.SizeNameText, theme.SizeNameHeadingText, theme.SizeNameSubHeadingText,
		theme.SizeNameCaptionText, theme.SizeNamePadding, theme.SizeNameInnerPadding,
		theme.SizeNameInlineIcon, theme.SizeNameScrollBar, theme.SizeNameScrollBarSmall,
		theme.SizeNameSeparatorThickness, theme.SizeNameInputBorder:
		return size * float32(t.scale)
	}
	return size
}
