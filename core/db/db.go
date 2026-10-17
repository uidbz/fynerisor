// Package db provides the opt-in SQL module option for fynerisor core contexts.
//
// The database drivers (MySQL, PostgreSQL, SQLite, SQL Server) are pulled into
// a binary only when this package is imported. Headless binaries that do not
// need database access can omit the import and the linker will drop the
// drivers entirely, saving several megabytes.
//
// Previously this was available as core.WithSQL(). That option was removed so
// that importing core no longer forces the DB drivers into every binary. To
// migrate:
//
//	import "github.com/uidbz/fynerisor/core/db"
//
//	ctx := core.NewContext(db.WithSQL())
//
// The gui package still provides gui.WithSQL() directly; GUI applications
// already link dynamically and are unaffected.
package db

import (
	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/modules/sql"
)

// WithSQL enables the SQL module for database connectivity from Risor scripts.
// Supports: MySQL, PostgreSQL, SQLite, SQL Server.
//
// Importing this package (and calling this option) links the database drivers;
// binaries that do not import it are not affected.
//
// Example:
//
//	ctx := core.NewContext(db.WithSQL())
//
// Usage in script:
//
//	require(["@sql"])
//	let conn = sql.connect("sqlite3::memory:")
//	conn.exec("CREATE TABLE users (id INT, name TEXT)")
//	let rows = conn.query("SELECT * FROM users").collect()
func WithSQL() core.Option {
	return core.WithGlobal("sql", sql.Module())
}
