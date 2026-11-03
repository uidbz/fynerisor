package scene

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"

	"github.com/uidbz/fynerisor/gui"
	risorcanvas "github.com/uidbz/fynerisor/gui/canvas"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
	"github.com/uidbz/fynerisor/gui/vmexec"
	risorwidget "github.com/uidbz/fynerisor/gui/widget"
	scenedoc "github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

// PageType is the Risor type name of a live scene page.
const PageType object.Type = "scene.Page"

// Page is one rendered scene document. It implements render.Caller for
// declarative `call` actions and is the script-visible page object.
type Page struct {
	view    *render.View
	cfg     Config
	base    string
	version string

	mu       sync.Mutex
	handlers map[string]*vmexec.Callback
	pending  []pendingCall
	bound    bool
}

type pendingCall struct {
	handler string
	evt     map[string]any
}

func (p *Page) Type() object.Type           { return PageType }
func (p *Page) Inspect() string             { return "scene.Page" }
func (p *Page) Interface() interface{}      { return p.view }
func (p *Page) IsTruthy() bool              { return true }
func (p *Page) Equals(o object.Object) bool { return p == o }
func (p *Page) Attrs() []object.AttrSpec    { return pageAttrs.Specs() }
func (p *Page) SetAttr(string, object.Object) error {
	return fmt.Errorf("attribute error: scene.Page has no writable attributes")
}
func (p *Page) RunOperation(opType op.BinaryOpType, _ object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for scene.Page: %v", opType)
}
func (p *Page) GetAttr(name string) (object.Object, bool) { return pageAttrs.GetAttr(p, name) }

// CanvasObject lets window.SetContent(page) work.
func (p *Page) CanvasObject() fyne.CanvasObject { return p.view }

// View returns the renderer (Go hosts and tests).
func (p *Page) View() *render.View { return p.view }

// Node returns the script node handle for id.
func (p *Page) Node(id string) (*Node, bool) {
	if _, ok := p.view.Node(id); !ok {
		return nil, false
	}
	return &Node{page: p, id: id}, true
}

// Call implements render.Caller. Calls that arrive before page.Bind are
// buffered, because widget events can fire during the main body.
func (p *Page) Call(handler string, evt map[string]any) bool {
	p.mu.Lock()
	cb, ok := p.handlers[handler]
	bound := p.bound
	if !ok {
		if !bound && len(p.pending) < 256 {
			p.pending = append(p.pending, pendingCall{handler: handler, evt: cloneEvent(evt)})
			p.mu.Unlock()
			return true
		}
		p.mu.Unlock()
		return false
	}
	p.mu.Unlock()
	cb.Fire(objectFromAny(evt))
	return true
}

func (p *Page) bind(ctx context.Context, m *object.Map) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for name, val := range m.Value() {
		cb, err := vmexec.NewCallback(ctx, "page.Bind."+name, val)
		if err != nil {
			return err
		}
		p.handlers[name] = cb
	}
	p.bound = true
	for _, call := range p.pending {
		if cb, ok := p.handlers[call.handler]; ok {
			cb.Fire(objectFromAny(call.evt))
			continue
		}
		if p.cfg.Status != nil {
			p.cfg.Status(fmt.Sprintf("page.Bind: no handler %q for a buffered action", call.handler))
		}
	}
	p.pending = nil
	return nil
}

var pageAttrs = objattrs.Register(string(PageType), object.NewAttrRegistry[*Page](string(PageType)))

