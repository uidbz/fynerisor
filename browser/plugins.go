package browser

import "fyne.io/fyne/v2"

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
	// ShowAuthDialog displays login UI in the provided container
	// Returns username, password, and whether the form was submitted
	ShowAuthDialog(url string, contentContainer *fyne.Container) (username, password string, submitted bool)
}
