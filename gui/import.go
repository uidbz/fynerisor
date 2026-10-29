package gui

import (
	"context"
	"strings"
	"sync"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/vm"
	"github.com/uidbz/fynerisor/core"
)

// importState is the module system state of one script run: the entry
// script's base, the module cache and the in-progress import stack.
//
// It is per run, not per window (issue #16): a stopped run whose VM is still
// unwinding cannot push onto the import stack of the run that replaced it,
// and each Execute re-imports modules, so a reload (--watch, browser
// refresh) picks up edited modules and never shares module state — or
// callbacks registered by module code — with a stopped run.
type importState struct {
	mu    sync.Mutex
	base  string // entry script base (dir or URL); "" = process CWD
	cache map[string]*core.ImportedModule
	stack []core.ImportFrame
	env   map[string]any // the run's environment snapshot (module VM globals)
}

type importStateKey struct{}

func newImportState(base string, env map[string]any) *importState {
	return &importState{base: base, cache: map[string]*core.ImportedModule{}, env: env}
}

// importStateFor returns the import state of the run that owns ctx. A
// builtin called outside any run (not reachable from Execute) gets a
// throwaway state with the window's current script base.
func (w *Window) importStateFor(ctx context.Context) *importState {
	if st, ok := ctx.Value(importStateKey{}).(*importState); ok {
		return st
	}
	_, base := w.runner.Snapshot()
	env, _ := w.runOptions()
	return newImportState(base, env)
}

// newImportBuiltin creates the import() function that loads Risor modules
// and returns them as ImportedModule objects with namespace access.
//
// This implements proper module scoping where each import creates an isolated
// namespace. All top-level variables and functions defined in the imported
// module are accessible via dot notation.
//
// Features:
// - Module caching: Same path returns same module instance
// - Circular import detection: Prevents infinite loops
// - Global inheritance: Modules have access to http, sql, widget, etc.
// - File and URL support: Load from local files or HTTP(S) URLs
//
// Example:
//
//	// math_utils.risor
//	let add = (a, b) => a + b
//	let multiply = (a, b) => a * b
//	let PI = 3.14159
//
//	// main.risor
//	let math = import("math_utils.risor")
//	print(math.add(5, 3))      // 8
//	print(math.multiply(4, 2)) // 8
//	print(math.PI)             // 3.14159
func (w *Window) newImportBuiltin() *object.Builtin {
	return object.NewBuiltin("import", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		// Validate arguments
		if len(args) != 1 {
			return nil, object.Errorf("type error: import() requires exactly 1 argument, got %d", len(args))
		}

		path, ok := args[0].(*object.String)
		if !ok {
			return nil, object.Errorf("type error: import() requires string argument, got %s", args[0].Type())
		}

		modulePath := path.Value()
		st := w.importStateFor(ctx)

		// Resolve against the importing script's base (the entry script's
		// base, or the base of the module at the top of the import stack)
		// and enforce the remote same-origin policy (issue #8). Same
		// resolver and policy as core.
		st.mu.Lock()
		base := st.base
		if n := len(st.stack); n > 0 {
			base = st.stack[n-1].Base
		}
		st.mu.Unlock()
		baseIsRemote := core.IsHTTPURL(base)

		key, err := core.ResolveImportPath(modulePath, base)
		if err != nil {
			return nil, object.Errorf("import error: %w", err)
		}

		// Thread-safe cache check (keyed by the canonical path/URL)
		st.mu.Lock()
		if module, found := st.cache[key]; found {
			st.mu.Unlock()
			return module, nil
		}

		// Check for circular imports
		keys := make([]string, len(st.stack))
		for i, f := range st.stack {
			keys[i] = f.Key
		}
		for _, importing := range keys {
			if importing == key {
				st.mu.Unlock()
				cycle := strings.Join(append(keys, key), " -> ")
				return nil, object.Errorf("import error: circular import detected: %s", cycle)
			}
		}

		// Add to import stack
		st.stack = append(st.stack, core.ImportFrame{Key: key, Base: core.ImportBaseFor(key)})
		st.mu.Unlock()

		defer func() {
			st.mu.Lock()
			st.stack = st.stack[:len(st.stack)-1]
			st.mu.Unlock()
		}()

		// Check if HTTP(S) import is allowed. A remote script importing
		// from its own origin is implicit (same-origin policy already
		// applied); loading a URL from a local script needs httpimport.
		if core.IsHTTPURL(key) && !baseIsRemote {
			if !w.moduleEnabled("httpimport") {
				return nil, object.Errorf("import error: HTTP(S) imports are disabled. Use gui.WithHTTPImport() to enable importing from URLs")
			}
		}

		// Load script source
		source, err := core.LoadImportSource(ctx, key)
		if err != nil {
			return nil, object.Errorf("import error: %w", err)
		}

		// Compile the module with exactly the globals the module VM will
		// get (issue #27): names from opaque WithRisorOptions are not in
		// the module VM, so compiling with them would let modules pass
		// compilation and then fail at runtime. They fail now, clearly.
		code, err := risor.Compile(ctx, source, risor.WithEnv(st.env))
		if err != nil {
			return nil, object.Errorf("import %q: compilation failed: %w", modulePath, err)
		}

		// Create the module VM with the run's environment (widget, http,
		// sql, ...), so module code and functions can use those globals.
		moduleVM, err := vm.New(code, vm.WithGlobals(st.env))
		if err != nil {
			return nil, object.Errorf("import %q: VM creation failed: %w", modulePath, err)
		}

		err = moduleVM.Run(ctx)
		if err != nil {
			return nil, object.Errorf("import %q: execution failed: %w", modulePath, err)
		}

		// Extract exported globals. Exported functions are bound to the module's
		// own VM so that references to other module-level variables and functions
		// resolve correctly (see core.WrapModuleExports for details).
		globals := core.WrapModuleExports(moduleVM)

		// Create module object
		module := core.NewImportedModule(key, globals)

		// Cache it
		st.mu.Lock()
		st.cache[key] = module
		st.mu.Unlock()

		return module, nil
	})
}
