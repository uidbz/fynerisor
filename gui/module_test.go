package gui

import (
	"context"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

// mockModule is a simple test object for WithGlobal
type mockModule struct{}

const mockModuleType object.Type = "mockmodule"

func (m *mockModule) Type() object.Type {
	return mockModuleType
}

func (m *mockModule) Inspect() string {
	return "mockmodule"
}

func (m *mockModule) GetAttr(name string) (object.Object, bool) {
	if name == "Test" {
		return object.NewBuiltin("Test", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			return object.NewString("test called"), nil
		}), true
	}
	return nil, false
}

func (m *mockModule) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "Test", Doc: "Test method"},
	}
}

func (m *mockModule) IsTruthy() bool {
	return true
}

func (m *mockModule) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for mockmodule: %v", opType), nil
}

func (m *mockModule) SetAttr(name string, value object.Object) error {
	return object.Errorf("attribute error: mockmodule object does not support attribute assignment")
}

func (m *mockModule) Interface() any {
	return m
}

func (m *mockModule) Equals(other object.Object) bool {
	_, ok := other.(*mockModule)
	return ok
}

// TestWithGlobal tests that WithGlobal registers a custom module
func TestWithGlobal(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")

	mock := &mockModule{}
	fw := NewWindow(w, WithGlobal("custom", mock))

	// Test that require() succeeds when module is registered
	script := `
		require(["@custom"])
		let result = custom.Test()
		window.SetContent(widget.NewLabel(result))
	`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(100 * time.Millisecond)

	if fw.Status != "Ready!" {
		t.Errorf("Expected Ready!, got: %s", fw.Status)
	}
}

// TestWithGlobalRequireFails tests that require() fails when module is not registered
func TestWithGlobalRequireFails(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")

	fw := NewWindow(w) // No WithGlobal

	// Test that require() fails when module is not available
	script := `
		require(["@custom"])
		window.SetContent(widget.NewLabel("Should not reach here"))
	`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(100 * time.Millisecond)

	// Should have error status
	if !strings.Contains(fw.Status, "ERROR") {
		t.Errorf("Expected ERROR status, got: %s", fw.Status)
	}

	// Error should mention the module name and suggest the option
	if !strings.Contains(fw.Status, "@custom") {
		t.Errorf("Error should mention @custom, got: %s", fw.Status)
	}
}

// TestWithGlobalMultiple tests that multiple custom globals can be registered
func TestWithGlobalMultiple(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")

	mock1 := &mockModule{}
	mock2 := &mockModule{}
	fw := NewWindow(w,
		WithGlobal("module1", mock1),
		WithGlobal("module2", mock2),
	)

	// Test that both modules can be required
	script := `
		require(["@module1", "@module2"])
		let r1 = module1.Test()
		let r2 = module2.Test()
		window.SetContent(widget.NewLabel(r1 + " " + r2))
	`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(100 * time.Millisecond)

	if fw.Status != "Ready!" {
		t.Errorf("Expected Ready! with multiple modules, got: %s", fw.Status)
	}
}
