package render

import (
	"context"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
)

// Options configures a View.
type Options struct {
	Registry  *Registry
	Assets    AssetResolver
	Embed     EmbedResolver
	Mode      Mode
	Base      string
	Context   context.Context
	Navigator Navigator
	Opener    Opener
	Caller    Caller
	Status    func(string)
}

// View renders a document and applies document changes incrementally. It is a
// Fyne widget; put it in a Stack under the designer overlay, or use it as page
// content directly.
//
// All methods must be called on the UI thread (Fyne's main goroutine), like
// every other widget. Async component work marshals through guithread.Do.
type View struct {
	widget.BaseWidget

	doc  *scene.Document
	reg  *Registry
	opts Options

	root      Live
	buildErr  error
	lives     map[string]Live
	parents   map[string]string
	rects     map[string]Rect
	runtimePl map[string]scene.Place

	mu        sync.Mutex
	listeners map[string]map[string][]func(map[string]any)

	engine *Engine
	ctx    context.Context
	dead   bool
}

// NewView builds a view for doc. Registry must contain factories for every
// component type used by the document; unknown types render as placeholders.
func NewView(doc *scene.Document, opts Options) *View {
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	if opts.Assets == nil {
		opts.Assets = NewLocalAssets()
	}
	if opts.Embed == nil {
		opts.Embed = &StoreResolver{}
	}
	v := &View{
		doc:       doc,
		reg:       opts.Registry,
		opts:      opts,
		lives:     make(map[string]Live),
		parents:   make(map[string]string),
		runtimePl: make(map[string]scene.Place),
		listeners: make(map[string]map[string][]func(map[string]any)),
		ctx:       opts.Context,
	}
	v.engine = &Engine{view: v}
	v.ExtendBaseWidget(v)
	v.rebuild()
	return v
}

// CanvasObject lets the View satisfy fynerisor's IsCanvasObject interface, so
// window.SetContent(view) works.
func (v *View) CanvasObject() fyne.CanvasObject { return v }

// Doc returns the displayed document.
func (v *View) Doc() *scene.Document { return v.doc }

// LoadDocument replaces the document in the same view widget (used after a
// conflict reload or draft restore). Script listeners are cleared because the
// old node IDs no longer refer to the visible tree.
func (v *View) LoadDocument(doc *scene.Document) {
	v.doc = doc
	v.runtimePl = make(map[string]scene.Place)
	v.mu.Lock()
	v.listeners = make(map[string]map[string][]func(map[string]any))
	v.mu.Unlock()
	v.rebuild()
	v.rects = nil
	v.Refresh()
}

// Registry returns the render registry.
func (v *View) Registry() *Registry { return v.reg }

// Embedded builds a nested view for an embed node, sharing the registry,
// assets, resolver and action targets but using the embedded document's base.
func (v *View) Embedded(doc *scene.Document, base string) *View {
	return NewView(doc, Options{
		Registry:  v.reg,
		Assets:    v.opts.Assets,
		Embed:     v.opts.Embed,
		Mode:      v.opts.Mode,
		Base:      base,
		Context:   v.ctx,
		Navigator: v.opts.Navigator,
		Opener:    v.opts.Opener,
		Caller:    v.opts.Caller,
		Status:    v.opts.Status,
	})
}

// Mode returns the current rendering mode.
func (v *View) Mode() Mode { return v.opts.Mode }

// SetMode switches between play and edit rendering. The tree is rebuilt (mode
// changes what components show), but registered script listeners survive.
func (v *View) SetMode(mode Mode) {
	if v.opts.Mode == mode {
		return
	}
	v.opts.Mode = mode
	v.rebuild()
}

// SetCaller installs the script call handler for `call` actions.
func (v *View) SetCaller(c Caller) { v.opts.Caller = c }

// SetNavigator installs the navigation target.
func (v *View) SetNavigator(n Navigator) { v.opts.Navigator = n }

// SetOpener installs the media/external opener.
func (v *View) SetOpener(o Opener) { v.opts.Opener = o }

// SetStatus installs the status sink.
func (v *View) SetStatus(fn func(string)) { v.opts.Status = fn }

// CreateRenderer implements fyne.Widget.
func (v *View) CreateRenderer() fyne.WidgetRenderer { return &viewRenderer{v: v} }

// RefreshRects invalidates the cached geometry.
func (v *View) RefreshRects() { v.rects = nil }

