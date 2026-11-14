package v3scene

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	risor "github.com/deepnoodle-ai/risor/go/v3"

	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

// Mode selects how a page reaches the world. Pick one per page; the prototype
// deliberately does not mix them inside one demo.
type Mode string

const (
	// ModeActions turns side effects into data. The program returns fetch or
	// save actions; the host performs them and, for fetch, sends a loaded
	// event back. Nothing in the program blocks, and a run can be replayed.
	ModeActions Mode = "actions"

	// ModeHostFunctions registers store.customers and store.add. The program
	// calls them during the run; they are synchronous and may block the run.
	ModeHostFunctions Mode = "host-functions"
)

// Config configures one page.
type Config struct {
	Document *scene.Document
	Script   string
	Mode     Mode
	Store    *Store
	Base     string
	Status   func(string)
	Report   func(error)
}

// App is one compiled page: an immutable v3 program, a scene view, and the
// host state. Events are serialized through Step; the compiled program may be
// shared by goroutines, but this App keeps the page state single-threaded.
type App struct {
	cfg  Config
	view *render.View
	prog *risor.Program

	mu    sync.Mutex
	state State

	events chan map[string]any
	done   chan struct{}
	cancel context.CancelFunc
	once   sync.Once
}

// Check compiles the page and discards it. It is the headless path used by
// `v3scene -check` and CI; it needs no Fyne app.
func Check(cfg Config) error {
	cfg = normalize(cfg)
	_, err := compile(cfg)
	return err
}

// New compiles the page and builds its scene view. A type or undefined-name
// error fails here, before any event runs.
func New(cfg Config) (*App, error) {
	cfg = normalize(cfg)
	prog, err := compile(cfg)
	if err != nil {
		return nil, err
	}
	view := render.NewView(cfg.Document, render.Options{
		Registry: renderbuiltin.NewRegistry(),
		Base:     cfg.Base,
		Status:   cfg.Status,
	})
	return &App{
		cfg:    cfg,
		view:   view,
		prog:   prog,
		events: make(chan map[string]any, 128),
		done:   make(chan struct{}),
	}, nil
}

func normalize(cfg Config) Config {
	if cfg.Mode == "" {
		cfg.Mode = ModeActions
	}
	if cfg.Store == nil {
		cfg.Store = NewStore()
	}
	if cfg.Status == nil {
		cfg.Status = func(string) {}
	}
	if cfg.Report == nil {
		cfg.Report = func(err error) { cfg.Status("ERROR: " + err.Error()) }
	}
	return cfg
}

func compile(cfg Config) (*risor.Program, error) {
	if cfg.Document == nil {
		return nil, errors.New("v3scene: nil document")
	}
	decls := []risor.Decl{
		risor.NamedType("Customer", typeCustomer),
		risor.NamedType("State", typeState),
		risor.NamedType("Event", typeEvent),
		risor.NamedType("Action", typeAction),
		risor.NamedType("Result", typeResult),
		risor.Var("state", "State"),
		risor.Var("event", "Event"),
	}
	if cfg.Mode == ModeHostFunctions {
		decls = append(decls, storeFunctions(cfg.Store)...)
	} else if cfg.Mode != ModeActions {
		return nil, fmt.Errorf("v3scene: unknown mode %q", cfg.Mode)
	}
	env, err := risor.NewEnv(decls...)
	if err != nil {
		return nil, fmt.Errorf("v3scene: environment: %w", err)
	}
	expr, err := risor.Parse(cfg.Script)
	if err != nil {
		return nil, fmt.Errorf("v3scene: parse: %w", err)
	}
	prog, err := env.Compile(expr, risor.Expect("Result"))
	if err != nil {
		return nil, fmt.Errorf("v3scene: check: %w", err)
	}
	return prog, nil
}

// storeFunctions is the ModeHostFunctions capability. Signatures are part of
// the environment, so a script that calls store.nope fails to compile.
func storeFunctions(s *Store) []risor.Decl {
	return []risor.Decl{
		risor.Function{
			Name: "store.customers",
			Type: "() => Customer[]",
			Impl: func(_ *risor.Call, _ []any) (any, error) {
				return s.Customers(), nil
			},
		},
		risor.Function{
			Name: "store.add",
			Type: "(name: string, city: string) => Customer !",
			Impl: func(_ *risor.Call, args []any) (any, error) {
				name, _ := args[0].(string)
				city, _ := args[1].(string)
				return s.Add(name, city)
			},
		},
	}
}

// CanvasObject returns the scene view, ready for window.SetContent.
func (a *App) CanvasObject() fyne.CanvasObject { return a.view }

// View exposes the renderer (tests and hosts that need node inspection).
func (a *App) View() *render.View { return a.view }

// State returns a snapshot of the host-owned state.
func (a *App) State() State {
	a.mu.Lock()
	defer a.mu.Unlock()
	return cloneState(a.state)
}

// Start registers listeners, starts the event goroutine and posts init. Call
// it on the UI thread (or before the Fyne main loop starts).
func (a *App) Start(ctx context.Context) {
	ctx, a.cancel = context.WithCancel(ctx)
	a.registerListeners()
	go a.loop(ctx)
	a.Dispatch(map[string]any{"type": "init"})
}

func (a *App) registerListeners() {
	events := append([]string(nil), scene.Events...)
	events = append(events, "cellTapped")
	seen := map[string]bool{}
	a.cfg.Document.Walk(func(n *scene.Node) bool {
		for _, event := range events {
			if seen[n.ID+"\x00"+event] {
				continue
			}
			seen[n.ID+"\x00"+event] = true
			id, ev := n.ID, event
			a.view.On(id, ev, func(payload map[string]any) {
				a.enqueue(id, ev, payload)
			})
		}
		return true
	})
}

