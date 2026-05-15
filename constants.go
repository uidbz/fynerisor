package fynerisor

import (
	"fmt"

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

func (obj *Constants) Inspect() string {
	return "constants"
}

func (obj *Constants) Interface() interface{} {
	return nil
}

func (obj *Constants) IsTruthy() bool {
	return true
}

func (obj *Constants) Cost() int {
	return 0
}

func (obj *Constants) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'constants'")
}

func (obj *Constants) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for %s: %v", ConstantsType, opType), nil
}

func (obj *Constants) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Constants) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object is read-only", ConstantsType)
}

func (obj *Constants) GetAttr(name string) (object.Object, bool) {
	switch name {
	// Button Importance
	case "ImportanceHigh":
		return object.NewInt(int64(widget.HighImportance)), true
	case "ImportanceMedium":
		return object.NewInt(int64(widget.MediumImportance)), true
	case "ImportanceLow":
		return object.NewInt(int64(widget.LowImportance)), true
	case "DangerImportance":
		return object.NewInt(int64(widget.DangerImportance)), true
	case "WarningImportance":
		return object.NewInt(int64(widget.WarningImportance)), true
	case "SuccessImportance":
		return object.NewInt(int64(widget.SuccessImportance)), true

	// Button Icon Placement
	case "ButtonIconLeadingText":
		return object.NewInt(int64(widget.ButtonIconLeadingText)), true
	case "ButtonIconTrailingText":
		return object.NewInt(int64(widget.ButtonIconTrailingText)), true

	// Button Alignment
	case "ButtonAlignCenter":
		return object.NewInt(int64(widget.ButtonAlignCenter)), true
	case "ButtonAlignLeading":
		return object.NewInt(int64(widget.ButtonAlignLeading)), true
	case "ButtonAlignTrailing":
		return object.NewInt(int64(widget.ButtonAlignTrailing)), true

	// Text Wrapping (fyne.TextWrap)
	case "TextWrapOff":
		return object.NewInt(0), true // fyne.TextWrapOff
	case "TextWrapBreak":
		return object.NewInt(1), true // fyne.TextWrapBreak
	case "TextWrapWord":
		return object.NewInt(2), true // fyne.TextWrapWord

	// Text Truncation (fyne.TextTruncation)
	case "TextTruncateOff":
		return object.NewInt(0), true // fyne.TextTruncateOff
	case "TextTruncateClip":
		return object.NewInt(1), true // fyne.TextTruncateClip
	case "TextTruncateEllipsis":
		return object.NewInt(2), true // fyne.TextTruncateEllipsis

	// Scroll Direction (fyne.ScrollDirection)
	case "ScrollBoth":
		return object.NewInt(0), true
	case "ScrollHorizontalOnly":
		return object.NewInt(1), true
	case "ScrollVerticalOnly":
		return object.NewInt(2), true
	case "ScrollNone":
		return object.NewInt(3), true

	// Orientation (fyne.Orientation)
	case "Horizontal":
		return object.NewInt(0), true
	case "Vertical":
		return object.NewInt(1), true

	// Text Alignment (fyne.TextAlign)
	case "TextAlignLeading":
		return object.NewInt(0), true
	case "TextAlignCenter":
		return object.NewInt(1), true
	case "TextAlignTrailing":
		return object.NewInt(2), true
	}
	return nil, false
}

func (obj *Constants) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		// Button Importance
		{Name: "ImportanceHigh"},
		{Name: "ImportanceMedium"},
		{Name: "ImportanceLow"},
		{Name: "DangerImportance"},
		{Name: "WarningImportance"},
		{Name: "SuccessImportance"},
		// Button Icon Placement
		{Name: "ButtonIconLeadingText"},
		{Name: "ButtonIconTrailingText"},
		// Button Alignment
		{Name: "ButtonAlignCenter"},
		{Name: "ButtonAlignLeading"},
		{Name: "ButtonAlignTrailing"},
		// Text Wrapping
		{Name: "TextWrapOff"},
		{Name: "TextWrapBreak"},
		{Name: "TextWrapWord"},
		// Text Truncation
		{Name: "TextTruncateOff"},
		{Name: "TextTruncateClip"},
		{Name: "TextTruncateEllipsis"},
		// Scroll Direction
		{Name: "ScrollBoth"},
		{Name: "ScrollHorizontalOnly"},
		{Name: "ScrollVerticalOnly"},
		{Name: "ScrollNone"},
		// Orientation
		{Name: "Horizontal"},
		{Name: "Vertical"},
		// Text Alignment
		{Name: "TextAlignLeading"},
		{Name: "TextAlignCenter"},
		{Name: "TextAlignTrailing"},
	}
}

func newConstantsObject() *Constants {
	return &Constants{}
}