// RectOf returns a node's rectangle relative to the View. The result is cached;
// call RefreshRects after layout changes.
func (v *View) RectOf(id string) (Rect, bool) {
	if v.rects == nil {
		v.computeRects()
	}
	r, ok := v.rects[id]
	return r, ok
}

// EachRect calls fn for every node rect in document order.
func (v *View) EachRect(fn func(id string, r Rect)) {
	if v.rects == nil {
		v.computeRects()
	}
	v.doc.Walk(func(n *scene.Node) bool {
		if r, ok := v.rects[n.ID]; ok {
			fn(n.ID, r)
		}
		return true
	})
}

// Order returns the node IDs in paint order (document order).
func (v *View) Order() []string {
	var out []string
	v.doc.Walk(func(n *scene.Node) bool { out = append(out, n.ID); return true })
	return out
}

// Node returns a node by ID.
func (v *View) Node(id string) (*scene.Node, bool) { return v.doc.Find(id) }

// LiveObject returns a node's underlying Fyne object (without the size-hint
// wrapper). It is used by the designer and the script binding.
func (v *View) LiveObject(id string) (fyne.CanvasObject, bool) {
	l, ok := v.lives[id]
	if !ok {
		return nil, false
	}
	return Unwrap(l.Object()), true
}

// Live returns the live component for a node.
func (v *View) Live(id string) (Live, bool) {
	l, ok := v.lives[id]
	return l, ok
}

// RuntimePlace changes a node's placement without touching the document.
func (v *View) RuntimePlace(id string, p scene.Place) bool {
	parent := v.parents[id]
	cont, _ := v.lives[parent].(ContainerLive)
	child, ok := v.lives[id]
	if !ok || cont == nil {
		return false
	}
	if err := cont.SetPlace(child, p); err != nil {
		return false
	}
	v.runtimePl[id] = p
	v.rects = nil
	return true
}

// Unwrap removes size-hint and theme-override wrappers from an object.
func Unwrap(o fyne.CanvasObject) fyne.CanvasObject {
	for i := 0; i < 8; i++ {
		switch x := o.(type) {
		case *hintObject:
			o = x.content
		case *container.ThemeOverride:
			o = x.Content
		default:
			return o
		}
	}
	return o
}

// ParentOf returns the parent of a node and its index.
func (v *View) ParentOf(id string) (*scene.Node, int, bool) { return v.doc.ParentOf(id) }

// ParentID returns the parent's node ID ("" for the root).
func (v *View) ParentID(id string) string { return v.parents[id] }

// ContainerKind returns the component kind of a node.
func (v *View) ContainerKind(id string) scene.ContainerKind {
	n, ok := v.doc.Find(id)
	if !ok || v.reg == nil {
		return scene.Leaf
	}
	spec, ok := v.reg.Spec(n.Type)
	if !ok {
		return scene.Leaf
	}
	return spec.Kind
}

// NodeAt returns the topmost selectable node at a point in view coordinates.
// Locked and hidden nodes are skipped.
func (v *View) NodeAt(pos fyne.Position) string {
	order := v.Order()
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		n, ok := v.doc.Find(id)
		if !ok || n.Locked || n.Hidden || !v.IsVisible(id) {
			continue
		}
		if r, ok := v.RectOf(id); ok && r.Contains(pos) {
			return id
		}
	}
	return ""
}

// ContainerAt returns the deepest container node containing pos.
func (v *View) ContainerAt(pos fyne.Position) string {
	order := v.Order()
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		n, ok := v.doc.Find(id)
		if !ok || n.Locked || n.Hidden || !v.IsVisible(id) {
			continue
		}
		if v.ContainerKind(id) == scene.Leaf {
			continue
		}
		if r, ok := v.RectOf(id); ok && r.Contains(pos) {
			return id
		}
	}
	return ""
}

// DropSpotAt maps a view point to a container insertion point.
func (v *View) DropSpotAt(pos fyne.Position) (string, DropSpot, bool) {
	id := v.ContainerAt(pos)
	if id == "" {
		return "", DropSpot{}, false
	}
	cont, ok := v.lives[id].(ContainerLive)
	if !ok {
		return "", DropSpot{}, false
	}
	r, _ := v.RectOf(id)
	spot, ok := cont.DropSpot(fyne.NewPos(pos.X-r.X, pos.Y-r.Y))
	if !ok {
		return "", DropSpot{}, false
	}
	return id, spot, true
}

