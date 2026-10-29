// Package db provides the opt-in SQL module option for fynerisor GUI
// runtimes (issue #40).
//
// The database drivers (MySQL, PostgreSQL, SQLite, SQL Server) are
// pulled into a binary only when this package is imported. Applications
// that embed gui without database access can omit the import and the
// linker drops the drivers entirely — the same arrangement core uses
// (core/db).
//
// Previously this was gui.WithSQL(), which linked every driver into
// every gui binary. To migrate:
//
//	import "github.com/uidbz/fynerisor/gui/db"
//
//	window := gui.NewWindow(w, db.WithSQL())
package db

import (
	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/modules/sql"
)

// WithSQL enables the SQL module for database connectivity from Risor
// scripts. Supports: MySQL, PostgreSQL, SQLite, SQL Server.
//
// The returned option is a core.Option, so it configures gui runtimes,
// embedded CanvasObject views, and headless core contexts alike.
//
// Usage in script:
//
//	require(["@sql"])
//	let conn = sql.connect("sqlite3://:memory:")
//	conn.exec("CREATE TABLE users (id INT, name TEXT)")
func WithSQL() core.Option {
	return core.Module("sql", sql.Module())
}
