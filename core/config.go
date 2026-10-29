package core

import (
	"io"

	"github.com/deepnoodle-ai/risor/v2"
)

// Issue #40: one option mechanism for every runtime.
//
// Options assemble a neutral Config; core.Context and gui.Window each
// consume it in a single ordered pass. Because Option and Config are
// exported, a core.Option is accepted verbatim by gui.NewWindow and a
// gui-defined option by core.NewContext — policy (which preset) lives
// in the host, mechanism here.

// Config is the assembled runtime configuration: everything needed to
// build a script runtime that carries no Fyne types.
type Config struct {
	// Globals are script-visible globals contributed by options (module
	// objects from Module(), user values from WithGlobal).
	Globals map[string]any
	// UserVMOptions are opaque risor.Option VM-tuning values (see
	// WithRisorOptions); they are not forwarded to imported module VMs.
	UserVMOptions []risor.Option
	// EnabledModules names the require()-validatable modules.
	EnabledModules map[string]bool

	// AppName/appVersion surface to scripts as app.name/app.version.
	AppName    string
	AppVersion string
	// ScriptArgs populates os.args for the runtime (per-runtime since #44).
	ScriptArgs []string

	// Text streams for the runtime (#24/#42); nil keeps the per-runtime
	// default (the process streams for the first runtime, never shared
	// state).
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// Host hooks (plain Go funcs so core stays Fyne-free; gui consumes
	// them, core ignores what it has no concept for).
	StatusCallback func(status string)
	ResultCallback func(result string)
	ErrorHandler   func(err error)

	// HostWindow is the *fyne.Window an embedded view lives in; boxed as
	// any so this package stays Fyne-free. gui.NewCanvasObject consumes
	// it; NewWindow ignores it (it receives its window as a parameter).
	HostWindow any
}

// NewConfig returns a Config with the library defaults applied.
func NewConfig() *Config {
	return &Config{
		Globals:        make(map[string]any),
		EnabledModules: make(map[string]bool),
		AppName:        "fynerisor",
	}
}

// Option configures a runtime during creation. The same Option values
// work for core.NewContext and gui.NewWindow/NewCanvasObject.
type Option interface {
	ApplyTo(*Config)
}

type optionFunc func(*Config)

func (f optionFunc) ApplyTo(c *Config) { f(c) }

// NewOption wraps a configuration step as an Option. Hosts use it for
// app-specific settings; library options (WithHTTP, WithStdout, ...) are
// built the same way.
func NewOption(f func(*Config)) Option {
	return optionFunc(f)
}

// Module registers a script-visible global (typically a module object)
// and makes it validatable with require(["@name"]). WithHTTP() is
// Module("http", http.Module()); hosts use it for their own APIs.
func Module(name string, value any) Option {
	return NewOption(func(c *Config) {
		c.Globals[name] = value
		c.EnabledModules[name] = true
	})
}

// Apply assembles opts into a Config starting from NewConfig defaults,
// in order, once (issue #40: no type-switch passes, no silently ignored
// option types).
func Apply(opts ...Option) *Config {
	cfg := NewConfig()
	for _, opt := range opts {
		opt.ApplyTo(cfg)
	}
	return cfg
}
