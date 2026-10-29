package widget

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Log{}

const LogType object.Type = "widget.Log"

type Log struct {
	maxItems int
	mu       sync.Mutex // guards data (written from any goroutine, read by the list callbacks)
	data     []string
	writeMu  sync.Mutex // guards pending (io.Write buffering)
	pending  []byte
	instance *widget.List
}

func (obj *Log) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Log) Type() object.Type {
	return LogType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Log) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Log) inspectUI() string {
	return "widget.Log"
}

func (obj *Log) Interface() interface{} {
	return obj.instance
}

func (obj *Log) IsTruthy() bool {
	return true
}

func (obj *Log) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", LogType, opType)
}

func (obj *Log) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Log) Attrs() []object.AttrSpec {
	return logAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Log) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Log) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", LogType, name)
}

var logAttrs = object.NewAttrRegistry[*Log](string(LogType))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Log) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Log) getAttrUI(name string) (object.Object, bool) {
	return logAttrs.GetAttr(obj, name)

}

// Lines returns a snapshot of the buffered lines, safe from any
// goroutine (issue #24: hosts and tests inspect what the runtime
// printed).
func (obj *Log) Lines() []string {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	return append([]string(nil), obj.data...)
}

// AppendGo appends a line from Go code without entering the Risor VM.
// It is safe from any goroutine (issue #24: it is the sink of the
// runtime's stdout writer, which the print builtin feeds from the VM
// goroutine); the list refresh is marshalled to Fyne's UI thread.
func (obj *Log) AppendGo(line string) {
	obj.appendLine(line)
}

// appendLine is the single mutation path for data; every reader of
// obj.data (the list callbacks, Clear, the script Append builtin) takes
// the same lock, so a Log stays consistent no matter which goroutine
// writes it.
func (obj *Log) appendLine(line string) {
	obj.mu.Lock()
	if len(obj.data) > obj.maxItems-1 {
		obj.data = obj.data[1:]
	}
	obj.data = append(obj.data, line)
	obj.mu.Unlock()
	guithread.Do(func() {
		obj.instance.Refresh()
		obj.instance.ScrollToBottom()
	})
}

// Write implements io.Writer (issue #24): a Log can be used anywhere a
// text sink is expected — window.SetStdout(log), io.MultiWriter, the
// exec module. Complete lines are appended via AppendGo; a trailing
// partial line stays buffered until the newline arrives.
func (obj *Log) Write(p []byte) (int, error) {
	obj.writeMu.Lock()
	obj.pending = append(obj.pending, p...)
	var lines []string
	for {
		i := bytes.IndexByte(obj.pending, '\n')
		if i < 0 {
			break
		}
		lines = append(lines, string(obj.pending[:i]))
		obj.pending = obj.pending[i+1:]
	}
	obj.writeMu.Unlock()
	for _, line := range lines {
		obj.AppendGo(line)
	}
	return len(p), nil
}

func NewLog(maxItems int) *Log {
	log := &Log{maxItems: maxItems}
	log.instance = widget.NewList(
		func() int {
			log.mu.Lock()
			defer log.mu.Unlock()
			return len(log.data)
		},
		func() fyne.CanvasObject {
			l := widget.NewLabel("template")
			l.Selectable = true
			return l
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			log.mu.Lock()
			defer log.mu.Unlock()
			if i >= len(log.data) {
				return // trimmed concurrently; keep the label as-is
			}
			o.(*widget.Label).SetText(log.data[i])
		})
	log.instance.HideSeparators = true

	return log
}
func init() {
	objattrs.Add(logAttrs)
	logAttrs.Define(
		"Append").
		Args("arg1").Returns("nil").Impl(func(obj *Log, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		line, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		obj.appendLine(line)
		return object.Nil, nil
	})
	logAttrs.Define(
		"Clear").
		Doc("Clear the contents.").
		Args().Returns("nil").Impl(func(obj *Log, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		obj.mu.Lock()
		obj.data = []string{}
		obj.mu.Unlock()
		guithread.Do(func() {
			obj.
				instance.
				Refresh()
			obj.instance.
				ScrollToBottom()
		})
		return object.Nil,

			nil
	})
}
