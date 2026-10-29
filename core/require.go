package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// RequireEnv binds the require() engine to one runtime (issue #32:
// a single implementation for both Context and Window).
type RequireEnv struct {
	// PkgName prefixes the fix-it hints ("core" or "gui").
	PkgName string
	// GUIAvailable reports whether @gui may be required (Window: yes,
	// headless Context: no).
	GUIAvailable bool
	// AppVersion returns the runtime's app version at call time ("" =
	// check against fynerisor's own Version).
	AppVersion func() string
	// ModuleEnabled reports whether @name is currently enabled; it is
	// consulted live, so globals registered after construction count.
	ModuleEnabled func(name string) bool
}

// NewRequireBuiltin builds the require() builtin for one runtime.
func NewRequireBuiltin(env RequireEnv) *object.Builtin {
	return object.NewBuiltin("require", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("require: expected 1 argument (string or list), got %d", len(args))
		}

		// Handle string argument
		if str, ok := args[0].(*object.String); ok {
			return processRequirement(env, str.Value())
		}

		// Handle list argument
		if list, ok := args[0].(*object.List); ok {
			for _, item := range list.Value() {
				str, ok := item.(*object.String)
				if !ok {
					return nil, fmt.Errorf("require: list items must be strings, got %s", item.Type())
				}
				if _, err := processRequirement(env, str.Value()); err != nil {
					return nil, err
				}
			}
			return object.Nil, nil
		}

		return nil, fmt.Errorf("require: argument must be a string or list, got %s", args[0].Type())
	})
}

// newRequireBuiltinForContext creates the require() function for Context
func newRequireBuiltinForContext(cb *Context) *object.Builtin {
	return NewRequireBuiltin(RequireEnv{
		PkgName:       "core",
		GUIAvailable:  false,
		AppVersion:    func() string { return cb.appVersion },
		ModuleEnabled: func(name string) bool { return cb.enabledModules[name] },
	})
}

// ModuleOptionName maps a built-in module name to the WithX option that
// enables it, for require() error hints. It returns "" when the name is
// not a built-in module (issue #20).
func ModuleOptionName(moduleName string) string {
	switch moduleName {
	case "csv":
		return "WithCSV"
	case "exec":
		return "WithExec"
	case "filepath":
		return "WithFilepath"
	case "http":
		return "WithHTTP"
	case "httpimport":
		return "WithHTTPImport"
	case "io":
		return "WithIO"
	case "json":
		return "WithJSON"
	case "os":
		return "WithOS"
	case "sql":
		return "WithSQL"
	case "strings":
		return "WithStrings"
	case "time":
		return "WithTime"
	case "tie":
		return "WithTie"
	}
	return ""
}

// processRequirement handles a single requirement string against one
// runtime's RequireEnv.
func processRequirement(env RequireEnv, req string) (object.Object, error) {
	// Module requirement (@module)
	if strings.HasPrefix(req, "@") {
		moduleName := strings.TrimPrefix(req, "@")

		if moduleName == "" {
			return nil, fmt.Errorf("require: empty module name in %q (use '@sql' style for modules)", req)
		}

		// Special case: @gui is only available where a window exists.
		if moduleName == "gui" {
			if env.GUIAvailable {
				return object.Nil, nil
			}
			return nil, fmt.Errorf("require: @gui is not available in headless mode (script requires GUI window)")
		}

		if !env.ModuleEnabled(moduleName) {
			if moduleName == "sql" {
				return nil, fmt.Errorf("require: module @sql is not enabled (import github.com/uidbz/fynerisor/%s/db and pass db.WithSQL())", env.PkgName)
			}
			if optionName := ModuleOptionName(moduleName); optionName != "" {
				return nil, fmt.Errorf("require: module @%s is not enabled (use %s.%s() option)", moduleName, env.PkgName, optionName)
			}
			return nil, fmt.Errorf("require: module @%s is not enabled and is not a built-in module (host-provided modules are registered via %s.WithGlobal())", moduleName, env.PkgName)
		}
		return object.Nil, nil
	}

	// Version requirement (v0.2.0 or ==v0.2.0)
	if strings.HasPrefix(req, "v") || strings.HasPrefix(req, "==v") {
		// Use app version if set, otherwise use fynerisor version
		versionToCheck := "v" + Version
		versionName := "fynerisor"
		if appVersion := env.AppVersion(); appVersion != "" {
			versionToCheck = appVersion
			if !strings.HasPrefix(versionToCheck, "v") {
				versionToCheck = "v" + versionToCheck
			}
			versionName = "application"
		}

		if err := checkVersion(req, versionToCheck, versionName); err != nil {
			return nil, err
		}
		return object.Nil, nil
	}

	return nil, fmt.Errorf("require: invalid requirement '%s' (use 'v0.2.0' for version or '@sql' for module)", req)
}

// checkVersion compares required version against current version
func checkVersion(required, current, versionName string) error {
	// Check for exact version match (==)
	exactMatch := false
	if strings.HasPrefix(required, "==") {
		exactMatch = true
		required = strings.TrimPrefix(required, "==")
	}

	// Strip 'v' prefix
	required = strings.TrimPrefix(required, "v")
	current = strings.TrimPrefix(current, "v")

	reqParts, err := parseVersion(required)
	if err != nil {
		return fmt.Errorf("require: invalid version format 'v%s': %w", required, err)
	}

	curParts, err := parseVersion(current)
	if err != nil {
		return fmt.Errorf("require: invalid current version 'v%s': %w", current, err)
	}

	// For exact match, all three components must match
	if exactMatch {
		if reqParts != curParts {
			return fmt.Errorf("%s version ==v%s required (exact match), but running v%s", versionName, required, current)
		}
		return nil
	}

	// For minimum version (>=), compare major.minor.patch
	for i := 0; i < 3; i++ {
		if curParts[i] < reqParts[i] {
			return fmt.Errorf("%s version v%s or higher required, but running v%s", versionName, required, current)
		}
		if curParts[i] > reqParts[i] {
			// Current version is higher, requirement satisfied
			return nil
		}
		// Equal, check next part
	}

	// All parts equal, requirement satisfied
	return nil
}

// parseVersion parses a semantic version string into [major, minor, patch]
func parseVersion(version string) ([3]int, error) {
	var parts [3]int

	// Split by dots
	segments := strings.Split(version, ".")
	if len(segments) == 0 || len(segments) > 3 {
		return parts, fmt.Errorf("invalid version format")
	}

	// Parse each segment
	for i := 0; i < len(segments) && i < 3; i++ {
		val, err := strconv.Atoi(segments[i])
		if err != nil {
			return parts, fmt.Errorf("invalid version number: %s", segments[i])
		}
		if val < 0 {
			return parts, fmt.Errorf("version numbers cannot be negative")
		}
		parts[i] = val
	}

	// If version is "0.2", treat as "0.2.0"
	// parts already initialized to [0, 0, 0]

	return parts, nil
}
