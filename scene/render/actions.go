package render

import (
	"fmt"

	"github.com/uidbz/fynerisor/scene"
)

// Engine executes declarative actions and dispatches script listeners. It runs
// on the UI thread; script listeners must enqueue their work on the VM
// goroutine and return immediately.
type Engine struct {
	view *View
}

// Run handles one event for a node: declarative actions in document order,
// then the script listeners registered for that node and event.
func (e *Engine) Run(n *scene.Node, event string, evt map[string]any) {
	v := e.view
	if v == nil || n == nil || v.dead {
		return
	}
	actions := n.On[event]
	for _, a := range actions {
		e.runAction(n, a, evt)
	}
	// A scene link without an explicit action navigates to its target. This
	// keeps point-and-click pages terse: a link node with `to` just works.
	if len(actions) == 0 && n.Type == "link" && event == "tapped" {
		if to := n.StringProp("to", ""); to != "" {
			e.navigate(to)
		}
	}
	v.mu.Lock()
	fns := append([]func(map[string]any){}, v.listeners[n.ID][event]...)
	v.mu.Unlock()
	for _, fn := range fns {
		if fn != nil {
			fn(evt)
		}
	}
}

func (e *Engine) runAction(n *scene.Node, a scene.Action, evt map[string]any) {
	v := e.view
	target := a.Node
	if target == "" || target == "self" {
		target = n.ID
	}
	switch a.Do {
	case "navigate":
		e.navigate(a.To)
	case "back":
		if v.opts.Navigator != nil {
			v.opts.Navigator.Back()
		} else {
			e.status("back: no navigator")
		}
	case "show":
		v.SetVisible(target, true)
	case "hide":
		v.SetVisible(target, false)
	case "toggle":
		v.SetVisible(target, !v.IsVisible(target))
	case "set":
		if !v.RuntimeSet(target, a.Prop, a.Value) {
			e.status(fmt.Sprintf("set %s.%s: no such node or property", target, a.Prop))
		}
	case "call":
		if v.opts.Caller == nil {
			e.status(fmt.Sprintf("call %q: no script bound", a.Handler))
			return
		}
		if !v.opts.Caller.Call(a.Handler, evt) {
			e.status(fmt.Sprintf("call %q: handler not bound", a.Handler))
		}
	case "open":
		ref := a.Ref
		if ref == "" {
			ref, _ = evt["item"].(string)
		}
		if ref == "" {
			ref = a.To
		}
		if ref == "" {
			e.status("open: nothing to open")
			return
		}
		if v.opts.Opener != nil {
			v.opts.Opener.Open(ref)
		} else {
			e.status("open: no opener")
		}
	case "play":
		ref := firstRef(a, evt)
		if v.opts.Opener != nil {
			v.opts.Opener.Play(ref)
		}
	case "pause":
		if v.opts.Opener != nil {
			v.opts.Opener.Pause()
		}
	case "select":
		e.status("select: editor-only action")
	default:
		e.status(fmt.Sprintf("unknown action %q", a.Do))
	}
}

func (e *Engine) navigate(to string) {
	if to == "" {
		e.status("navigate: empty target")
		return
	}
	ref := to
	if e.view.opts.Base != "" {
		if resolved, err := scene.ResolveRef(e.view.opts.Base, to); err == nil {
			ref = resolved
		}
	}
	if e.view.opts.Navigator == nil {
		e.status("navigate: no navigator (preview mode?)")
		return
	}
	e.view.opts.Navigator.Navigate(ref)
}

func (e *Engine) status(msg string) {
	if e.view != nil && e.view.opts.Status != nil {
		e.view.opts.Status(msg)
	}
}

func firstRef(a scene.Action, evt map[string]any) string {
	if a.Ref != "" {
		return a.Ref
	}
	if s, ok := evt["item"].(string); ok && s != "" {
		return s
	}
	return a.To
}
