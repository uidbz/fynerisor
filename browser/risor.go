package browser

import (
	"context"
	"errors"
	"fmt"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

// RisorBrowser wraps a Browser to expose it as a Risor object
// Applications can register this with gui.WithGlobal("browser", browserObj)
type RisorBrowser struct {
	browser *Browser
}

// NewRisorBrowser creates a Risor-accessible browser object
func NewRisorBrowser(b *Browser) *RisorBrowser {
	return &RisorBrowser{browser: b}
}

// Type returns the Risor type name
func (rb *RisorBrowser) Type() object.Type {
	return "browser"
}

// Inspect returns a string representation
func (rb *RisorBrowser) Inspect() string {
	return "browser"
}

// String returns a string representation
func (rb *RisorBrowser) String() string {
	return "browser"
}

// Interface returns the underlying Go value
func (rb *RisorBrowser) Interface() interface{} {
	return rb.browser
}

// IsTruthy returns whether the object is truthy
func (rb *RisorBrowser) IsTruthy() bool {
	return true
}

// Cost returns the memory cost
func (rb *RisorBrowser) Cost() int {
	return 0
}

// MarshalJSON implements JSON marshaling
func (rb *RisorBrowser) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'browser'")
}

// RunOperation implements binary operations
func (rb *RisorBrowser) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for browser: %v", opType),
		errors.New("eval error: unsupported operation for browser")
}

// Equals checks equality
func (rb *RisorBrowser) Equals(other object.Object) bool {
	otherBrowser, ok := other.(*RisorBrowser)
	if !ok {
		return false
	}
	return rb.browser == otherBrowser.browser
}

// Attrs returns available attributes
func (rb *RisorBrowser) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "Open", Doc: "Navigate to a URL programmatically: browser.Open(\"https://example.com\")"},
		{Name: "GetURL", Doc: "Get current URL: browser.GetURL()"},
		{Name: "SetStatus", Doc: "Set status bar text: browser.SetStatus(\"Loading...\")"},
	}
}

// SetAttr sets an attribute (not supported)
func (rb *RisorBrowser) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: browser object has no attribute %q", name)
}

// GetAttr returns an attribute
func (rb *RisorBrowser) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Open":
		return object.NewBuiltin("browser.Open", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			url, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			rb.browser.NavigateToURL(url)
			return object.Nil, nil
		}), true

	case "GetURL":
		return object.NewBuiltin("browser.GetURL", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return object.NewString(rb.browser.GetURL()), nil
		}), true

	case "SetStatus":
		return object.NewBuiltin("browser.SetStatus", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			status, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			rb.browser.SetStatus(status)
			return object.Nil, nil
		}), true
	}

	return nil, false
}
