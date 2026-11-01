// Package scene binds the scene runtime into Risor scripts. It is opt-in, like
// gui/db: import it and register the scene global with WithScene, or hand a
// page-scoped Scene to Window.ExecuteWith.
//
// Scripts get two kinds of objects:
//
//   - scene: the document namespace (Load, Parse, New).
//   - page: one loaded scene page, with Bind, Node, Find and Values.
//
// A scene node delegates attributes it does not define to the familiar
// fynerisor widget/canvas wrapper, so page.Node("title").SetText("x") works
// while scene-specific attributes (SetPlace, On, Prop) stay available.
package scene

import (
	"context"
	"fmt"
	"strings"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"

	"github.com/uidbz/fynerisor/gui"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
	"github.com/uidbz/fynerisor/gui/vmexec"
	scenedoc "github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

// Config configures the scene binding.
type Config struct {
	// Registry is the component registry (specs + factories). Nil uses the
	// standard built-in registry.
	Registry *render.Registry
	// Assets resolves scene assets. Nil uses file/http LocalAssets.
	Assets render.AssetResolver
	// Store loads (and, for editors, saves) scene documents. Nil picks a
	// FileStore or HTTPStore from the reference scheme.
	Store scenedoc.Store
	// Loader overrides document loading entirely (tie-browser uses it).
	Loader func(ctx context.Context, ref string) (data []byte, version string, err error)
	// Window provides widget callbacks and dialogs for node wrappers.
	Window *gui.Window
	// Navigator and Opener receive declarative navigation/media actions.
	Navigator render.Navigator
	Opener    render.Opener
	// Status receives runtime messages.
	Status func(string)
}

func (c Config) normalized() Config {
	if c.Registry == nil {
		c.Registry = renderbuiltin.NewRegistry()
	}
	if c.Assets == nil {
		c.Assets = render.NewLocalAssets()
	}
	return c
}

// Scene is the script-visible "scene" namespace.
type Scene struct {
	cfg  Config
	base string
}

// NewScene returns a scene namespace. base is the document/script URL used to
// resolve relative scene references.
func NewScene(cfg Config, base string) *Scene {
	return &Scene{cfg: cfg.normalized(), base: base}
}

// WithScene registers the scene namespace as the global "scene" (and makes
// require(["@scene"]) succeed).
func WithScene(cfg Config) gui.Option {
	return gui.WithGlobal("scene", NewScene(cfg, ""))
}

const SceneType object.Type = "scene"

func (s *Scene) Type() object.Type           { return SceneType }
func (s *Scene) Inspect() string             { return "scene" }
func (s *Scene) Interface() interface{}      { return nil }
func (s *Scene) IsTruthy() bool              { return true }
func (s *Scene) Equals(o object.Object) bool { return s == o }
func (s *Scene) Attrs() []object.AttrSpec    { return sceneAttrs.Specs() }
func (s *Scene) SetAttr(string, object.Object) error {
	return fmt.Errorf("attribute error: scene object has no attribute %q", "set")
}
func (s *Scene) RunOperation(opType op.BinaryOpType, _ object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for scene: %v", opType)
}
func (s *Scene) GetAttr(name string) (object.Object, bool) { return sceneAttrs.GetAttr(s, name) }

var sceneAttrs = objattrs.Register(string(SceneType), object.NewAttrRegistry[*Scene](string(SceneType)))

func init() {
	sceneAttrs.Define("Load").
		Doc("Load a scene document relative to this page and return a page object: window.SetContent(scene.Load(\"living-room.scene.json\")).").
		Arg("ref").Returns("scene.Page").
		Impl(func(s *Scene, ctx context.Context, args ...object.Object) (object.Object, error) {
			ref, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			return s.load(ctx, ref)
		})
	sceneAttrs.Define("Parse").
		Doc("Parse a scene document from JSON.").
		Arg("json").Returns("scene.Doc").
		Impl(func(s *Scene, ctx context.Context, args ...object.Object) (object.Object, error) {
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			doc, err := scenedoc.Decode([]byte(text))
			if err != nil {
				return nil, object.NewError(err)
			}
			return &Doc{doc: doc, cfg: s.cfg, base: s.base}, nil
		})
	sceneAttrs.Define("New").
		Doc("Create a blank scene document with a free root.").
		Returns("scene.Doc").
		Impl(func(s *Scene, _ context.Context, _ ...object.Object) (object.Object, error) {
			doc := &scenedoc.Document{
				Format: scenedoc.Format,
				Title:  "Untitled scene",
				Root: &scenedoc.Node{
					ID: "root", Type: "free",
					Props: map[string]any{"designW": 1600.0, "designH": 900.0, "scale": "fit"},
				},
			}
			return &Doc{doc: doc, cfg: s.cfg, base: s.base}, nil
		})
}

func (s *Scene) load(ctx context.Context, ref string) (object.Object, error) {
	resolved, err := s.resolve(ref)
	if err != nil {
		return nil, err
	}
	data, version, err := s.loadBytes(ctx, resolved)
	if err != nil {
		return nil, err
	}
	doc, err := scenedoc.Decode(data)
	if err != nil {
		return nil, object.NewError(err)
	}
	issues := scenedoc.Validate(doc, s.cfg.Registry.Registry)
	if scenedoc.Errors(issues) > 0 {
		return nil, object.Errorf("scene %s has %d validation errors: %s", resolved, scenedoc.Errors(issues), firstIssue(issues))
	}
	page, err := newPage(ctx, doc, s.cfg, resolved, version)
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (s *Scene) resolve(ref string) (string, error) {
	if s.cfg.Store != nil {
		return s.cfg.Store.Resolve(s.base, ref)
	}
	base := s.base
	if base == "" && s.cfg.Window != nil {
		// scripts with no page base resolve against the process working
		// directory, like import().
	}
	return scenedoc.ResolveRef(base, ref)
}

func (s *Scene) loadBytes(ctx context.Context, ref string) ([]byte, string, error) {
	if s.cfg.Loader != nil {
		return s.cfg.Loader(ctx, ref)
	}
	if s.cfg.Store != nil {
		snap, err := s.cfg.Store.Load(ctx, ref)
		if err != nil {
			return nil, "", err
		}
		return snap.Data, snap.Version, nil
	}
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		hs := &scenedoc.HTTPStore{}
		snap, err := hs.Load(ctx, ref)
		if err != nil {
			return nil, "", err
		}
		return snap.Data, snap.Version, nil
	}
	snap, err := (scenedoc.FileStore{}).Load(ctx, ref)
	if err != nil {
		return nil, "", err
	}
	return snap.Data, snap.Version, nil
}

func firstIssue(issues []scenedoc.Issue) string {
	for _, i := range issues {
		if i.Severity == scenedoc.Error {
			return i.String()
		}
	}
	return ""
}

// BuildPage builds a page from an already decoded document. It is the entry
// point for hosts (the browser's scene renderer) that load documents
// themselves.
func BuildPage(ctx context.Context, doc *scenedoc.Document, cfg Config, base, version string) (*Page, error) {
	return newPage(ctx, doc, cfg.normalized(), base, version)
}

// newPage builds a page on the UI thread. Widgets must be created there once a
// Fyne app is running.
func newPage(ctx context.Context, doc *scenedoc.Document, cfg Config, base, version string) (*Page, error) {
	var view *render.View
	err := guithread.DoAndWaitContext(ctx, func() {
		view = render.NewView(doc, render.Options{
			Registry:  cfg.Registry,
			Assets:    cfg.Assets,
			Mode:      render.ModePlay,
			Base:      base,
			Navigator: cfg.Navigator,
			Opener:    cfg.Opener,
			Status:    cfg.Status,
		})
	})
	if err != nil {
		return nil, err
	}
	page := &Page{
		view:     view,
		cfg:      cfg,
		base:     base,
		version:  version,
		handlers: make(map[string]*vmexec.Callback),
		bound:    false,
	}
	view.SetCaller(page)
	return page, nil
}

// Doc is the script-visible document object returned by scene.Parse/New.
type Doc struct {
	doc  *scenedoc.Document
	cfg  Config
	base string
}

const DocType object.Type = "scene.Doc"

func (d *Doc) Type() object.Type           { return DocType }
func (d *Doc) Inspect() string             { return "scene.Doc" }
func (d *Doc) Interface() interface{}      { return d.doc }
func (d *Doc) IsTruthy() bool              { return true }
func (d *Doc) Equals(o object.Object) bool { return d == o }
func (d *Doc) Attrs() []object.AttrSpec    { return docAttrs.Specs() }
func (d *Doc) SetAttr(string, object.Object) error {
	return fmt.Errorf("attribute error: scene.Doc has no writable attributes")
}
func (d *Doc) RunOperation(opType op.BinaryOpType, _ object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for scene.Doc: %v", opType)
}
func (d *Doc) GetAttr(name string) (object.Object, bool) { return docAttrs.GetAttr(d, name) }

var docAttrs = objattrs.Register(string(DocType), object.NewAttrRegistry[*Doc](string(DocType)))

func init() {
	docAttrs.Define("Title").Getter(func(d *Doc) object.Object {
		if d.doc == nil {
			return object.NewString("")
		}
		return object.NewString(d.doc.Title)
	})
	docAttrs.Define("Encode").
		Doc("Canonical JSON for this document.").
		Returns("string").
		Impl(func(d *Doc, _ context.Context, _ ...object.Object) (object.Object, error) {
			data, err := scenedoc.Encode(d.doc)
			if err != nil {
				return nil, object.NewError(err)
			}
			return object.NewString(string(data)), nil
		})
	docAttrs.Define("Validate").
		Doc("Validation messages as a list of strings.").
		Returns("list").
		Impl(func(d *Doc, _ context.Context, _ ...object.Object) (object.Object, error) {
			issues := scenedoc.Validate(d.doc, d.cfg.Registry.Registry)
			out := make([]object.Object, len(issues))
			for i, issue := range issues {
				out[i] = object.NewString(issue.String())
			}
			return object.NewList(out), nil
		})
	docAttrs.Define("Page").
		Doc("Build a live page from this document.").
		Returns("scene.Page").
		Impl(func(d *Doc, ctx context.Context, _ ...object.Object) (object.Object, error) {
			return newPage(ctx, d.doc, d.cfg, d.base, "")
		})
}

var _ gui.IsCanvasObject = (*Page)(nil)