// enqueue runs on the UI thread. It never blocks Fyne: a full queue drops the
// event and reports it, because a frozen UI is worse than a lost coalesced
// change (this is a prototype; a real host would use the existing unbounded
// queue in vmexec).
func (a *App) enqueue(id, event string, payload map[string]any) {
	evt := make(map[string]any, len(payload)+3)
	for k, v := range payload {
		evt[k] = v
	}
	evt["type"] = event
	evt["node"] = id
	if values := readValues(a.view); len(values) > 0 {
		evt["values"] = values
	}
	select {
	case a.events <- evt:
	case <-a.done:
	default:
		a.cfg.Report(fmt.Errorf("v3scene: event queue full, dropped %s on %s", event, id))
	}
}

// Dispatch posts an event from the host. It is non-blocking and is also the
// way a real host resumes after an async fetch.
func (a *App) Dispatch(event map[string]any) {
	if event == nil {
		return
	}
	select {
	case a.events <- event:
	case <-a.done:
	}
}

func (a *App) loop(ctx context.Context) {
	defer close(a.done)
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-a.events:
			if err := a.Step(ctx, evt); err != nil {
				a.cfg.Report(err)
			}
		}
	}
}

// Step runs one event and applies its actions. It is synchronous and is what
// the tests use; the event loop calls it one event at a time.
func (a *App) Step(ctx context.Context, event map[string]any) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.step(ctx, event, 0)
}

func (a *App) step(ctx context.Context, event map[string]any, depth int) error {
	if depth > 16 {
		return errors.New("v3scene: effect chain deeper than 16 (fetch loop?)")
	}
	out, err := a.prog.Run(ctx, risor.Map{
		"state": a.state,
		"event": event,
	})
	if err != nil {
		return err
	}
	var result Result
	if err := risor.Decode(out, &result); err != nil {
		return fmt.Errorf("v3scene: result does not match Result: %w", err)
	}
	a.state = cloneState(result.State)
	return a.apply(ctx, result.Actions, depth)
}

func (a *App) apply(ctx context.Context, actions []Action, depth int) error {
	for _, action := range actions {
		switch action.Do {
		case "fetch":
			if a.cfg.Mode != ModeActions {
				return fmt.Errorf("v3scene: fetch action is not available in %s mode", a.cfg.Mode)
			}
			data := a.cfg.Store.Customers()
			if err := a.step(ctx, map[string]any{
				"type": "loaded",
				"node": action.Node,
				"data": data,
			}, depth+1); err != nil {
				return err
			}
		case "save":
			if a.cfg.Mode != ModeActions {
				return fmt.Errorf("v3scene: save action is not available in %s mode", a.cfg.Mode)
			}
			if action.Customer == nil {
				return errors.New("v3scene: save action without a customer")
			}
			a.cfg.Store.Save(*action.Customer)
		case "status":
			a.cfg.Status(action.Text)
		default:
			if err := a.applyScene(ctx, action); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) applyScene(ctx context.Context, action Action) error {
	var applyErr error
	doErr := guithread.DoAndWaitContext(ctx, func() {
		l, ok := a.view.Live(action.Node)
		switch action.Do {
		case "set":
			if !a.view.RuntimeSet(action.Node, action.Prop, action.Value) {
				applyErr = fmt.Errorf("set %s.%s: no such node or property", action.Node, action.Prop)
			}
		case "show":
			a.view.SetVisible(action.Node, true)
		case "hide":
			a.view.SetVisible(action.Node, false)
		case "toggle":
			a.view.SetVisible(action.Node, !a.view.IsVisible(action.Node))
		case "rows":
			rm, ok := render.RowModelOf(l)
			if !ok {
				applyErr = fmt.Errorf("rows %s: not a row model", action.Node)
				return
			}
			rm.SetRows(action.Rows)
		case "append":
			ap, ok := render.AppenderOf(l)
			if !ok {
				applyErr = fmt.Errorf("append %s: not an appender", action.Node)
				return
			}
			ap.AppendText(action.Text)
		default:
			applyErr = fmt.Errorf("unknown action %q", action.Do)
		}
		if !ok && applyErr == nil {
			applyErr = fmt.Errorf("%s %s: node not found", action.Do, action.Node)
		}
	})
	if doErr != nil {
		return doErr
	}
	return applyErr
}

// Stop ends the event loop.
func (a *App) Stop() {
	a.once.Do(func() {
		if a.cancel != nil {
			a.cancel()
		} else {
			close(a.done)
		}
	})
}

func cloneState(s State) State {
	s.Customers = append([]Customer(nil), s.Customers...)
	s.Log = append([]string(nil), s.Log...)
	return s
}

// readValues gathers input-node values for the event's values record. It runs
// on the UI thread inside a listener.
func readValues(v *render.View) map[string]any {
	out := map[string]any{}
	v.Doc().Walk(func(n *scene.Node) bool {
		obj, ok := v.LiveObject(n.ID)
		if !ok {
			return true
		}
		switch w := obj.(type) {
		case *widget.Entry:
			out[n.ID] = w.Text
		case *widget.Check:
			out[n.ID] = w.Checked
		case *widget.Select:
			out[n.ID] = w.Selected
		case *widget.RadioGroup:
			out[n.ID] = w.Selected
		case *widget.Slider:
			out[n.ID] = w.Value
		}
		return true
	})
	return out
}

// LoadDocument reads a scene document from disk.
func LoadDocument(path string) (*scene.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return scene.Decode(data)
}
