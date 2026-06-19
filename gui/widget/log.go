package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Log{}

const LogType object.Type = "widget.Log"

type Log struct {
	maxItems int
	data     []string
	instance *widget.List
}

func (obj *Log) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Log) Type() object.Type {
	return LogType
}

func (obj *Log) Inspect() string {
	return "widget.Log"
}

func (obj *Log) Interface() interface{} {
	return obj.instance
}

func (obj *Log) IsTruthy() bool {
	return true
}

func (obj *Log) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Log'")
}

func (obj *Log) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(LogType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", LogType, opType)
	return errObj, err
}

func (obj *Log) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Log) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Log) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", LogType, name)
}

func (obj *Log) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Append":
		return object.NewBuiltin("widget.Log.Append", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}

			line, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			if len(obj.data) > obj.maxItems-1 {
				obj.data = obj.data[1:]
			}
			obj.data = append(obj.data, line)
			obj.instance.Refresh()
			obj.instance.ScrollToBottom()

			return object.Nil, nil
		}), true

	case "Clear":
		return object.NewBuiltin("widget.Log.Clear", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 0", len(args)), nil
			}

			obj.data = []string{}
			guithread.Do(func() {
				obj.instance.Refresh()
				obj.instance.ScrollToBottom()
			})

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Log.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Log.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewLog(maxItems int) *Log {
	log := &Log{maxItems: maxItems}
	log.instance = widget.NewList(
		func() int {
			return len(log.data)
		},
		func() fyne.CanvasObject {
			l := widget.NewLabel("template")
			l.Selectable = true
			return l
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(log.data[i])
		})
	log.instance.HideSeparators = true

	return log
}