func init() {
	pageAttrs.Define("Root").Getter(func(p *Page) object.Object {
		if p.view.Doc().Root == nil {
			return object.Nil
		}
		return &Node{page: p, id: p.view.Doc().Root.ID}
	})
	pageAttrs.Define("Title").Getter(func(p *Page) object.Object {
		return object.NewString(p.view.Doc().Title)
	})
	pageAttrs.Define("Node").
		Doc("Return a node handle by ID: page.Node(\"door\").SetPlace({x: 10, y: 20}).").
		Arg("id").Returns("scene.Node").
		Impl(func(p *Page, _ context.Context, args ...object.Object) (object.Object, error) {
			id, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			n, ok := p.Node(id)
			if !ok {
				return nil, object.Errorf("page.Node: no node %q", id)
			}
			return n, nil
		})
	pageAttrs.Define("Find").
		Doc("Return node handles matching a map filter: page.Find({type: \"hotspot\"}).").
		Arg("filter").Returns("list").
		Impl(func(p *Page, _ context.Context, args ...object.Object) (object.Object, error) {
			m, ok := args[0].(*object.Map)
			if !ok {
				return nil, object.Errorf("page.Find: expected a map filter")
			}
			filter := make(map[string]string, len(m.Value()))
			for k, v := range m.Value() {
				s, err := object.AsString(v)
				if err != nil {
					return nil, object.Errorf("page.Find: filter %q must be a string", k)
				}
				filter[k] = s
			}
			var out []object.Object
			p.view.Doc().Walk(func(n *scenedoc.Node) bool {
				if matchFilter(n, filter) {
					out = append(out, &Node{page: p, id: n.ID})
				}
				return true
			})
			return object.NewList(out), nil
		})
	pageAttrs.Define("Bind").
		Doc("Register named handlers for declarative call actions: page.Bind({toggleMusic: (evt) => {...}}).").
		Arg("handlers").Returns("nil").
		Impl(func(p *Page, ctx context.Context, args ...object.Object) (object.Object, error) {
			m, ok := args[0].(*object.Map)
			if !ok {
				return nil, object.Errorf("page.Bind: expected a map of functions")
			}
			if err := p.bind(ctx, m); err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	pageAttrs.Define("Values").
		Doc("Values of every input node by ID (entry, check, select, radio, slider), for forms.").
		Returns("map").
		Impl(func(p *Page, ctx context.Context, _ ...object.Object) (object.Object, error) {
			out := make(map[string]object.Object)
			err := guithread.DoAndWaitContext(ctx, func() {
				p.view.Doc().Walk(func(n *scenedoc.Node) bool {
					obj, ok := p.view.LiveObject(n.ID)
					if !ok {
						return true
					}
					switch w := obj.(type) {
					case *widget.Entry:
						out[n.ID] = object.NewString(w.Text)
					case *widget.Check:
						out[n.ID] = object.NewBool(w.Checked)
					case *widget.Select:
						out[n.ID] = object.NewString(w.Selected)
					case *widget.RadioGroup:
						out[n.ID] = object.NewString(w.Selected)
					case *widget.Slider:
						out[n.ID] = object.NewFloat(w.Value)
					}
					return true
				})
			})
			if err != nil {
				return nil, err
			}
			return object.NewMap(out), nil
		})
}

func matchFilter(n *scenedoc.Node, filter map[string]string) bool {
	for k, want := range filter {
		switch strings.ToLower(k) {
		case "type":
			if n.Type != want {
				return false
			}
		case "name":
			if !strings.Contains(strings.ToLower(n.Name), strings.ToLower(want)) {
				return false
			}
		case "id":
			if n.ID != want {
				return false
			}
		case "tag", "tags":
			// Tags are a tie concept; a page host may store them in meta.
			found := false
			for _, v := range n.Props {
				if s, ok := v.(string); ok && strings.Contains(s, want) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// NodeType is the Risor type name of a scene node handle.
const NodeType object.Type = "scene.Node"

// Node is a script handle to one scene node. Unknown attributes delegate to
// the node's fynerisor widget wrapper.
type Node struct {
	page *Page
	id   string

	mu        sync.Mutex
	listeners map[string][]func()
}

func (n *Node) Type() object.Type      { return NodeType }
func (n *Node) Inspect() string        { return "scene.Node(" + n.id + ")" }
func (n *Node) Interface() interface{} { return n.id }
func (n *Node) IsTruthy() bool         { return true }
func (n *Node) Equals(o object.Object) bool {
	other, ok := o.(*Node)
	return ok && other.page == n.page && other.id == n.id
}
func (n *Node) Attrs() []object.AttrSpec { return nodeAttrs.Specs() }
func (n *Node) RunOperation(opType op.BinaryOpType, _ object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for scene.Node: %v", opType)
}

func (n *Node) node() (*scenedoc.Node, bool) {
	doc, ok := n.page.view.Node(n.id)
	return doc, ok
}

func (n *Node) GetAttr(name string) (object.Object, bool) {
	if o, ok := nodeAttrs.GetAttr(n, name); ok {
		return o, true
	}
	if w := n.page.wrapper(n.id); w != nil {
		return w.GetAttr(name)
	}
	return nil, false
}

func (n *Node) SetAttr(name string, value object.Object) error {
	switch name {
	case "Name":
		s, err := object.AsString(value)
		if err != nil {
			return err
		}
		return n.setName(s)
	case "Hidden":
		b, err := object.AsBool(value)
		if err != nil {
			return err
		}
		return n.setVisible(!b)
	}
	if w := n.page.wrapper(n.id); w != nil {
		return w.SetAttr(name, value)
	}
	return fmt.Errorf("attribute error: scene.Node has no attribute %q", name)
}

func (n *Node) setName(name string) error {
	var err error
	guithread.DoAndWait(func() {
		doc, ok := n.node()
		if !ok {
			err = fmt.Errorf("scene.Node: no node %q", n.id)
			return
		}
		doc.Name = name
	})
	return err
}

func (n *Node) setVisible(v bool) error {
	var err error
	guithread.DoAndWait(func() { n.page.view.SetVisible(n.id, v) })
	return err
}

var nodeAttrs = objattrs.Register(string(NodeType), object.NewAttrRegistry[*Node](string(NodeType)))

func init() {
	nodeAttrs.Define("ID").Getter(func(n *Node) object.Object { return object.NewString(n.id) })
	nodeAttrs.Define("Type").Getter(func(n *Node) object.Object {
		doc, _ := n.node()
		if doc == nil {
			return object.NewString("")
		}
		return object.NewString(doc.Type)
	})
	nodeAttrs.Define("Name").Getter(func(n *Node) object.Object {
		doc, _ := n.node()
		if doc == nil {
			return object.NewString("")
		}
		return object.NewString(doc.Name)
	})
	nodeAttrs.Define("Props").Getter(func(n *Node) object.Object {
		doc, _ := n.node()
		if doc == nil {
			return object.NewMap(map[string]object.Object{})
		}
		return objectFromAny(doc.Props)
	})
	nodeAttrs.Define("Place").Getter(func(n *Node) object.Object {
		doc, _ := n.node()
		if doc == nil {
			return object.NewMap(map[string]object.Object{})
		}
		return placeObject(doc.Place)
	})
	nodeAttrs.Define("Children").Getter(func(n *Node) object.Object {
		var out []object.Object
		if doc, ok := n.node(); ok {
			for _, c := range doc.Children {
				out = append(out, &Node{page: n.page, id: c.ID})
			}
		}
		return object.NewList(out)
	})
	nodeAttrs.Define("Parent").Getter(func(n *Node) object.Object {
		if p := n.page.view.ParentID(n.id); p != "" {
			return &Node{page: n.page, id: p}
		}
		return object.Nil
	})
	nodeAttrs.Define("Prop").
		Doc("Read a component property.").
		Arg("name").Returns("any").
		Impl(func(n *Node, _ context.Context, args ...object.Object) (object.Object, error) {
			key, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			doc, ok := n.node()
			if !ok {
				return nil, object.Errorf("scene.Node.Prop: no node %q", n.id)
			}
			if v, ok := doc.Prop(key); ok {
				return objectFromAny(v), nil
			}
			if spec, ok := n.page.cfg.Registry.Spec(doc.Type); ok {
				if p, ok := spec.Prop(key); ok {
					return objectFromAny(p.Default), nil
				}
			}
			return object.Nil, nil
		})
	nodeAttrs.Define("Set").
		Doc("Set a component property at runtime (not saved to the document).").
		Args("name", "value").Returns("nil").
		Impl(func(n *Node, ctx context.Context, args ...object.Object) (object.Object, error) {
			key, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			value := goFromObject(args[1])
			var ok bool
			err = guithread.DoAndWaitContext(ctx, func() {
				ok = n.page.view.RuntimeSet(n.id, key, value)
			})
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, object.Errorf("scene.Node.Set: no property %q on %q", key, n.id)
			}
			return object.Nil, nil
		})
	nodeAttrs.Define("SetPlace").
		Doc("Change a node's placement at runtime. Keys: x, y, w, h, minW, minH, slot, title, icon.").
		Arg("place").Returns("nil").
		Impl(func(n *Node, ctx context.Context, args ...object.Object) (object.Object, error) {
			doc, ok := n.node()
			if !ok {
				return nil, object.Errorf("scene.Node.SetPlace: no node %q", n.id)
			}
			p, err := placeFromObject(doc.Place, args[0])
			if err != nil {
				return nil, err
			}
			var applied bool
			err = guithread.DoAndWaitContext(ctx, func() {
				applied = n.page.view.RuntimePlace(n.id, p)
			})
			if err != nil {
				return nil, err
			}
			if !applied {
				return nil, object.Errorf("scene.Node.SetPlace: cannot place %q", n.id)
			}
			return object.Nil, nil
		})
	nodeAttrs.Define("SetName").
		Doc("Set the node's display name.").
		Arg("name").Returns("nil").
		Impl(func(n *Node, _ context.Context, args ...object.Object) (object.Object, error) {
			s, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			if err := n.setName(s); err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
	nodeAttrs.Define("Show").Doc("Show the node.").Returns("nil").
		Impl(func(n *Node, _ context.Context, _ ...object.Object) (object.Object, error) {
			return object.Nil, n.setVisible(true)
		})
	nodeAttrs.Define("Hide").Doc("Hide the node.").Returns("nil").
		Impl(func(n *Node, _ context.Context, _ ...object.Object) (object.Object, error) {
			return object.Nil, n.setVisible(false)
		})
	nodeAttrs.Define("On").
		Doc("Register an event listener. Listeners run after declarative actions.").
		Args("event", "callback").Returns("nil").
		Impl(func(n *Node, ctx context.Context, args ...object.Object) (object.Object, error) {
			event, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			cb, err := vmexec.NewCallback(ctx, "scene.Node.On."+event, args[1])
			if err != nil {
				return nil, err
			}
			off := n.page.view.On(n.id, event, func(evt map[string]any) {
				cb.Fire(objectFromAny(evt))
			})
			n.mu.Lock()
			if n.listeners == nil {
				n.listeners = make(map[string][]func())
			}
			n.listeners[event] = append(n.listeners[event], off)
			n.mu.Unlock()
			return object.Nil, nil
		})
	nodeAttrs.Define("Off").
		Doc("Remove listeners registered with On for an event.").
		Arg("event").Returns("nil").
		Impl(func(n *Node, _ context.Context, args ...object.Object) (object.Object, error) {
			event, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			n.mu.Lock()
			offs := n.listeners[event]
			delete(n.listeners, event)
			n.mu.Unlock()
			for _, off := range offs {
				off()
			}
			return object.Nil, nil
		})
	nodeAttrs.Define("SetContent").
		Doc("Fill a slot node with a widget or container at runtime.").
		Arg("object").Returns("nil").
		Impl(func(n *Node, ctx context.Context, args ...object.Object) (object.Object, error) {
			co, ok := args[0].(gui.IsCanvasObject)
			if !ok {
				return nil, object.Errorf("scene.Node.SetContent: expected a widget or container")
			}
			err := guithread.DoAndWaitContext(ctx, func() {
				if live, ok := n.page.view.Live(n.id); ok {
					if setter, ok := live.(interface{ SetContent(fyne.CanvasObject) }); ok {
						setter.SetContent(co.CanvasObject())
					}
				}
			})
			if err != nil {
				return nil, err
			}
			return object.Nil, nil
		})
}

// wrapper returns the fynerisor widget/canvas wrapper for a node, walking
// through the stack containers components use for event surfaces.
func (p *Page) wrapper(id string) object.Object {
	obj, ok := p.view.LiveObject(id)
	if !ok {
		return nil
	}
	var w risorwidget.WindowInterface = windowShim{}
	if p.cfg.Window != nil {
		w = p.cfg.Window
	}
	return wrapFyne(obj, w, 0)
}

func wrapFyne(obj fyne.CanvasObject, w risorwidget.WindowInterface, depth int) object.Object {
	if obj == nil || depth > 4 {
		return nil
	}
	if o := risorwidget.Wrap(obj, w); o != nil {
		return o
	}
	if o := risorcanvas.Wrap(obj); o != nil {
		return o
	}
	if c, ok := obj.(*fyne.Container); ok {
		for i := len(c.Objects) - 1; i >= 0; i-- {
			if o := wrapFyne(c.Objects[i], w, depth+1); o != nil {
				return o
			}
		}
	}
	return nil
}

type windowShim struct{}

func (windowShim) Do(fn func())         { guithread.Do(fn) }
func (windowShim) Report(string, error) {}

// ------------------------------------------------------------- conversions

func objectFromAny(v any) object.Object {
	switch x := v.(type) {
	case nil:
		return object.Nil
	case object.Object:
		return x
	case string:
		return object.NewString(x)
	case bool:
		return object.NewBool(x)
	case int:
		return object.NewInt(int64(x))
	case int64:
		return object.NewInt(x)
	case float64:
		return object.NewFloat(x)
	case float32:
		return object.NewFloat(float64(x))
	case []string:
		out := make([]object.Object, len(x))
		for i, s := range x {
			out[i] = object.NewString(s)
		}
		return object.NewList(out)
	case []any:
		out := make([]object.Object, len(x))
		for i, e := range x {
			out[i] = objectFromAny(e)
		}
		return object.NewList(out)
	case map[string]any:
		out := make(map[string]object.Object, len(x))
		for k, e := range x {
			out[k] = objectFromAny(e)
		}
		return object.NewMap(out)
	case map[string]string:
		out := make(map[string]object.Object, len(x))
		for k, e := range x {
			out[k] = object.NewString(e)
		}
		return object.NewMap(out)
	case scenedoc.Place:
		return placeObject(x)
	default:
		return object.NewString(fmt.Sprint(v))
	}
}

func goFromObject(o object.Object) any {
	switch x := o.(type) {
	case *object.String:
		return x.Value()
	case *object.Bool:
		return x.Value()
	case *object.Int:
		return x.Value()
	case *object.Float:
		return x.Value()
	case *object.List:
		out := make([]any, 0, len(x.Value()))
		for _, e := range x.Value() {
			out = append(out, goFromObject(e))
		}
		return out
	case *object.Map:
		out := make(map[string]any, len(x.Value()))
		for k, e := range x.Value() {
			out[k] = goFromObject(e)
		}
		return out
	case *object.NilType:
		return nil
	}
	return o.Interface()
}

func placeObject(p scenedoc.Place) object.Object {
	m := map[string]object.Object{
		"x": object.NewFloat(p.X), "y": object.NewFloat(p.Y),
		"w": object.NewFloat(p.W), "h": object.NewFloat(p.H),
		"minW": object.NewFloat(p.MinW), "minH": object.NewFloat(p.MinH),
		"slot": object.NewString(p.Slot), "title": object.NewString(p.Title),
		"icon": object.NewString(p.Icon),
	}
	if p.Anchor != nil {
		m["anchor"] = object.NewMap(map[string]object.Object{
			"h": object.NewString(p.Anchor.H),
			"v": object.NewString(p.Anchor.V),
		})
	}
	return object.NewMap(m)
}

func placeFromObject(base scenedoc.Place, o object.Object) (scenedoc.Place, error) {
	m, ok := o.(*object.Map)
	if !ok {
		return base, fmt.Errorf("scene.Node.SetPlace: expected a map")
	}
	p := base
	for k, v := range m.Value() {
		switch strings.ToLower(k) {
		case "x":
			p.X = numValue(v, p.X)
		case "y":
			p.Y = numValue(v, p.Y)
		case "w":
			p.W = numValue(v, p.W)
		case "h":
			p.H = numValue(v, p.H)
		case "minw":
			p.MinW = numValue(v, p.MinW)
		case "minh":
			p.MinH = numValue(v, p.MinH)
		case "slot":
			s, err := object.AsString(v)
			if err != nil {
				return p, err
			}
			p.Slot = s
		case "title":
			s, err := object.AsString(v)
			if err != nil {
				return p, err
			}
			p.Title = s
		case "icon":
			s, err := object.AsString(v)
			if err != nil {
				return p, err
			}
			p.Icon = s
		case "anchor":
			am, ok := v.(*object.Map)
			if !ok {
				continue
			}
			a := &scenedoc.Anchor{}
			if hv, ok := am.Value()["h"]; ok {
				a.H, _ = object.AsString(hv)
			}
			if vv, ok := am.Value()["v"]; ok {
				a.V, _ = object.AsString(vv)
			}
			if a.H == "" && a.V == "" {
				p.Anchor = nil
			} else {
				p.Anchor = a
			}
		}
	}
	return p, nil
}

func numValue(o object.Object, def float64) float64 {
	if f, err := object.AsFloat(o); err == nil {
		return f
	}
	return def
}

func cloneEvent(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// sortedKeys is a small helper used by tests and diagnostics.
func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
