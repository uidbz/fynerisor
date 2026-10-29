package gui

import (
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"fyne.io/fyne/v2/widget"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const ConstantsType object.Type = "constants"

// Constants provides access to Fyne constants and enums in Risor scripts.
// It exposes all Fyne enumeration values as integer constants.
//
// Available in scripts as the global 'constants' object:
//
//	// Button importance
//	btn.Importance = constants.ImportanceHigh
//	btn.Importance = constants.SuccessImportance
//
//	// Text wrapping
//	label.Wrapping = constants.TextWrapWord
//
//	// Scroll direction
//	container.Scroll = constants.ScrollVerticalOnly
//
// Button Importance (visual hierarchy):
//   - ImportanceHigh, ImportanceMedium, ImportanceLow
//
// Button Importance (semantic):
//   - SuccessImportance, WarningImportance, DangerImportance
//
// Button Icon Placement:
//   - ButtonIconLeadingText, ButtonIconTrailingText
//
// Button Alignment:
//   - ButtonAlignCenter, ButtonAlignLeading, ButtonAlignTrailing
//
// Text Wrapping:
//   - TextWrapOff, TextWrapBreak, TextWrapWord
//
// Text Truncation:
//   - TextTruncateOff, TextTruncateClip, TextTruncateEllipsis
//
// Scroll Direction:
//   - ScrollBoth, ScrollHorizontalOnly, ScrollVerticalOnly, ScrollNone
//
// Orientation:
//   - Horizontal, Vertical
type Constants struct{}

func (obj *Constants) Type() object.Type {
	return ConstantsType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Constants) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Constants) inspectUI() string {
	return "constants"
}

func (obj *Constants) Interface() interface{} {
	return nil
}

func (obj *Constants) IsTruthy() bool {
	return true
}

func (obj *Constants) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ConstantsType, opType)
}

func (obj *Constants) Equals(other object.Object) bool {
	return obj == other
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Constants) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Constants) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object is read-only", ConstantsType)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Constants) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var constantsAttrs = objattrs.Register(string(ConstantsType), object.NewAttrRegistry[*Constants](string(ConstantsType)))

func (obj *Constants) getAttrUI(name string) (object.Object, bool) {
	// Constant groups: Button Importance, Button Icon Placement,
	// Button Alignment, Text Wrapping (fyne.TextWrap), Text Truncation
	// (fyne.TextTruncation), Scroll Direction (fyne.ScrollDirection),
	// Orientation (fyne.Orientation) and Text Alignment (fyne.TextAlign).
	return constantsAttrs.GetAttr(obj, name)
}

func (obj *Constants) Attrs() []object.AttrSpec {
	return constantsAttrs.Specs()
}

func newConstantsObject() *Constants {
	return &Constants{}
}
func init() {
	constantsAttrs.
		Define(
			"ImportanceHigh",
		).Getter(func(obj *Constants) object.
		Object {
		return object.NewInt(int64(widget.HighImportance))
	})
	constantsAttrs.
		Define(
			"ImportanceMedium",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(int64(widget.MediumImportance))
	})
	constantsAttrs.
		Define(
			"ImportanceLow",
		).Getter(func(obj *Constants) object.
		Object {
		return object.NewInt(int64(widget.LowImportance))
	})
	constantsAttrs.
		Define(
			"DangerImportance",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(int64(widget.DangerImportance))
	})
	constantsAttrs.
		Define(
			"WarningImportance",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(int64(widget.WarningImportance))
	})
	constantsAttrs.
		Define(
			"SuccessImportance",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(int64(widget.SuccessImportance))
	})
	constantsAttrs.
		Define(
			"ButtonIconLeadingText",
		).Getter(func(obj *Constants) object.Object {
		return object.
			NewInt(int64(widget.ButtonIconLeadingText))
	})
	constantsAttrs.
		Define(
			"ButtonIconTrailingText",
		).Getter(func(obj *Constants) object.Object {
		return object.
			NewInt(int64(widget.ButtonIconTrailingText))
	})
	constantsAttrs.
		Define(
			"ButtonAlignCenter",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(int64(widget.ButtonAlignCenter))
	})
	constantsAttrs.
		Define(
			"ButtonAlignLeading",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(int64(widget.ButtonAlignLeading))
	})
	constantsAttrs.
		Define(
			"ButtonAlignTrailing",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(int64(widget.ButtonAlignTrailing))
	})
	constantsAttrs.
		Define(
			"TextWrapOff",
		).Getter(func(obj *Constants) object.Object {
		return object.NewInt(0)
	})
	constantsAttrs.
		Define(
			"TextWrapBreak",
		).Getter(func(obj *Constants) object.
		Object {
		return object.NewInt(1)
	})
	constantsAttrs.
		Define(
			"TextWrapWord",
		).Getter(func(obj *Constants) object.
		Object {
		return object.NewInt(2)
	})
	constantsAttrs.
		Define(
			"TextTruncateOff",
		).Getter(func(obj *Constants) object.
		Object {
		return object.NewInt(0)
	})
	constantsAttrs.
		Define(
			"TextTruncateClip",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(1)
	})
	constantsAttrs.
		Define(
			"TextTruncateEllipsis",
		).Getter(func(obj *Constants) object.Object {
		return object.
			NewInt(2)
	})
	constantsAttrs.
		Define(
			"ScrollBoth",
		).Getter(func(obj *Constants) object.Object {
		return object.NewInt(0)
	})
	constantsAttrs.
		Define(
			"ScrollHorizontalOnly",
		).Getter(func(obj *Constants) object.Object {
		return object.
			NewInt(1)
	})
	constantsAttrs.
		Define(
			"ScrollVerticalOnly",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(2)
	})
	constantsAttrs.
		Define(
			"ScrollNone",
		).Getter(func(obj *Constants) object.Object {
		return object.NewInt(3)
	})
	constantsAttrs.
		Define(
			"Horizontal",
		).Getter(func(obj *Constants) object.Object {
		return object.NewInt(0)
	})
	constantsAttrs.
		Define(
			"Vertical",
		).Getter(func(obj *Constants) object.Object {
		return object.NewInt(1)
	})
	constantsAttrs.
		Define(
			"TextAlignLeading",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(0)
	})
	constantsAttrs.
		Define(
			"TextAlignCenter",
		).Getter(func(obj *Constants) object.
		Object {
		return object.NewInt(1)
	})
	constantsAttrs.
		Define(
			"TextAlignTrailing",
		).Getter(func(obj *Constants) object.
		Object {
		return object.
			NewInt(2)
	})
}
