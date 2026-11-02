package designer

import (
	"encoding/json"
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/internal/colors"
	"github.com/uidbz/fynerisor/scene"
)

type inspector struct {
	ed      *Editor
	box     *fyne.Container
	scroll  *container.Scroll
	first   fyne.Focusable
	loading bool
}

func newInspector(e *Editor) *inspector {
	i := &inspector{ed: e, box: container.NewVBox()}
	i.scroll = container.NewVScroll(i.box)
	return i
}

func (i *inspector) object() fyne.CanvasObject { return i.scroll }

func (i *inspector) focusFirst() {
	if i.first != nil {
		if c := i.ed.canvas(); c != nil {
			c.Focus(i.first)
		}
	}
}

func (i *inspector) section(title string) {
	l := widget.NewLabel(title)
	l.TextStyle.Bold = true
	i.box.Add(l)
}

func (i *inspector) row(label string, editor fyne.CanvasObject) {
	form := container.New(layout.NewFormLayout(), widget.NewLabel(label), editor)
	i.box.Add(form)
}

func (i *inspector) refresh() {
	i.loading = true
	defer func() { i.loading = false }()
	i.first = nil
	i.box.Objects = nil
	ids := i.ed.selectedIDs()
	if len(ids) == 0 {
		i.box.Add(widget.NewLabel("Nothing selected.\n\nClick a node on the canvas, or pick a component from the Palette."))
		i.box.Refresh()
		return
	}
	primary, ok := i.ed.Primary()
	if !ok {
		i.box.Add(widget.NewLabel("Nothing selected."))
		i.box.Refresh()
		return
	}
	if len(ids) > 1 {
		l := widget.NewLabel(fmt.Sprintf("%d nodes selected — editing properties of %s", len(ids), primary.Type))
		l.Wrapping = fyne.TextWrapWord
		i.box.Add(l)
		i.box.Add(widget.NewSeparator())
	}

	i.section(fmt.Sprintf("%s  ·  %s", primary.Type, primary.ID))
	name := widget.NewEntry()
	name.SetText(primary.Name)
	name.SetPlaceHolder("Display name")
	name.OnChanged = func(s string) {
		if i.loading {
			return
		}
		var ops []scene.Op
		for _, id := range i.targetIDs() {
			if n, ok := i.ed.doc.Find(id); ok && n.Name != s {
				ops = append(ops, scene.SetFieldOp(id, "name", n.Name, s))
			}
		}
		if len(ops) > 0 {
			i.ed.applyEdit("rename", "name:"+primary.ID, ops...)
			i.ed.outline.refresh()
		}
	}
	i.row("Name", name)
	if i.first == nil {
		i.first = name
	}
	i.row("ID", widget.NewLabel(primary.ID))

	i.placement(primary)
	i.props(primary)
	i.events(primary)
	i.box.Add(widget.NewLabel(""))
	i.box.Refresh()
}

func (i *inspector) targetIDs() []string {
	ids := i.ed.selectedIDs()
	if len(ids) <= 1 {
		return ids
	}
	primary, ok := i.ed.Primary()
	if !ok {
		return ids
	}
	var out []string
	for _, id := range ids {
		if n, ok := i.ed.doc.Find(id); ok && n.Type == primary.Type {
			out = append(out, id)
		}
	}
	return out
}

