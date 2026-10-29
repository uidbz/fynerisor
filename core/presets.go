package core

// Capability presets (issue #40): hosts state policy — what a script
// may touch — instead of accepting the all-or-nothing per-module grants.

import (
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"

	csvmod "github.com/uidbz/fynerisor/modules/csv"
	"github.com/uidbz/fynerisor/modules/exec"
	filepathmod "github.com/uidbz/fynerisor/modules/filepath"
	"github.com/uidbz/fynerisor/modules/http"
	iomod "github.com/uidbz/fynerisor/modules/io"
	jsonmod "github.com/uidbz/fynerisor/modules/json"
	osmod "github.com/uidbz/fynerisor/modules/os"
	stringsmod "github.com/uidbz/fynerisor/modules/strings"
	tiemod "github.com/uidbz/fynerisor/modules/tie"
	timemod "github.com/uidbz/fynerisor/modules/time"
)

// Safe enables the pure data modules: strings, filepath, time, json and
// csv. No file, network, process or OS access.
func Safe() Option {
	return NewOption(func(c *Config) {
		applyAll(c,
			Module("strings", stringsmod.Module()),
			Module("filepath", filepathmod.Module()),
			Module("time", timemod.Module()),
			Module("json", jsonmod.Module()),
			Module("csv", csvmod.Module()),
		)
	})
}

// Files enables the file-system modules (io, os). When readOnly is true
// the mutating operations are *absent* — io.write/copy/cp and
// os.write_file/create/copy_file/remove_all/mkdir/mkdir_all/exec/
// open_browser raise "attribute not found" in scripts, so a read-only
// grant cannot be misused rather than merely being discouraged.
func Files(readOnly bool) Option {
	return NewOption(func(c *Config) {
		if readOnly {
			applyAll(c,
				Module("io", denyAttrs(iomod.NewIO(), "write", "copy", "cp")),
				Module("os", denyAttrs(osmod.Module(),
					"write_file", "create", "copy_file", "remove_all",
					"mkdir", "mkdir_all", "exec", "open_browser")),
			)
			return
		}
		applyAll(c,
			Module("io", iomod.NewIO()),
			Module("os", osmod.Module()),
		)
	})
}

// Exec enables the exec module (processes are a separate capability
// from file access; neither Safe nor Files grants it).
func Exec() Option {
	return Module("exec", exec.Module())
}

// HTTP enables outbound HTTP (and optionally HTTP URL imports, which
// fetch remote *code* and stay a separate decision, see WithHTTPImport).
func HTTP() Option {
	return Module("http", http.Module())
}

// Full enables everything the fynerisor CLI enables by default (except
// SQL, which is driver-linking and stays opt-in via core/db or gui/db).
func Full() Option {
	return NewOption(func(c *Config) {
		applyAll(c,
			HTTP(),
			markEnabled("httpimport"), // require-validatable, no global
			Files(false),
			Exec(),
			Safe(),
			Module("tie", tiemod.Module()),
		)
	})
}

// markEnabled makes a feature validatable with require(["@name"])
// without contributing a global (features implemented in the runtime).
func markEnabled(name string) Option {
	return NewOption(func(c *Config) { c.EnabledModules[name] = true })
}

func applyAll(c *Config, opts ...Option) {
	for _, o := range opts {
		o.ApplyTo(c)
	}
}

// deniedObject hides a fixed set of attributes from a module object —
// the enforcement point for capability presets. Script lookups of a
// denied attribute fail like any unknown name ("attribute error"), and
// the Attrs listing stays truthful.
type deniedObject struct {
	inner  object.Object
	denied map[string]bool
}

func denyAttrs(inner object.Object, names ...string) object.Object {
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	return &deniedObject{inner: inner, denied: set}
}

func (d *deniedObject) Type() object.Type { return d.inner.Type() }
func (d *deniedObject) Inspect() string   { return d.inner.Inspect() }
func (d *deniedObject) Interface() any    { return d.inner.Interface() }
func (d *deniedObject) IsTruthy() bool    { return d.inner.IsTruthy() }
func (d *deniedObject) Equals(other object.Object) bool {
	if o, ok := other.(*deniedObject); ok {
		return d.inner.Equals(o.inner)
	}
	return d.inner.Equals(other)
}
func (d *deniedObject) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return d.inner.RunOperation(opType, right)
}
func (d *deniedObject) SetAttr(name string, value object.Object) error {
	if d.denied[name] {
		return object.Errorf("attribute error: %s object has no attribute %q", d.inner.Type(), name)
	}
	return d.inner.SetAttr(name, value)
}
func (d *deniedObject) GetAttr(name string) (object.Object, bool) {
	if d.denied[name] {
		return nil, false
	}
	return d.inner.GetAttr(name)
}
func (d *deniedObject) Attrs() []object.AttrSpec {
	specs := d.inner.Attrs()
	if specs == nil {
		return nil
	}
	out := make([]object.AttrSpec, 0, len(specs))
	for _, s := range specs {
		if !d.denied[s.Name] {
			out = append(out, s)
		}
	}
	return out
}
