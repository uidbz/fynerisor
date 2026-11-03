// Package builtin defines the standard component specs: the plain-data
// contract for layout, display, input and navigation components. It has no
// Fyne dependency, so validation, documentation and the CLI can use it
// without a GUI build. The matching factories live in scene/render/builtin.
package builtin

import "github.com/uidbz/fynerisor/scene"

// Register adds the standard components to r.
func Register(r *scene.Registry) {
	for _, s := range Specs() {
		r.Register(s)
	}
}

// Registry returns a new registry containing the standard components.
func Registry() *scene.Registry {
	r := scene.NewRegistry()
	Register(r)
	return r
}

func p(name, title, group string, kind scene.PropKind, def any) scene.PropSpec {
	return scene.PropSpec{Name: name, Title: title, Group: group, Kind: kind, Default: def}
}

func enum(name, title, group, def string, values ...string) scene.PropSpec {
	return scene.PropSpec{Name: name, Title: title, Group: group, Kind: scene.PropEnum, Default: def, Enum: values}
}

func num(name, title, group string, def, min, max float64) scene.PropSpec {
	return scene.PropSpec{
		Name: name, Title: title, Group: group, Kind: scene.PropNumber,
		Default: def, Min: min, Max: max, HasMin: true, HasMax: true,
	}
}

func ev(name, doc string) scene.EventSpec { return scene.EventSpec{Name: name, Doc: doc} }

var (
	evTapped    = ev("tapped", "A primary tap or click.")
	evDouble    = ev("doubleTapped", "A double tap.")
	evSecondary = ev("secondaryTapped", "A right click or long press.")
	evHoverIn   = ev("hoverIn", "The pointer entered the component.")
	evHoverOut  = ev("hoverOut", "The pointer left the component.")
	evChanged   = ev("changed", "The value changed.")
	evSubmitted = ev("submitted", "The value was submitted.")
)

func defaultPlace(w, h float64) scene.Place { return scene.Place{W: w, H: h} }