func (i *inspector) placement(n *scene.Node) {
	parentID := i.ed.view.ParentID(n.ID)
	if parentID == "" {
		i.section("Page")
		i.propEditors(n, false)
		return
	}
	parent, ok := i.ed.doc.Find(parentID)
	if !ok {
		return
	}
	pspec, ok := i.ed.opts.Registry.Spec(parent.Type)
	if !ok {
		return
	}
	i.section("Placement in " + parent.Type)

	numEntry := func(label string, get func() float64, set func(float64)) {
		entry := widget.NewEntry()
		entry.SetText(formatNum(get()))
		entry.OnChanged = func(s string) {
			if i.loading {
				return
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				return
			}
			if v == get() {
				return
			}
			set(v)
		}
		i.row(label, entry)
	}

	place := func(update func(*scene.Place)) {
		cur, ok := i.ed.doc.Find(n.ID)
		if !ok {
			return
		}
		p := cur.Place
		update(&p)
		i.ed.applyEdit("layout", "place:"+n.ID, scene.SetPlaceOp(n.ID, cur.Place, p))
	}

	switch pspec.Kind {
	case scene.Free:
		numEntry("X", func() float64 { return n.Place.X }, func(v float64) { place(func(p *scene.Place) { p.X = v }) })
		numEntry("Y", func() float64 { return n.Place.Y }, func(v float64) { place(func(p *scene.Place) { p.Y = v }) })
		numEntry("Width", func() float64 { return n.Place.W }, func(v float64) { place(func(p *scene.Place) { p.W = v }) })
		numEntry("Height", func() float64 { return n.Place.H }, func(v float64) { place(func(p *scene.Place) { p.H = v }) })

		h := widget.NewSelect([]string{"(none)", "left", "right", "center", "stretch"}, nil)
		v := widget.NewSelect([]string{"(none)", "top", "bottom", "center", "stretch"}, nil)
		h.SetSelected(anchorValue(n.Place.Anchor, true))
		v.SetSelected(anchorValue(n.Place.Anchor, false))
		apply := func() {
			cur, ok := i.ed.doc.Find(n.ID)
			if !ok {
				return
			}
			p := cur.Place
			ah, av := h.Selected, v.Selected
			if ah == "(none)" && av == "(none)" {
				p.Anchor = nil
			} else {
				a := &scene.Anchor{}
				if ah != "(none)" {
					a.H = ah
				}
				if av != "(none)" {
					a.V = av
				}
				p.Anchor = a
			}
			i.ed.applyEdit("anchor", "place:"+n.ID, scene.SetPlaceOp(n.ID, cur.Place, p))
		}
		h.OnChanged = func(string) { apply() }
		v.OnChanged = func(string) { apply() }
		i.row("Anchor H", h)
		i.row("Anchor V", v)

	case scene.Ordered, scene.Tabs:
		numEntry("Min width", func() float64 { return n.Place.MinW }, func(v float64) { place(func(p *scene.Place) { p.MinW = v }) })
		numEntry("Min height", func() float64 { return n.Place.MinH }, func(v float64) { place(func(p *scene.Place) { p.MinH = v }) })
		if pspec.Kind == scene.Tabs {
			title := widget.NewEntry()
			title.SetText(n.Place.Title)
			title.OnChanged = func(s string) { place(func(p *scene.Place) { p.Title = s }) }
			i.row("Tab title", title)
			icon := widget.NewEntry()
			icon.SetText(n.Place.Icon)
			icon.OnChanged = func(s string) { place(func(p *scene.Place) { p.Icon = s }) }
			i.row("Tab icon", icon)
		}

	case scene.Slots:
		if len(pspec.Slots) > 0 {
			sel := widget.NewSelect(pspec.Slots, func(s string) {
				place(func(p *scene.Place) { p.Slot = s })
			})
			sel.SetSelected(n.Place.Slot)
			i.row("Slot", sel)
		}
		numEntry("Min width", func() float64 { return n.Place.MinW }, func(v float64) { place(func(p *scene.Place) { p.MinW = v }) })
		numEntry("Min height", func() float64 { return n.Place.MinH }, func(v float64) { place(func(p *scene.Place) { p.MinH = v }) })
	}
}

func (i *inspector) props(n *scene.Node) {
	i.propEditors(n, true)
}

func (i *inspector) propEditors(n *scene.Node, includeGroups bool) {
	spec, ok := i.ed.opts.Registry.Spec(n.Type)
	if !ok {
		return
	}
	groups := make(map[string][]scene.PropSpec)
	var order []string
	for _, p := range spec.Props {
		g := p.Group
		if g == "" {
			g = "Properties"
		}
		if _, ok := groups[g]; !ok {
			order = append(order, g)
		}
		groups[g] = append(groups[g], p)
	}
	sort.SliceStable(order, func(a, b int) bool {
		rank := func(s string) int {
			switch s {
			case "Content":
				return 0
			case "Layout":
				return 1
			case "Look":
				return 2
			case "Data":
				return 3
			default:
				return 4
			}
		}
		return rank(order[a]) < rank(order[b])
	})
	for _, g := range order {
		if includeGroups {
			i.section(g)
		}
		for _, p := range groups[g] {
			i.propRow(n, p)
		}
	}
}

