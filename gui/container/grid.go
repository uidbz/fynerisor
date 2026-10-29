package container

import (
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *GridWithColumns) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *GridWithColumns) inspectUI() string {
	return "container.gridwithcolumns"
}

func (obj *GridWithColumns) Interface() interface{} {
	return obj.instance
}

func (obj *GridWithColumns) IsTruthy() bool {
	return true
}

func (obj *GridWithColumns) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", GridWithColumnsType, opType)
}

func (obj *GridWithColumns) Equals(other object.Object) bool {
	return obj == other
}

func (obj *GridWithColumns) Attrs() []object.AttrSpec {
	return gridWithColumnsAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *GridWithColumns) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *GridWithColumns) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", GridWithColumnsType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *GridWithColumns) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var gridWithColumnsAttrs = object.NewAttrRegistry[*GridWithColumns](string(GridWithColumnsType))

func (obj *GridWithColumns) getAttrUI(name string) (object.Object, bool) {
	return gridWithColumnsAttrs.GetAttr(obj, name)

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *GridWithRows) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *GridWithRows) inspectUI() string {
	return "container.gridwithrows"
}

func (obj *GridWithRows) Interface() interface{} {
	return obj.instance
}

func (obj *GridWithRows) IsTruthy() bool {
	return true
}

func (obj *GridWithRows) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", GridWithRowsType, opType)
}

func (obj *GridWithRows) Equals(other object.Object) bool {
	return obj == other
}

func (obj *GridWithRows) Attrs() []object.AttrSpec {
	return gridWithRowsAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *GridWithRows) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *GridWithRows) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", GridWithRowsType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *GridWithRows) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var gridWithRowsAttrs = object.NewAttrRegistry[*GridWithRows](string(GridWithRowsType))

func (obj *GridWithRows) getAttrUI(name string) (object.Object, bool) {
	return gridWithRowsAttrs.GetAttr(obj, name)

}

func NewGridWithRows(rows int, objects ...fyne.CanvasObject) *GridWithRows {
	return &GridWithRows{
		instance: container.NewGridWithRows(rows, objects...),
	}
}
func init() {
	objattrs.ContainerOps(gridWithColumnsAttrs)
}
func init() {
	objattrs.ContainerOps(gridWithRowsAttrs)
}
