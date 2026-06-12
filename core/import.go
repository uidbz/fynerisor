package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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

		// Create VM and execute module
		moduleVM, err := vm.New(code)
		if err != nil {
			return nil, object.Errorf("import %q: VM creation failed: %w", modulePath, err)
		}

		err = moduleVM.Run(ctx)
		if err != nil {
			return nil, object.Errorf("import %q: execution failed: %w", modulePath, err)
		}

		// Extract all globals from VM
		globals := make(map[string]object.Object)
		for _, name := range moduleVM.GlobalNames() {
			value, err := moduleVM.Get(name)
			if err == nil && value != nil {
				globals[name] = value
			}
		}

		// Create module object
		module := NewImportedModule(modulePath, globals)

		// Cache it
		cb.moduleMutex.Lock()
		cb.moduleCache[modulePath] = module
		cb.moduleMutex.Unlock()

		return module, nil
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
