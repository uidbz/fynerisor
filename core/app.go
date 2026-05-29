package core

import (
	"context"
	"fmt"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

// appObject represents the app global object in Risor scripts
type appObject struct {
	ctx *Context
}

func newAppObjectForContext(ctx *Context) *appObject {
	return &appObject{ctx: ctx}
}

func (a *appObject) Type() object.Type {
	return "app"
}

func (a *appObject) Inspect() string {
	return "app"
}

func (a *appObject) Interface() interface{} {
	return a
}

func (a *appObject) IsTruthy() bool {
	return true
}

func (a *appObject) Equals(other object.Object) bool {
	_, ok := other.(*appObject)
	return ok
}

func (a *appObject) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for app: %v", opType)
}

func (a *appObject) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "name":
		return object.NewString(a.ctx.appName), true
	case "version":
		return object.NewString(GetAppVersion()), true
	}
	return nil, false
}

func (a *appObject) SetAttr(name string, value object.Object) error {
	return object.Errorf("type error: app object attributes are read-only")
}

func (a *appObject) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "name", Doc: "Name of the host application"},
		{Name: "version", Doc: "Version of the host application"},
	}
}

// printBuiltin returns a print builtin function for core (no GUI dependency)
func newPrintBuiltin() object.Object {
	return object.NewBuiltin("print", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		// Use standard fmt.Println for headless mode
		var values []interface{}
		for _, arg := range args {
			values = append(values, arg.Interface())
		}
		fmt.Println(values...)
		return object.Nil, nil
	})
}

func (a *appObject) Cost() int {
	return 0
}
