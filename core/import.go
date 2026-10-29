package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/vm"
)

// ImportFrame records one in-progress import. Key is the canonical cache
// key of the module being imported; Base is the base its own relative
// imports must resolve against (directory for local modules, the module URL
// for remote ones). See issue #8.
type ImportFrame struct {
	Key  string
	Base string
}

// ImportBaseFor maps a script's source (file path or URL) to the base used
// to resolve its relative imports: the containing directory for local
// files, or the script URL itself for http(s) sources. An empty source
// means "unknown", which keeps the legacy process-CWD behaviour.
func ImportBaseFor(source string) string {
	if source == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(source, "http://"), strings.HasPrefix(source, "https://"):
		return source
	case strings.HasPrefix(source, "file://"):
		if u, err := url.Parse(source); err == nil {
			return filepath.Dir(u.Path)
		}
		return ""
	default:
		return filepath.Dir(source)
	}
}

// isHTTPURL reports whether s is an http(s) URL.
func isHTTPURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// importBaseForKey returns the base for resolving the imports *made by* the
// module identified by a canonical key.
func importBaseForKey(key string) string {
	if isHTTPURL(key) {
		return key
	}
	return filepath.Dir(key)
}

// ResolveImportPath resolves an import() argument against the importing
// script's base: a directory for local scripts, the importing script's URL
// for remote (http/https) ones, or "" for the legacy process-CWD behaviour
// (issue #8). The returned string is the canonical cache key: an absolute
// file path or an absolute URL, so "./a.risor" and "a.risor" identify the
// same module.
//
// Policy: a script loaded over HTTP(S) may only import from its own origin.
// Its relative paths resolve against the remote URL and are never consulted
// on the local disk; absolute or cross-origin URLs are rejected.
func ResolveImportPath(modulePath, base string) (string, error) {
	argIsURL := isHTTPURL(modulePath)
	baseIsRemote := isHTTPURL(base)

	if baseIsRemote {
		baseURL, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("invalid script base URL %q: %w", base, err)
		}
		var target *url.URL
		if argIsURL {
			target, err = url.Parse(modulePath)
			if err != nil {
				return "", fmt.Errorf("invalid import URL %q: %w", modulePath, err)
			}
		} else {
			// Relative (or root-relative) reference resolved against the
			// importing script's URL directory. A local absolute path
			// parses as a plain path reference and stays on the remote
			// origin, so remote scripts cannot reach the local disk.
			ref, err := url.Parse(strings.ReplaceAll(modulePath, "\\", "/"))
			if err != nil {
				return "", fmt.Errorf("invalid import path %q: %w", modulePath, err)
			}
			dir := *baseURL
			dir.Path = path.Dir(baseURL.Path) + "/"
			target = dir.ResolveReference(ref)
		}
		if target.Scheme != baseURL.Scheme || target.Host != baseURL.Host {
			return "", fmt.Errorf("cross-origin import blocked: %q is not on the script origin %s://%s",
				modulePath, baseURL.Scheme, baseURL.Host)
		}
		return target.String(), nil
	}

	// Local script (or unknown base): URL imports are allowed as-is; the
	// caller enforces the httpimport gate.
	if argIsURL {
		return modulePath, nil
	}
	if base == "" {
		abs, err := filepath.Abs(modulePath)
		if err != nil {
			return "", fmt.Errorf("resolve %q: %w", modulePath, err)
		}
		return filepath.Clean(abs), nil
	}
	if filepath.IsAbs(modulePath) || filepath.VolumeName(modulePath) != "" {
		return filepath.Clean(modulePath), nil
	}
	return filepath.Clean(filepath.Join(base, modulePath)), nil
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

		// Resolve against the importing script's base (the entry script's
		// base, or the base of the module at the top of the import stack)
		// and enforce the remote same-origin policy (issue #8).
		cb.moduleMutex.Lock()
		base := cb.scriptBase
		if n := len(cb.importStack); n > 0 {
			base = cb.importStack[n-1].Base
		}
		cb.moduleMutex.Unlock()
		baseIsRemote := isHTTPURL(base)

		key, err := ResolveImportPath(modulePath, base)
		if err != nil {
			return nil, object.Errorf("import error: %w", err)
		}

		// Thread-safe cache check (keyed by the canonical path/URL)
		cb.moduleMutex.Lock()
		if module, found := cb.moduleCache[key]; found {
			cb.moduleMutex.Unlock()
			return module, nil
		}

		// Check for circular imports
		keys := make([]string, len(cb.importStack))
		for i, f := range cb.importStack {
			keys[i] = f.Key
		}
		for _, importing := range keys {
			if importing == key {
				cb.moduleMutex.Unlock()
				cycle := strings.Join(append(keys, key), " -> ")
				return nil, object.Errorf("import error: circular import detected: %s", cycle)
			}
		}

		// Add to import stack
		cb.importStack = append(cb.importStack, ImportFrame{Key: key, Base: importBaseForKey(key)})
		cb.moduleMutex.Unlock()

		defer func() {
			cb.moduleMutex.Lock()
			cb.importStack = cb.importStack[:len(cb.importStack)-1]
			cb.moduleMutex.Unlock()
		}()

		// Check if HTTP(S) import is allowed. A remote script importing
		// from its own origin is implicit (same-origin policy already
		// applied); loading a URL from a local script needs httpimport.
		if isHTTPURL(key) && !baseIsRemote {
			if !cb.enabledModules["httpimport"] {
				return nil, object.Errorf("import error: HTTP(S) imports are disabled. Use core.WithHTTPImport() to enable importing from URLs")
			}
		}

		// Load script source
		source, err := loadImportSource(ctx, key)
		if err != nil {
			return nil, object.Errorf("import error: %w", err)
		}

		// Compile the module with exactly the globals the module VM will
		// get (issue #27): names from opaque WithRisorOptions are not in
		// the module VM, so compiling with them would let modules pass
		// compilation and then fail at runtime. They fail now, clearly.
		code, err := risor.Compile(ctx, source, risor.WithEnv(cb.env))
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
		module := NewImportedModule(key, globals)

		// Cache it
		cb.moduleMutex.Lock()
		cb.moduleCache[key] = module
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
func loadImportSource(ctx context.Context, path string) (string, error) {
	// Check if it's an HTTP(S) URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return FetchScript(ctx, path)
	}

	// Load from file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %q: %w", path, err)
	}
	return string(data), nil
}

const (
	// scriptFetchTimeout bounds script downloads started without a deadline.
	scriptFetchTimeout = 30 * time.Second
	// scriptFetchLimit caps the size of a downloaded script (issue #19).
	scriptFetchLimit = 10 << 20 // 10 MiB
)

// FetchScript downloads a script from an HTTP(S) URL. It is the single
// loader shared by core and gui (issue #19): non-200 responses are errors,
// the request honours ctx plus a fixed timeout, and the body size is
// limited.
func FetchScript(ctx context.Context, url string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, scriptFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %q: %w", url, err)
	}

	client := &http.Client{Timeout: scriptFetchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %q: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch %q: HTTP %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, scriptFetchLimit+1))
	if err != nil {
		return "", fmt.Errorf("failed to read response from %q: %w", url, err)
	}
	if int64(len(body)) > scriptFetchLimit {
		return "", fmt.Errorf("failed to fetch %q: response exceeds %d bytes", url, scriptFetchLimit)
	}

	return string(body), nil
}
