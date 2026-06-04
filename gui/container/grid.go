package container

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &GridWithColumns{}
var _ object.Object = &GridWithRows{}

const GridWithColumnsType object.Type = "container.gridwithcolumns"
const GridWithRowsType object.Type = "container.gridwithrows"

type GridWithColumns struct {
	instance fyne.CanvasObject
}

func (obj *GridWithColumns) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *GridWithColumns) Type() object.Type {
	return GridWithColumnsType
}

func (obj *GridWithColumns) Inspect() string {
	return "container.gridwithcolumns"
}

func (obj *GridWithColumns) Interface() interface{} {
	return obj.instance
}

func (obj *GridWithColumns) IsTruthy() bool {
	return true
}

func (obj *GridWithColumns) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.gridwithcolumns'")
}

func (obj *GridWithColumns) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(GridWithColumnsType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", GridWithColumnsType, opType)
	return errObj, err
}

func (obj *GridWithColumns) Equals(other object.Object) bool {
	return obj == other
}

func (obj *GridWithColumns) Attrs() []object.AttrSpec {
	return nil
}

func (obj *GridWithColumns) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", GridWithColumnsType, name)
}

func (obj *GridWithColumns) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("container.GridWithColumns.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.GridWithColumns.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewGridWithColumns(cols int, objects ...fyne.CanvasObject) *GridWithColumns {
	return &GridWithColumns{
		instance: container.NewGridWithColumns(cols, objects...),
	}
}

type GridWithRows struct {
	instance fyne.CanvasObject
}

func (obj *GridWithRows) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *GridWithRows) Type() object.Type {
	return GridWithRowsType
}

func (obj *GridWithRows) Inspect() string {
	return "container.gridwithrows"
}

func (obj *GridWithRows) Interface() interface{} {
	return obj.instance
}

func (obj *GridWithRows) IsTruthy() bool {
	return true
}

func (obj *GridWithRows) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.gridwithrows'")
}

func (obj *GridWithRows) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(GridWithRowsType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", GridWithRowsType, opType)
	return errObj, err
}

func (obj *GridWithRows) Equals(other object.Object) bool {
	return obj == other
}

func (obj *GridWithRows) Attrs() []object.AttrSpec {
	return nil
}

func (obj *GridWithRows) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", GridWithRowsType, name)
}

func (obj *GridWithRows) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("container.GridWithRows.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.GridWithRows.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewGridWithRows(rows int, objects ...fyne.CanvasObject) *GridWithRows {
	return &GridWithRows{
		instance: container.NewGridWithRows(rows, objects...),
	}
}
