// Package fynerisor provides both headless (core) and GUI scripting capabilities using Risor.
//
// This package re-exports types and functions from the core and gui subpackages for backward compatibility.
//
// For headless execution (no GUI dependencies, enables static compilation):
//   import "github.com/uidbz/fynerisor/core"
//   ctx := core.NewContext(core.WithHTTP(), core.WithSQL())
//   ctx.LoadScript(`print("Hello from headless!")`)
//   ctx.Eval()
//
// For GUI applications (requires Fyne):
//   import "github.com/uidbz/fynerisor/gui"
//   fw := gui.NewApp("My App", gui.WithHTTP())
//   fw.LoadScript(`window.SetContent(widget.NewLabel("Hello!"))`)
//   fw.Execute()
//   fw.ShowAndRun()
//
// Legacy code using github.com/uidbz/fynerisor directly continues to work unchanged.
package fynerisor

import (
	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/gui"
)

// Re-export core types (for headless usage)
type Context = core.Context

// Re-export GUI types (for backward compatibility)
type Window = gui.Window
type Option = gui.Option

// Re-export GUI functions (backward compatible with original API)
var (
	// Window creation
	NewWindow = gui.NewWindow
	NewApp    = gui.NewApp

	// Module options (work with both Context via core and Window via gui)
	WithHTTP     = gui.WithHTTP
	WithOS       = gui.WithOS
	WithSQL      = gui.WithSQL
	WithStrings  = gui.WithStrings
	WithFilepath = gui.WithFilepath
	WithTime     = gui.WithTime
	WithIO       = gui.WithIO
	WithGlobal   = gui.WithGlobal
	WithGlobals  = gui.WithGlobals

	// GUI-specific options
	WithStatusCallback = gui.WithStatusCallback
	WithResultCallback = gui.WithResultCallback
	WithAppName        = gui.WithAppName
)

// Re-export core functions (for direct access)
var (
	// Context creation for headless mode
	NewContext = core.NewContext

	// Version management
	SetAppVersion = core.SetAppVersion

	// Script analysis
	AnalyzeRequirements = core.AnalyzeRequirements
)