// PreviewPlace shows a temporary placement during a drag.
func (v *View) PreviewPlace(id string, p scene.Place) {
	parent := v.parents[id]
	cont, _ := v.lives[parent].(ContainerLive)
	child, ok := v.lives[id]
	if !ok || cont == nil {
		return
	}
	cont.Preview(child, p)
}

// ClearPreview removes a drag preview.
func (v *View) ClearPreview(id string) {
	parent := v.parents[id]
	cont, _ := v.lives[parent].(ContainerLive)
	child, ok := v.lives[id]
	if !ok || cont == nil {
		return
	}
	if c, ok := cont.(interface{ CancelPreview(Live) }); ok {
		c.CancelPreview(child)
	}
}

// BuildError reports an error that prevented the root from rendering.
func (v *View) BuildError() error { return v.buildErr }

// On registers a script listener for one node event. The returned function
// removes the listener.
func (v *View) On(id, event string, fn func(map[string]any)) func() {
	v.mu.Lock()
	defer v.mu.Unlock()
	byEvent := v.listeners[id]
	if byEvent == nil {
		byEvent = make(map[string][]func(map[string]any))
		v.listeners[id] = byEvent
	}
	byEvent[event] = append(byEvent[event], fn)
	idx := len(byEvent[event]) - 1
	return func() {
		v.mu.Lock()
		defer v.mu.Unlock()
		cur := v.listeners[id][event]
		if idx < len(cur) {
			cur[idx] = nil
		}
	}
}

// SetVisible shows or hides a node at runtime.
func (v *View) SetVisible(id string, visible bool) {
	if l, ok := v.lives[id]; ok {
		if visible {
			l.Object().Show()
		} else {
			l.Object().Hide()
		}
		l.Object().Refresh()
	}
}

// IsVisible reports a node's runtime visibility.
func (v *View) IsVisible(id string) bool {
	if l, ok := v.lives[id]; ok {
		return l.Object().Visible()
	}
	return false
}

// RuntimeSet changes a property at runtime without touching the document. It
// is what declarative `set` actions and node.Set use.
func (v *View) RuntimeSet(id, prop string, value any) bool {
	n, ok := v.doc.Find(id)
	if !ok {
		return false
	}
	l, ok := v.lives[id]
	if !ok {
		return false
	}
	clone := n.Clone()
	clone.SetProp(prop, value)
	if err := l.Update(clone, scene.Fields{prop: true}); err != nil {
		if err == ErrRebuild {
			_ = v.rebuildNode(id)
			return true
		}
		return false
	}
	l.Object().Refresh()
	return true
}

// Apply applies document changes to the live tree. It is the incremental path
// used by the designer; the document has already been mutated.
func (v *View) Apply(changes []scene.Change) error {
	for _, ch := range changes {
		switch ch.Kind {
		case scene.OpSetProp, scene.OpSetField:
			if ch.Fields["style"] || ch.Fields["type"] {
				if err := v.rebuildNode(ch.Node); err != nil {
					return err
				}
				continue
			}
			l, ok := v.lives[ch.Node]
			if !ok {
				continue
			}
			n, ok := v.doc.Find(ch.Node)
			if !ok {
				continue
			}
			if err := l.Update(n, ch.Fields); err != nil && err != ErrRebuild {
				return err
			} else if err == ErrRebuild {
				if err := v.rebuildNode(ch.Node); err != nil {
					return err
				}
				continue
			}
			if ch.Fields["hidden"] {
				v.SetVisible(ch.Node, !n.Hidden)
			}
			l.Object().Refresh()

		case scene.OpSetPlace:
			delete(v.runtimePl, ch.Node)
			parent := v.parents[ch.Node]
			cont, _ := v.lives[parent].(ContainerLive)
			child, ok := v.lives[ch.Node]
			if !ok || cont == nil {
				continue
			}
			n, _ := v.doc.Find(ch.Node)
			if err := cont.SetPlace(child, n.Place); err != nil {
				return err
			}

		case scene.OpInsert:
			parent := ch.Parent
			cont, _ := v.lives[parent].(ContainerLive)
			if cont == nil {
				continue
			}
			live, err := v.build(ch.Subtree, parent)
			if err != nil {
				return err
			}
			if err := cont.Insert(ch.Index, live, ch.Subtree.Place); err != nil {
				return err
			}
			if ch.Subtree.Hidden {
				live.Object().Hide()
			}

		case scene.OpRemove:
			parent := v.parents[ch.Node]
			cont, _ := v.lives[parent].(ContainerLive)
			child, ok := v.lives[ch.Node]
			if !ok || cont == nil {
				continue
			}
			if err := cont.Remove(child); err != nil {
				return err
			}
			v.unregister(ch.Node)

		case scene.OpMove:
			from, _ := v.lives[ch.FromParent].(ContainerLive)
			to, _ := v.lives[ch.Parent].(ContainerLive)
			child, ok := v.lives[ch.Node]
			n, _ := v.doc.Find(ch.Node)
			if !ok || from == nil || to == nil || n == nil {
				continue
			}
			if err := from.Remove(child); err != nil {
				return err
			}
			if err := to.Insert(ch.Index, child, n.Place); err != nil {
				return err
			}
			v.parents[ch.Node] = ch.Parent
		}
	}
	v.rects = nil
	// Repaint the view itself. View.Refresh() would cascade into every child,
	// and a canvas.Image refresh decodes its resource again — on every drag
	// commit, nudge and inspector keystroke.
	canvas.Refresh(v)
	return nil
}

