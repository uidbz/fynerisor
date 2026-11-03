// Package colors parses the color names and hex forms shared by fynerisor's
// canvas objects and the scene renderer. It has no Fyne dependency so scene
// validation and headless tools can use the same palette.
package colors

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// Named is the small named palette used by scripts and scenes.
var Named = map[string]color.Color{
	"red":         color.RGBA{R: 255, A: 255},
	"green":       color.RGBA{G: 128, A: 255},
	"blue":        color.RGBA{B: 255, A: 255},
	"black":       color.RGBA{A: 255},
	"white":       color.RGBA{R: 255, G: 255, B: 255, A: 255},
	"yellow":      color.RGBA{R: 255, G: 255, A: 255},
	"cyan":        color.RGBA{G: 255, B: 255, A: 255},
	"magenta":     color.RGBA{R: 255, B: 255, A: 255},
	"orange":      color.RGBA{R: 255, G: 165, A: 255},
	"purple":      color.RGBA{R: 128, G: 0, B: 128, A: 255},
	"gray":        color.RGBA{R: 128, G: 128, B: 128, A: 255},
	"grey":        color.RGBA{R: 128, G: 128, B: 128, A: 255},
	"transparent": color.RGBA{},
}

// Parse accepts a name, #RGB, #RRGGBB or #RRGGBBAA.
func Parse(s string) (color.Color, error) {
	name := strings.ToLower(strings.TrimSpace(s))
	if c, ok := Named[name]; ok {
		return c, nil
	}
	hex := strings.TrimPrefix(name, "#")
	switch len(hex) {
	case 3:
		hex = string(hex[0]) + string(hex[0]) +
			string(hex[1]) + string(hex[1]) +
			string(hex[2]) + string(hex[2]) + "ff"
	case 6:
		hex += "ff"
	case 8:
	default:
		return nil, fmt.Errorf("invalid color %q: use a name or #RGB/#RRGGBB/#RRGGBBAA", s)
	}
	var parts [4]uint8
	for i := 0; i < 4; i++ {
		v, err := strconv.ParseUint(hex[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid color %q: %v", s, err)
		}
		parts[i] = uint8(v)
	}
	return color.NRGBA{R: parts[0], G: parts[1], B: parts[2], A: parts[3]}, nil
}

// Hex returns #RRGGBBAA for a color.
func Hex(c color.Color) string {
	r, g, b, a := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
}
