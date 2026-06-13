package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/vm"
)

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
// - Global inheritance: Modules have access to http, sql, os, etc.
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
func (cb *Context) newImportBuiltin() *object.Builtin {
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

		// Thread-safe cache check
		cb.moduleMutex.Lock()
		if module, found := cb.moduleCache[modulePath]; found {
			cb.moduleMutex.Unlock()
			return module, nil
		}

		// Check for circular imports
		for _, importing := range cb.importStack {
			if importing == modulePath {
				cb.moduleMutex.Unlock()
				cycle := strings.Join(append(cb.importStack, modulePath), " -> ")
				return nil, object.Errorf("import error: circular import detected: %s", cycle)
			}
		}

		// Add to import stack
		cb.importStack = append(cb.importStack, modulePath)
		cb.moduleMutex.Unlock()

		defer func() {
			cb.moduleMutex.Lock()
			cb.importStack = cb.importStack[:len(cb.importStack)-1]
			cb.moduleMutex.Unlock()
		}()

		// Check if HTTP(S) import is allowed
		if strings.HasPrefix(modulePath, "http://") || strings.HasPrefix(modulePath, "https://") {
			if !cb.enabledModules["httpimport"] {
				return nil, object.Errorf("import error: HTTP(S) imports are disabled. Use core.WithHTTPImport() to enable importing from URLs")
			}
		}

		// Load script source
		source, err := loadImportSource(modulePath)
		if err != nil {
			return nil, object.Errorf("import error: %w", err)
		}

		// Compile the module
		code, err := risor.Compile(ctx, source, cb.globals...)
		if err != nil {
			return nil, object.Errorf("import %q: compilation failed: %w", modulePath, err)
		}

		// Create VM and execute module. The module VM is seeded with the same
		// merged environment as the main script so module code and functions can
		// access globals like http, sql, os, etc.
		moduleVM, err := vm.New(code, vm.WithGlobals(cb.env))
		if err != nil {
			return nil, object.Errorf("import %q: VM creation failed: %w", modulePath, err)
		}

		err = moduleVM.Run(ctx)
		if err != nil {
			return nil, object.Errorf("import %q: execution failed: %w", modulePath, err)
		}

		// Extract exported globals. Exported functions are bound to the module's
		// own VM so that references to other module-level variables and functions
		// resolve correctly (see wrapModuleExports for details).
		globals := wrapModuleExports(moduleVM)

		// Create module object
		module := NewImportedModule(modulePath, globals)

		// Cache it
		cb.moduleMutex.Lock()
		cb.moduleCache[modulePath] = module
		cb.moduleMutex.Unlock()

		return module, nil
	})
}

// wrapModuleExports converts a module VM's global variables into the export map
// returned to the importing script.
//
// Risor v2 resolves global variable references by index against the *currently
// executing* VM's globals array; globals are not captured into closures. If an
// exported function were called directly by the importing script's VM, its
// references to module-level variables/functions would resolve against the
// wrong globals array (producing wrong values or an index-out-of-range error).
//
// To make module-level references work, every exported function is wrapped in a
// builtin that re-enters the module's own VM via moduleVM.Call. Non-function
// values (numbers, strings, lists, maps, ...) are exported directly.
func wrapModuleExports(moduleVM *vm.VirtualMachine) map[string]object.Object {
	// One mutex per module serializes calls into its (single, stateful) VM.
	mu := &sync.Mutex{}
	exports := make(map[string]object.Object)
	for _, name := range moduleVM.GlobalNames() {
		value, err := moduleVM.Get(name)
		if err != nil || value == nil {
			continue
		}
		if closure, ok := value.(*object.Closure); ok {
			exports[name] = bindModuleFunction(moduleVM, mu, name, closure)
		} else {
			exports[name] = value
		}
	}
	return exports
}

// bindModuleFunction wraps an exported module closure so that invoking it runs
// the closure inside the module's own VM (where its global references resolve
// correctly).
func bindModuleFunction(moduleVM *vm.VirtualMachine, mu *sync.Mutex, name string, closure *object.Closure) *object.Builtin {
	return object.NewBuiltin(name, func(ctx context.Context, args ...object.Object) (object.Object, error) {
		mu.Lock()
		defer mu.Unlock()
		return moduleVM.Call(ctx, closure, args)
	})
}

// loadImportSource loads source code from a file path or URL.
func loadImportSource(path string) (string, error) {
	// Check if it's an HTTP(S) URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return fetchHTTPScript(path)
	}

	// Load from file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %q: %w", path, err)
	}
	return string(data), nil
}

// fetchHTTPScript downloads a script from an HTTP(S) URL.
func fetchHTTPScript(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %q: HTTP %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from %q: %w", url, err)
	}

	return string(body), nil
}
