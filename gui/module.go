package gui

import (
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

// ImportedModule wraps the exported globals from an imported Risor script.
// It provides GetAttr access to all top-level variables and functions
// defined in the imported module.
//
// Example:
//
//	// math_utils.risor
//	let add = (a, b) => a + b
//	let PI = 3.14159
//
//	// main.risor
//	let math = import("math_utils.risor")
//	print(math.add(5, 3))  // 8
//	print(math.PI)         // 3.14159
type ImportedModule struct {
	name    string
	globals map[string]object.Object
}

// NewImportedModule creates a new module object with the given exports.
func NewImportedModule(name string, globals map[string]object.Object) *ImportedModule {
	return &ImportedModule{
		name:    name,
		globals: globals,
	}
}

// Type returns the type identifier for this object.
func (m *ImportedModule) Type() object.Type {
	return "module"
}

// Inspect returns a string representation of the module.
func (m *ImportedModule) Inspect() string {
	return "<module: " + m.name + ">"
}

// GetAttr provides dot notation access to module exports.
// Returns the exported value and true if found, or nil and false if not found.
func (m *ImportedModule) GetAttr(name string) (object.Object, bool) {
	if obj, found := m.globals[name]; found {
		return obj, true
	}
	return nil, false
}

// SetAttr is not supported for modules (they are read-only after import).
func (m *ImportedModule) SetAttr(name string, value object.Object) error {
	return object.Errorf("type error: module is read-only")
}

// IsTruthy returns true (modules are always truthy).
func (m *ImportedModule) IsTruthy() bool {
	return true
}

// RunOperation is not supported for modules.
func (m *ImportedModule) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("type error: unsupported operation for module")
}

// Equals compares modules by name.
func (m *ImportedModule) Equals(other object.Object) bool {
	if om, ok := other.(*ImportedModule); ok {
		return m.name == om.name
	}
	return false
}

// Interface returns the module as a Go value (returns the module itself).
func (m *ImportedModule) Interface() interface{} {
	return m
}

// Attrs returns metadata about the module's exported attributes.
func (m *ImportedModule) Attrs() []object.AttrSpec {
	attrs := make([]object.AttrSpec, 0, len(m.globals))
	for name := range m.globals {
		attrs = append(attrs, object.AttrSpec{
			Name:    name,
			Doc:     "Exported from module",
			Returns: "any",
		})
	}
	return attrs
}