func (i *inspector) propRow(n *scene.Node, p scene.PropSpec) {
	val := propValue(n, p)
	switch p.Kind {
	case scene.PropBool:
		check := widget.NewCheck(p.Title, nil)
		check.SetChecked(boolOf(val))
		check.OnChanged = func(b bool) { i.setProp(p.Name, b) }
		i.row(p.Title, check)
	case scene.PropEnum:
		if len(p.Enum) == 0 {
			return
		}
		sel := widget.NewSelect(p.Enum, func(s string) { i.setProp(p.Name, s) })
		sel.SetSelected(stringOf(val))
		i.row(p.Title, sel)
	case scene.PropNumber, scene.PropInt:
		entry := widget.NewEntry()
		entry.SetText(formatNum(numOf(val, 0)))
		entry.OnChanged = func(s string) {
			if i.loading {
				return
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				return
			}
			if p.Kind == scene.PropInt {
				i.setProp(p.Name, int(v))
			} else {
				i.setProp(p.Name, v)
			}
		}
		i.row(p.Title, entry)
	case scene.PropColor:
		entry := widget.NewEntry()
		entry.SetText(stringOf(val))
		entry.OnChanged = func(s string) { i.setProp(p.Name, s) }
		if i.ed.opts.Window != nil {
			pick := widget.NewButton("…", func() {
				c := renderColor(stringOf(val))
				d := dialog.NewColorPicker(p.Title, "Pick a color", func(c color.Color) {
					hex := colors.Hex(c)
					entry.SetText(hex)
					i.setProp(p.Name, hex)
				}, i.ed.opts.Window)
				if c != nil {
					d.Advanced = true
				}
				d.Show()
			})
			box := container.NewBorder(nil, nil, nil, pick, entry)
			i.row(p.Title, box)
		} else {
			i.row(p.Title, entry)
		}
	case scene.PropOptions, scene.PropTags:
		entry := widget.NewEntry()
		entry.SetText(strings.Join(listOf(val), ", "))
		entry.OnChanged = func(s string) {
			var out []any
			for _, part := range strings.Split(s, ",") {
				if v := strings.TrimSpace(part); v != "" {
					out = append(out, v)
				}
			}
			i.setProp(p.Name, out)
		}
		i.row(p.Title, entry)
	case scene.PropNodeRef:
		ids := []string{""}
		i.ed.doc.Walk(func(x *scene.Node) bool {
			ids = append(ids, x.ID)
			return true
		})
		sel := widget.NewSelect(ids, func(s string) { i.setProp(p.Name, s) })
		sel.SetSelected(stringOf(val))
		i.row(p.Title, sel)
	case scene.PropPoints:
		entry := widget.NewMultiLineEntry()
		if val != nil {
			if data, err := json.MarshalIndent(val, "", "  "); err == nil {
				entry.SetText(string(data))
			}
		}
		entry.SetMinRowsVisible(4)
		entry.OnChanged = func(s string) {
			if i.loading {
				return
			}
			var pts []struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			}
			if err := json.Unmarshal([]byte(s), &pts); err != nil {
				return
			}
			out := make([]any, 0, len(pts))
			for _, p := range pts {
				out = append(out, map[string]any{"x": scene.Round(p.X), "y": scene.Round(p.Y)})
			}
			i.setProp(p.Name, out)
		}
		i.row(p.Title, entry)
	case scene.PropText:
		entry := widget.NewMultiLineEntry()
		entry.SetText(stringOf(val))
		entry.SetMinRowsVisible(3)
		entry.OnChanged = func(s string) { i.setProp(p.Name, s) }
		i.row(p.Title, entry)
	default: // string, asset, url, sceneRef, icon
		entry := widget.NewEntry()
		entry.SetText(stringOf(val))
		entry.OnChanged = func(s string) { i.setProp(p.Name, s) }
		if (p.Kind == scene.PropAsset || p.Kind == scene.PropSceneRef) && i.ed.opts.AssetPick != nil {
			pick := widget.NewButton("…", func() {
				if ref, ok := i.ed.opts.AssetPick(entry.Text); ok {
					entry.SetText(ref)
					i.setProp(p.Name, ref)
				}
			})
			i.row(p.Title, container.NewBorder(nil, nil, nil, pick, entry))
		} else {
			i.row(p.Title, entry)
		}
	}
}