// Destroy releases the tree. Call it when the page is discarded.
func (v *View) Destroy() {
	v.dead = true
	if v.root != nil {
		v.root.Destroy()
	}
}

func (v *View) rebuild() {
	if v.doc == nil || v.doc.Root == nil {
		v.buildErr = fmt.Errorf("scene/render: document has no root")
		v.root = newErrorLive("scene has no root")
		v.Refresh()
		return
	}
	if v.root != nil {
		v.root.Destroy()
	}
	v.lives = make(map[string]Live)
	v.parents = make(map[string]string)
	v.rects = nil
	v.buildErr = nil
	if v.reg == nil {
		v.buildErr = fmt.Errorf("scene/render: no component registry")
		v.root = newErrorLive(v.buildErr.Error())
		v.Refresh()
		return
	}
	root, err := v.build(v.doc.Root, "")
	if err != nil {
		v.buildErr = err
		v.root = newErrorLive(err.Error())
	} else {
		v.root = root
	}
	v.Refresh()
}

func (v *View) build(n *scene.Node, parent string) (Live, error) {
	bc := &BuildContext{
		Context: v.ctx,
		Mode:    v.opts.Mode,
		Base:    v.opts.Base,
		Assets:  v.opts.Assets,
		Embed:   v.opts.Embed,
		Status:  v.opts.Status,
		View:    v,
		emit:    v.EmitEvent,
	}
	live, err := v.reg.Build(bc, n)
	if err != nil {
		return nil, err
	}
	if cont, ok := live.(ContainerLive); ok {
		for i, ch := range n.Children {
			child, err := v.build(ch, n.ID)
			if err != nil {
				return nil, err
			}
			if err := cont.Insert(i, child, ch.Place); err != nil {
				return nil, err
			}
		}
	}
	live = wrapStyle(live, n.Style)
	live = wrapHint(live)
	v.lives[n.ID] = live
	v.parents[n.ID] = parent
	if n.Hidden {
		live.Object().Hide()
	}
	return live, nil
}

func (v *View) rebuildNode(id string) error {
	n, ok := v.doc.Find(id)
	if !ok {
		return fmt.Errorf("scene/render: node %q not found", id)
	}
	parentID := v.parents[id]
	cont, _ := v.lives[parentID].(ContainerLive)
	old, ok := v.lives[id]
	if !ok || cont == nil {
		return fmt.Errorf("scene/render: cannot rebuild %q", id)
	}
	_, index, ok := v.doc.ParentOf(id)
	if !ok {
		return fmt.Errorf("scene/render: node %q has no parent", id)
	}
	if err := cont.Remove(old); err != nil {
		return err
	}
	v.unregister(id)
	live, err := v.build(n, parentID)
	if err != nil {
		return err
	}
	if err := cont.Insert(index, live, n.Place); err != nil {
		return err
	}
	v.rects = nil
	canvas.Refresh(v)
	return nil
}

func (v *View) unregister(id string) {
	n, ok := v.doc.Find(id)
	if !ok {
		return
	}
	n.Walk(func(x *scene.Node) bool {
		delete(v.lives, x.ID)
		delete(v.parents, x.ID)
		return true
	})
}