// Specs returns the standard component specs.
func Specs() []scene.ComponentSpec {
	return []scene.ComponentSpec{
		// ---------------------------------------------------------------- layout
		{
			Type: "free", Title: "Free canvas", Category: "Layout", Icon: "grid",
			Doc:  "Absolute x/y placement. Hotspots and media keep their position on the background at any window size when scale is fit or fill.",
			Kind: scene.Free, DefaultPlace: defaultPlace(800, 600),
			Props: []scene.PropSpec{
				num("designW", "Design width", "Layout", 1600, 1, 100000),
				num("designH", "Design height", "Layout", 900, 1, 100000),
				enum("scale", "Scale", "Layout", "fit", "none", "fit", "fill", "stretch"),
				p("background", "Background", "Content", scene.PropAsset, ""),
				enum("backgroundFill", "Background fill", "Content", "stretch", "stretch", "contain", "cover", "original"),
				p("letterbox", "Letterbox color", "Look", scene.PropColor, "#000000"),
				p("backgroundColor", "Design background", "Look", scene.PropColor, ""),
				p("clip", "Clip children", "Layout", scene.PropBool, false),
				p("scaleText", "Scale text with canvas", "Layout", scene.PropBool, true),
			},
		},
		{
			Type: "vbox", Title: "Vertical box", Category: "Layout", Icon: "list",
			Doc: "Children stacked top to bottom in document order.", Kind: scene.Ordered,
			DefaultPlace: defaultPlace(240, 200),
		},
		{
			Type: "hbox", Title: "Horizontal box", Category: "Layout", Icon: "list",
			Doc: "Children laid out left to right in document order.", Kind: scene.Ordered,
			DefaultPlace: defaultPlace(320, 80),
		},
		{
			Type: "grid", Title: "Grid", Category: "Layout", Icon: "grid",
			Doc: "Children in a fixed number of columns.", Kind: scene.Ordered,
			DefaultPlace: defaultPlace(320, 240),
			Props:        []scene.PropSpec{num("cols", "Columns", "Layout", 2, 1, 32)},
		},
		{
			Type: "gridwrap", Title: "Grid wrap", Category: "Layout", Icon: "grid",
			Doc: "Children wrapped into rows at their minimum size.", Kind: scene.Ordered,
			DefaultPlace: defaultPlace(320, 240),
		},
		{
			Type: "stack", Title: "Stack", Category: "Layout", Icon: "layers",
			Doc: "All children overlap; later children are drawn on top.", Kind: scene.Ordered,
			DefaultPlace: defaultPlace(320, 240),
		},
		{
			Type: "center", Title: "Center", Category: "Layout", Icon: "center",
			Doc: "Children centered and overlapping.", Kind: scene.Ordered,
			DefaultPlace: defaultPlace(320, 240),
		},
		{
			Type: "padded", Title: "Padded", Category: "Layout", Icon: "padding",
			Doc: "Children padded by the theme padding.", Kind: scene.Ordered,
			DefaultPlace: defaultPlace(320, 240),
		},
		{
			Type: "border", Title: "Border", Category: "Layout", Icon: "layout",
			Doc: "Five slots: top, bottom, left, right, center. Center takes the remaining space.", Kind: scene.Slots,
			Slots: []string{"top", "bottom", "left", "right", "center"}, DefaultPlace: defaultPlace(400, 300),
		},
		{
			Type: "hsplit", Title: "Horizontal split", Category: "Layout", Icon: "split",
			Doc: "Two children, leading and trailing, with a draggable divider.", Kind: scene.Slots,
			Slots: []string{"leading", "trailing"}, MaxChildren: 2, DefaultPlace: defaultPlace(480, 320),
			Props: []scene.PropSpec{num("offset", "Divider offset", "Layout", 0.5, 0, 1)},
		},
		{
			Type: "vsplit", Title: "Vertical split", Category: "Layout", Icon: "split",
			Doc: "Two children, top and bottom, with a draggable divider.", Kind: scene.Slots,
			Slots: []string{"leading", "trailing"}, MaxChildren: 2, DefaultPlace: defaultPlace(480, 320),
			Props: []scene.PropSpec{num("offset", "Divider offset", "Layout", 0.5, 0, 1)},
		},
		{
			Type: "scroll", Title: "Scroll", Category: "Layout", Icon: "scroll",
			Doc: "One child inside a scroll area.", Kind: scene.Slots,
			Slots: []string{"content"}, MaxChildren: 1, DefaultPlace: defaultPlace(320, 240),
		},
		{
			Type: "card", Title: "Card", Category: "Layout", Icon: "card",
			Doc: "A titled card around one content child.", Kind: scene.Slots,
			Slots: []string{"content"}, MaxChildren: 1, DefaultPlace: defaultPlace(320, 200),
			Props: []scene.PropSpec{
				p("title", "Title", "Content", scene.PropString, ""),
				p("subtitle", "Subtitle", "Content", scene.PropString, ""),
			},
		},
		{
			Type: "tabs", Title: "Tabs", Category: "Layout", Icon: "tabs",
			Doc: "One tab per child; the child's place.title names the tab.", Kind: scene.Tabs,
			DefaultPlace: defaultPlace(480, 320),
			Props:        []scene.PropSpec{enum("location", "Tab location", "Layout", "top", "top", "bottom", "left", "right")},
		},
		{
			Type: "accordion", Title: "Accordion", Category: "Layout", Icon: "accordion",
			Doc: "One collapsible section per child; the child's place.title names the section.", Kind: scene.Tabs,
			DefaultPlace: defaultPlace(320, 320),
		},

		// --------------------------------------------------------------- display
		{
			Type: "label", Title: "Label", Category: "Display", Icon: "document",
			Doc: "Text with the Fyne label styling.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(160, 32),
			DefaultProps: map[string]any{"text": "Label"},
			Props: []scene.PropSpec{
				p("text", "Text", "Content", scene.PropText, ""),
				p("bold", "Bold", "Look", scene.PropBool, false),
				p("italic", "Italic", "Look", scene.PropBool, false),
				p("mono", "Monospace", "Look", scene.PropBool, false),
				enum("align", "Alignment", "Layout", "left", "left", "center", "right"),
				enum("wrap", "Wrapping", "Layout", "word", "off", "word", "break"),
				enum("truncation", "Truncation", "Layout", "off", "off", "ellipsis"),
			},
		},
		{
			Type: "text", Title: "Canvas text", Category: "Display", Icon: "text",
			Doc: "Scaled canvas text: its font size follows the scene scale, unlike widget text.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(200, 40),
			DefaultProps: map[string]any{"text": "Text", "size": 18.0},
			Props: []scene.PropSpec{
				p("text", "Text", "Content", scene.PropText, ""),
				num("size", "Size", "Look", 18, 1, 500),
				p("color", "Color", "Look", scene.PropColor, "#ffffff"),
				p("bold", "Bold", "Look", scene.PropBool, false),
				p("italic", "Italic", "Look", scene.PropBool, false),
				p("mono", "Monospace", "Look", scene.PropBool, false),
				enum("align", "Alignment", "Layout", "left", "left", "center", "right"),
			},
		},
		{
			Type: "markdown", Title: "Markdown", Category: "Display", Icon: "document",
			Doc: "Rich text rendered from Markdown.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(320, 160),
			DefaultProps: map[string]any{"text": "**Hello** from a scene."},
			Props:        []scene.PropSpec{p("text", "Markdown", "Content", scene.PropText, "")},
		},
		{
			Type: "image", Title: "Image", Category: "Display", Icon: "image",
			Doc: "A scene asset (tie:hash, http(s), file: or a path relative to the document).", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(320, 240),
			Props: []scene.PropSpec{
				p("src", "Source", "Content", scene.PropAsset, ""),
				enum("fill", "Fill", "Look", "cover", "stretch", "contain", "cover", "original"),
				p("alt", "Description", "Content", scene.PropString, ""),
			},
			Events: []scene.EventSpec{evTapped, evDouble, evHoverIn, evHoverOut},
		},
		{
			Type: "icon", Title: "Icon", Category: "Display", Icon: "star",
			Doc: "A theme icon by name (info, home, folder, settings, search, image, audio, video, delete, add, back, menu, warning, document).", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(48, 48),
			DefaultProps: map[string]any{"name": "info"},
			Props:        []scene.PropSpec{p("name", "Icon", "Content", scene.PropIcon, "info")},
			Events:       []scene.EventSpec{evTapped},
		},
		{
			Type: "separator", Title: "Separator", Category: "Display", Icon: "minus",
			Doc: "A horizontal separator line.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(200, 8),
		},
		{
			Type: "progress", Title: "Progress", Category: "Display", Icon: "progress",
			Doc: "A determinate progress bar (0..1).", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(200, 16),
			DefaultProps: map[string]any{"value": 0.5},
			Props:        []scene.PropSpec{num("value", "Value", "Content", 0.5, 0, 1)},
		},
		{
			Type: "activity", Title: "Activity", Category: "Display", Icon: "activity",
			Doc: "An indeterminate activity indicator.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(48, 48),
		},

		// ---------------------------------------------------------------- shapes
		{
			Type: "rect", Title: "Rectangle", Category: "Shapes", Icon: "square",
			Doc: "A filled rectangle with optional stroke and rounded corners.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(200, 120),
			DefaultProps: map[string]any{"fill": "#3b82f6"},
			Props: []scene.PropSpec{
				p("fill", "Fill", "Look", scene.PropColor, "#3b82f6"),
				p("stroke", "Stroke", "Look", scene.PropColor, "#00000000"),
				num("strokeWidth", "Stroke width", "Look", 0, 0, 40),
				num("cornerRadius", "Corner radius", "Look", 0, 0, 200),
			},
			Events: []scene.EventSpec{evTapped},
		},
		{
			Type: "circle", Title: "Circle", Category: "Shapes", Icon: "circle",
			Doc: "A filled circle with optional stroke.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(96, 96),
			DefaultProps: map[string]any{"fill": "#22c55e"},
			Props: []scene.PropSpec{
				p("fill", "Fill", "Look", scene.PropColor, "#22c55e"),
				p("stroke", "Stroke", "Look", scene.PropColor, "#00000000"),
				num("strokeWidth", "Stroke width", "Look", 0, 0, 40),
			},
			Events: []scene.EventSpec{evTapped},
		},
		{
			Type: "line", Title: "Line", Category: "Shapes", Icon: "line",
			Doc: "A straight line across the node's box.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(160, 2),
			DefaultProps: map[string]any{"color": "#ffffff", "width": 2.0},
			Props: []scene.PropSpec{
				p("color", "Color", "Look", scene.PropColor, "#ffffff"),
				num("width", "Width", "Look", 2, 0.5, 40),
			},
		},

		// ----------------------------------------------------------------- input
		{
			Type: "button", Title: "Button", Category: "Input", Icon: "button",
			Doc: "A button. Add actions to its tapped event or bind a call handler.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(140, 40),
			DefaultProps: map[string]any{"text": "Button"},
			Props: []scene.PropSpec{
				p("text", "Text", "Content", scene.PropString, "Button"),
				p("icon", "Icon", "Content", scene.PropIcon, ""),
				enum("importance", "Importance", "Look", "medium", "low", "medium", "high", "danger", "warning", "success"),
				p("disabled", "Disabled", "Content", scene.PropBool, false),
			},
			Events: []scene.EventSpec{evTapped},
		},
		{
			Type: "hyperlink", Title: "Hyperlink", Category: "Input", Icon: "link",
			Doc: "A link that opens its URL in the system browser.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(200, 32),
			DefaultProps: map[string]any{"text": "Link", "url": "https://fyne.io"},
			Props: []scene.PropSpec{
				p("text", "Text", "Content", scene.PropString, "Link"),
				p("url", "URL", "Content", scene.PropURL, "https://fyne.io"),
			},
			Events: []scene.EventSpec{evTapped},
		},
		{
			Type: "entry", Title: "Entry", Category: "Input", Icon: "edit",
			Doc: "Text input; page.Values() collects every entry by node id.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(220, 40),
			Props: []scene.PropSpec{
				p("text", "Text", "Content", scene.PropString, ""),
				p("placeholder", "Placeholder", "Content", scene.PropString, ""),
				p("multiline", "Multiline", "Content", scene.PropBool, false),
				p("password", "Password", "Content", scene.PropBool, false),
			},
			Events: []scene.EventSpec{evChanged, evSubmitted},
		},
		{
			Type: "check", Title: "Check", Category: "Input", Icon: "check",
			Doc: "A checkbox.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(180, 32),
			DefaultProps: map[string]any{"text": "Check"},
			Props: []scene.PropSpec{
				p("text", "Text", "Content", scene.PropString, "Check"),
				p("checked", "Checked", "Content", scene.PropBool, false),
			},
			Events: []scene.EventSpec{evChanged},
		},
		{
			Type: "select", Title: "Select", Category: "Input", Icon: "select",
			Doc: "A dropdown of options.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(200, 40),
			DefaultProps: map[string]any{"options": []any{"One", "Two", "Three"}},
			Props: []scene.PropSpec{
				p("options", "Options", "Content", scene.PropOptions, nil),
				p("selected", "Selected", "Content", scene.PropString, ""),
			},
			Events: []scene.EventSpec{evChanged},
		},
		{
			Type: "radio", Title: "Radio group", Category: "Input", Icon: "radio",
			Doc: "A group of radio buttons.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(200, 120),
			DefaultProps: map[string]any{"options": []any{"One", "Two", "Three"}},
			Props: []scene.PropSpec{
				p("options", "Options", "Content", scene.PropOptions, nil),
				p("selected", "Selected", "Content", scene.PropString, ""),
			},
			Events: []scene.EventSpec{evChanged},
		},
		{
			Type: "slider", Title: "Slider", Category: "Input", Icon: "slider",
			Doc: "A numeric slider.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(220, 40),
			DefaultProps: map[string]any{"value": 50.0, "min": 0.0, "max": 100.0, "step": 1.0},
			Props: []scene.PropSpec{
				p("value", "Value", "Content", scene.PropNumber, 0.0),
				p("min", "Minimum", "Content", scene.PropNumber, 0.0),
				p("max", "Maximum", "Content", scene.PropNumber, 100.0),
				p("step", "Step", "Content", scene.PropNumber, 1.0),
			},
			Events: []scene.EventSpec{evChanged},
		},

		// ------------------------------------------------------------ navigation
		{
			Type: "hotspot", Title: "Hotspot", Category: "Navigation", Icon: "target",
			Doc: "A transparent clickable region over a background; the core point-and-click primitive.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(120, 80),
			Props: []scene.PropSpec{
				enum("shape", "Shape", "Look", "rect", "rect", "ellipse", "polygon"),
				p("fill", "Fill", "Look", scene.PropColor, "#ffffff00"),
				p("hoverFill", "Hover fill", "Look", scene.PropColor, "#ffffff33"),
				p("stroke", "Edit stroke", "Look", scene.PropColor, "#ffffffaa"),
				num("strokeWidth", "Stroke width", "Look", 1, 0, 20),
				scene.PropSpec{Name: "points", Title: "Polygon points", Group: "Look", Kind: scene.PropPoints,
					Default: nil, Help: "Normalized {x,y} points (0..1) for shape=polygon, clockwise."},
				p("label", "Status text", "Content", scene.PropString, ""),
				p("showInPlay", "Visible in play mode", "Look", scene.PropBool, false),
			},
			Events: []scene.EventSpec{evTapped, evDouble, evSecondary, evHoverIn, evHoverOut},
		},
		{
			Type: "link", Title: "Scene link", Category: "Navigation", Icon: "link",
			Doc: "A button that navigates to another scene; it emits tapped and falls back to navigating to `to` when no action is set.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(160, 40),
			DefaultProps: map[string]any{"text": "Open", "to": ""},
			Props: []scene.PropSpec{
				p("text", "Text", "Content", scene.PropString, "Open"),
				p("to", "Target scene", "Content", scene.PropSceneRef, ""),
			},
			Events: []scene.EventSpec{evTapped},
		},
		{
			Type: "embed", Title: "Embedded scene", Category: "Navigation", Icon: "embed",
			Doc: "Embeds another scene document as a reusable component. Rendered as a placeholder until the host resolves it.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(320, 240),
			Props:        []scene.PropSpec{p("src", "Scene", "Content", scene.PropSceneRef, "")},
		},
		{
			Type: "slot", Title: "Script slot", Category: "Navigation", Icon: "slot",
			Doc: "An area the page script fills at runtime (node.SetContent / node.Append). In edit mode it shows its hint.", Kind: scene.Leaf,
			DefaultPlace: defaultPlace(320, 200),
			Props:        []scene.PropSpec{p("hint", "Hint", "Content", scene.PropString, "Script fills this area")},
		},
	}
}