func (i *inspector) setProp(key string, value any) {
	if i.loading {
		return
	}
	ids := i.targetIDs()
	var ops []scene.Op
	for _, id := range ids {
		if n, ok := i.ed.doc.Find(id); ok {
			old, _ := n.Prop(key)
			ops = append(ops, scene.SetPropOp(id, key, old, value))
		}
	}
	if len(ops) > 0 {
		i.ed.applyEdit("set "+key, "prop:"+i.ed.primary+":"+key, ops...)
	}
}

func (i *inspector) events(n *scene.Node) {
	spec, ok := i.ed.opts.Registry.Spec(n.Type)
	if !ok || len(spec.Events) == 0 {
		return
	}
	i.section("Events")
	eventNames := make([]string, 0, len(spec.Events))
	for _, e := range spec.Events {
		eventNames = append(eventNames, e.Name)
	}
	for ev := range n.On {
		if !spec.Event(ev) {
			eventNames = append(eventNames, ev)
		}
	}
	sort.Strings(eventNames)
	for _, ev := range eventNames {
		actions := n.On[ev]
		label := widget.NewLabel(ev)
		label.TextStyle.Bold = true
		i.box.Add(label)
		if len(actions) == 0 {
			i.box.Add(widget.NewLabel("  (no actions)"))
		}
		for idx := range actions {
			a := actions[idx]
			verb := widget.NewSelect(scene.Verbs, nil)
			verb.SetSelected(a.Do)
			targetLabel, target := "Node", a.Node
			switch a.Do {
			case "navigate":
				targetLabel, target = "URL", a.To
			case "call":
				targetLabel, target = "Handler", a.Handler
			}
			entry := widget.NewEntry()
			entry.SetText(target)
			entry.SetPlaceHolder(targetLabel)
			apply := func() {
				cur := n.On[ev]
				if idx >= len(cur) {
					return
				}
				next := append([]scene.Action(nil), cur...)
				act := next[idx]
				act.Do = verb.Selected
				switch act.Do {
				case "navigate":
					act.To = entry.Text
					act.Handler, act.Node = "", ""
				case "call":
					act.Handler = entry.Text
					act.To, act.Node = "", ""
				default:
					act.Node = entry.Text
					act.To, act.Handler = "", ""
				}
				next[idx] = act
				i.setOn(ev, n, next)
			}
			verb.OnChanged = func(string) { apply() }
			entry.OnChanged = func(string) { apply() }
			del := widget.NewButton("×", func() {
				cur := n.On[ev]
				if idx >= len(cur) {
					return
				}
				next := append(append([]scene.Action(nil), cur[:idx]...), cur[idx+1:]...)
				i.setOn(ev, n, next)
			})
			i.box.Add(container.NewBorder(nil, nil, nil, del, container.NewVBox(
				container.NewGridWithColumns(2, verb, entry),
			)))
		}
		addEv := ev
		add := widget.NewButton("+ action", func() {
			cur := append([]scene.Action(nil), n.On[addEv]...)
			cur = append(cur, scene.Action{Do: "navigate"})
			i.setOn(addEv, n, cur)
		})
		i.box.Add(add)
	}
}

func (i *inspector) setOn(event string, n *scene.Node, actions []scene.Action) {
	if i.loading {
		return
	}
	cur, ok := i.ed.doc.Find(n.ID)
	if !ok {
		return
	}
	next := make(map[string][]scene.Action, len(cur.On))
	for ev, list := range cur.On {
		next[ev] = append([]scene.Action(nil), list...)
	}
	if len(actions) == 0 {
		delete(next, event)
	} else {
		next[event] = actions
	}
	i.ed.applyEdit("event "+event, "on:"+n.ID+":"+event, scene.SetFieldOp(n.ID, "on", cur.On, next))
	i.refresh()
}

func propValue(n *scene.Node, p scene.PropSpec) any {
	if v, ok := n.Prop(p.Name); ok {
		return v
	}
	return p.Default
}

func stringOf(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func boolOf(v any) bool {
	b, _ := v.(bool)
	return b
}

func numOf(v any, def float64) float64 {
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
		return x
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

func formatNum(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func anchorValue(a *scene.Anchor, horizontal bool) string {
	if a == nil {
		return "(none)"
	}
	if horizontal {
		if a.H == "" {
			return "(none)"
		}
		return a.H
	}
	if a.V == "" {
		return "(none)"
	}
	return a.V
}

func renderColor(s string) color.Color {
	c, err := colors.Parse(s)
	if err != nil {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}
	return c
}