// EmitEvent is called by components. It completes the event payload (every key
// is always present, because Risor treats a missing map key as an error) and
// runs declarative actions, then script listeners.
func (v *View) EmitEvent(id, event string, payload map[string]any) {
	n, ok := v.doc.Find(id)
	if !ok || v.dead {
		return
	}
	evt := map[string]any{
		"id": id, "name": n.Name, "type": n.Type, "event": event,
		"x": 0.0, "y": 0.0, "item": "", "value": "",
	}
	for k, val := range payload {
		evt[k] = val
	}
	v.engine.Run(n, event, evt)
}

func (v *View) computeRects() {
	v.rects = make(map[string]Rect)
	if v.doc == nil || v.doc.Root == nil || v.root == nil {
		return
	}
	size := v.Size()
	if size.Width == 0 || size.Height == 0 {
		size = v.root.Object().MinSize()
	}
	viewAbs := fyne.NewPos(0, 0)
	if app := fyne.CurrentApp(); app != nil && app.Driver() != nil {
		viewAbs = app.Driver().AbsolutePositionForObject(v)
	}
	v.walkRects(v.doc.Root, NewRect(0, 0, size.Width, size.Height), viewAbs)
}

func (v *View) walkRects(n *scene.Node, rect Rect, viewAbs fyne.Position) {
	v.rects[n.ID] = rect
	if len(n.Children) == 0 {
		return
	}
	kind := scene.Leaf
	if v.reg != nil {
		if spec, ok := v.reg.Spec(n.Type); ok {
			kind = spec.Kind
		}
	}
	free, _ := v.lives[n.ID].(interface{ FreeLayout() *FreeLayout })
	for _, ch := range n.Children {
		var childRect Rect
		if kind == scene.Free && free != nil {
			place := ch.Place
			if rp, ok := v.runtimePl[ch.ID]; ok {
				place = rp
			}
			p := free.FreeLayout().ChildRect(rect.Size(), place)
			childRect = NewRect(rect.X+p.X, rect.Y+p.Y, p.W, p.H)
		} else {
			childRect = v.absoluteRect(ch.ID, rect, viewAbs)
		}
		v.walkRects(ch, childRect, viewAbs)
	}
}

func (v *View) absoluteRect(id string, fallback Rect, viewAbs fyne.Position) Rect {
	l, ok := v.lives[id]
	if !ok {
		return fallback
	}
	app := fyne.CurrentApp()
	if app == nil || app.Driver() == nil {
		return fallback
	}
	pos := app.Driver().AbsolutePositionForObject(l.Object())
	size := l.Object().Size()
	return NewRect(pos.X-viewAbs.X, pos.Y-viewAbs.Y, size.Width, size.Height)
}

type errorLive struct {
	label *widget.Label
	box   fyne.CanvasObject
}

func newErrorLive(msg string) Live {
	l := widget.NewLabel("Scene error:\n" + msg)
	l.Wrapping = fyne.TextWrapWord
	// Build the box once: Object() is called by the renderer on every layout
	// and paint, and a fresh object each time breaks identity and caching.
	return &errorLive{label: l, box: container.NewPadded(l)}
}

func (e *errorLive) Object() fyne.CanvasObject { return e.box }
func (e *errorLive) Update(n *scene.Node, _ scene.Fields) error {
	e.label.SetText("Scene error:\n" + n.Type)
	return nil
}
func (e *errorLive) Destroy() {}

type viewRenderer struct{ v *View }

func (r *viewRenderer) Layout(size fyne.Size) {
	if r.v.root != nil {
		r.v.root.Object().Move(fyne.NewPos(0, 0))
		r.v.root.Object().Resize(size)
	}
	r.v.rects = nil
}
func (r *viewRenderer) MinSize() fyne.Size {
	if r.v.root == nil {
		return fyne.NewSize(1, 1)
	}
	return r.v.root.Object().MinSize()
}
func (r *viewRenderer) Refresh() {
	if r.v.root != nil {
		r.v.root.Object().Refresh()
	}
	r.v.rects = nil
}
func (r *viewRenderer) Objects() []fyne.CanvasObject {
	if r.v.root == nil {
		return nil
	}
	return []fyne.CanvasObject{r.v.root.Object()}
}
func (r *viewRenderer) Destroy() {}
