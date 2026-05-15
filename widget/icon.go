package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Icon{}

const IconType object.Type = "widget.Icon"

// Icon wraps widget.Icon for Risor scripting
type Icon struct {
	instance *widget.Icon
}

// NewIcon creates a new icon widget with a theme resource
func NewIcon(resourceName string) (*Icon, error) {
	res := GetThemeResource(resourceName)
	if res == nil {
		return nil, fmt.Errorf("unknown icon resource: %s", resourceName)
	}
	return &Icon{instance: widget.NewIcon(res)}, nil
}

// GetThemeResource maps string names to theme icon resources
func GetThemeResource(name string) fyne.Resource {
	switch name {
	// Navigation
	case "cancel", "Cancel":
		return theme.CancelIcon()
	case "confirm", "Confirm":
		return theme.ConfirmIcon()
	case "delete", "Delete":
		return theme.DeleteIcon()
	case "search", "Search":
		return theme.SearchIcon()
	case "searchReplace", "SearchReplace":
		return theme.SearchReplaceIcon()

	// Actions
	case "check", "Check":
		return theme.CheckButtonIcon()
	case "add", "Add":
		return theme.ContentAddIcon()
	case "remove", "Remove":
		return theme.ContentRemoveIcon()
	case "cut", "Cut":
		return theme.ContentCutIcon()
	case "copy", "Copy":
		return theme.ContentCopyIcon()
	case "paste", "Paste":
		return theme.ContentPasteIcon()
	case "undo", "Undo":
		return theme.ContentUndoIcon()
	case "redo", "Redo":
		return theme.ContentRedoIcon()

	// Media
	case "mediaPlay", "MediaPlay":
		return theme.MediaPlayIcon()
	case "mediaPause", "MediaPause":
		return theme.MediaPauseIcon()
	case "mediaStop", "MediaStop":
		return theme.MediaStopIcon()
	case "mediaRecord", "MediaRecord":
		return theme.MediaRecordIcon()
	case "mediaReplay", "MediaReplay":
		return theme.MediaReplayIcon()
	case "mediaSkipNext", "MediaSkipNext":
		return theme.MediaSkipNextIcon()
	case "mediaSkipPrevious", "MediaSkipPrevious":
		return theme.MediaSkipPreviousIcon()

	// Files
	case "file", "File":
		return theme.FileIcon()
	case "folder", "Folder":
		return theme.FolderIcon()
	case "folderOpen", "FolderOpen":
		return theme.FolderOpenIcon()
	case "document", "Document":
		return theme.DocumentIcon()
	case "documentCreate", "DocumentCreate":
		return theme.DocumentCreateIcon()
	case "documentPrint", "DocumentPrint":
		return theme.DocumentPrintIcon()
	case "documentSave", "DocumentSave":
		return theme.DocumentSaveIcon()

	// Navigation arrows
	case "navigateBack", "NavigateBack":
		return theme.NavigateBackIcon()
	case "navigateNext", "NavigateNext":
		return theme.NavigateNextIcon()
	case "arrowUp", "ArrowUp":
		return theme.MoveUpIcon()
	case "arrowDown", "ArrowDown":
		return theme.MoveDownIcon()

	// Info
	case "info", "Info":
		return theme.InfoIcon()
	case "question", "Question":
		return theme.QuestionIcon()
	case "warning", "Warning":
		return theme.WarningIcon()
	case "error", "Error":
		return theme.ErrorIcon()

	// Misc
	case "settings", "Settings":
		return theme.SettingsIcon()
	case "home", "Home":
		return theme.HomeIcon()
	case "help", "Help":
		return theme.HelpIcon()
	case "history", "History":
		return theme.HistoryIcon()
	case "mail", "Mail":
		return theme.MailComposeIcon()
	case "visibility", "Visibility":
		return theme.VisibilityIcon()
	case "visibilityOff", "VisibilityOff":
		return theme.VisibilityOffIcon()

	default:
		return nil
	}
}

func (obj *Icon) Type() object.Type {
	return IconType
}

func (obj *Icon) Inspect() string {
	return "widget.Icon()"
}

func (obj *Icon) Interface() interface{} {
	return obj.instance
}

func (obj *Icon) IsTruthy() bool {
	return true
}

func (obj *Icon) Cost() int {
	return 0
}

func (obj *Icon) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Icon) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Icon'")
}

func (obj *Icon) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(IconType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", IconType, opType)
	return errObj, err
}

func (obj *Icon) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Icon) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Icon) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "SetResource":
		return object.NewBuiltin("widget.Icon.SetResource", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			resourceName, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			res := GetThemeResource(resourceName)
			if res == nil {
				return object.Errorf("unknown icon resource: %s", resourceName), nil
			}
			fyne.Do(func() {
				obj.instance.SetResource(res)
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (obj *Icon) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", IconType, name)
}
