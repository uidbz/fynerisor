package browser

import (
	"context"

	"fyne.io/fyne/v2"
)

// MenuProvider supplies application-specific menu items
type MenuProvider interface {
	GetMenuItems() []MenuItem
}

// MenuItem represents a single menu action
type MenuItem struct {
	Label  string
	Action func()
}

// AuthProvider handles authentication when needed
type AuthProvider interface {
	// ShowAuthDialog displays login UI (in parent's window) and invokes
	// done exactly once: with the entered credentials and ok=true on
	// submit, or ok=false on cancel. It must not block — the browser
	// calls it on the UI thread and continues when done fires (issue #28).
	// ctx is cancelled if the navigation it belongs to was superseded.
	ShowAuthDialog(ctx context.Context, url string, parent fyne.Window, done func(username, password string, ok bool))
}
