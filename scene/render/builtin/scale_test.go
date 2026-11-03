package builtin

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestScaleThemeSizes(t *testing.T) {
	base := theme.DefaultTheme()
	th := &scaleTheme{base: base, scale: 0.5}
	for _, name := range []fyne.ThemeSizeName{
		theme.SizeNameText, theme.SizeNameHeadingText, theme.SizeNameSubHeadingText,
		theme.SizeNameCaptionText, theme.SizeNamePadding, theme.SizeNameInnerPadding,
		theme.SizeNameInlineIcon,
	} {
		if got, want := th.Size(name), base.Size(name)*0.5; got != want {
			t.Errorf("Size(%v) = %v, want %v", name, got, want)
		}
	}
	// Sizes that must not shrink with the canvas.
	for _, name := range []fyne.ThemeSizeName{
		theme.SizeNameInputRadius, theme.SizeNameSelectionRadius,
	} {
		if got, want := th.Size(name), base.Size(name); got != want {
			t.Errorf("Size(%v) = %v, want %v (unscaled)", name, got, want)
		}
	}
}

func TestScaleThemeOneIsIdentity(t *testing.T) {
	base := theme.DefaultTheme()
	th := &scaleTheme{base: base, scale: 1}
	if got, want := th.Size(theme.SizeNameText), base.Size(theme.SizeNameText); got != want {
		t.Fatalf("scale 1 changed text size: %v != %v", got, want)
	}
}
