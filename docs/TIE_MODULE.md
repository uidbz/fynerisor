# Tie Module

The **tie module** provides Risor bindings for [Tie](https://github.com/uidbz/tie), a triple store database designed for storing and querying relationships between entities. Triple stores excel at representing graph-like data structures where information is stored as `(key, relation, value)` triples.

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Enabling the Module](#enabling-the-module)
- [Connection](#connection)
- [Core Operations](#core-operations)
- [Query Operations](#query-operations)
- [Batch Operations](#batch-operations)
- [Utility Operations](#utility-operations)
- [Tables](#tables)
- [Backup and Restore](#backup-and-restore)
- [Data Model](#data-model)
- [Complete Examples](#complete-examples)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Overview

The tie module allows Risor scripts to:
- Store data as **triples** (key-relation-value)
- Perform **forward associations** (get all attributes of a key)
- Perform **reverse associations** (find all keys with a specific value)
- Execute **batch operations** for efficient bulk writes
- **Stream large datasets** with memory-efficient callbacks
- **Backup and restore** entire collections

**Triple Store Model:**
```
key          relation      value
-----------------------------------------
rust-guide | tag         | programming
rust-guide | tag         | tutorial
rust-guide | author      | Alex Chen
```

## Requirements

The tie module requires a running **tie-triplestore** server. Install and start it from the [Tie repository](https://github.com/uidbz/tie):

```bash
# Clone and build
git clone https://github.com/uidbz/tie
cd tie

# Option 1: Use test environment (includes default config)
./test-env/build.sh
./test-env/start.sh

# Option 2: Run the triplestore server directly
go run ./cmd/tie-triplestore
```

The server listens on `http://localhost:1161` by default.

## Enabling the Module

### GUI Applications

```go
package main

import (
    "github.com/uidbz/fynerisor/gui"
)

func main() {
    fw := gui.NewApp("My App",
        gui.WithTie(),  // Enable tie module
    )
    
    fw.LoadScript(`
        require(["v0.7", "@tie"])
        let db = tie.connect("http://localhost:1161")
        // ... use db
    `)
    fw.Execute()
    fw.ShowAndRun()
}
```

### Headless (Core) Applications

```go
package main

import (
    "github.com/uidbz/fynerisor/core"
)

func main() {
    ctx := core.NewContext(
        core.WithTie(),  // Enable tie module
    )
    
    result, err := ctx.Eval(`
        require(["v0.7", "@tie"])
        let db = tie.connect("http://localhost:1161")
        db.add("example", "tag", "test")
        db.sync()
        return "success"
    `)
    // ...
}
```

## Connection

### tie.connect(url, options?)

Establishes a connection to a tie daemon.

**Parameters:**
- `url` (string): WebService URL of the tie daemon
- `options` (map, optional): Connection configuration

**Options:**
| Option | Type | Description |
|--------|------|-------------|
| `username` | string | Authentication username (default: from config) |
| `password` | string | Authentication password (default: from config) |
| `namespace` | string | Database namespace (default: "default") |
| `collection` | string | Collection name (default: "default") |
| `insecure` | bool | Allow insecure TLS connections (default: false) |

**Returns:** `tie.client` object

**Examples:**

```js
// Basic connection (uses defaults)
let db = tie.connect("http://localhost:1161")

// With authentication
let db = tie.connect("http://localhost:1161", {
    username: "admin",
    password: "secret123"
})

// Custom namespace and collection
let db = tie.connect("http://localhost:1161", {
    namespace: "production",
    collection: "users"
})

// Remote connection with custom config
let db = tie.connect("https://tie.example.com:8443", {
    username: "apiuser",
    password: "apikey",
    namespace: "app",
    collection: "documents",
    insecure: false
})
```

## Core Operations

### db.add(key, relation, value)

Adds a triple to the store.

**Parameters:**
- `key` (string): The entity identifier
- `relation` (string): The attribute/relationship name
- `value` (string): The attribute value

**Returns:** `nil`

**Example:**
```js
db.add("rust-guide", "tag", "programming")
db.add("rust-guide", "tag", "tutorial")
db.add("rust-guide", "author", "Alex Chen")
db.sync()  // Commit the changes
```

### db.get(key)

Retrieves all attributes for a key (forward association).

**Parameters:**
- `key` (string): The entity identifier

**Returns:** Document object or `nil` if not found

**Document structure:**
```js
{
    key: "rust-guide",
    attributes: {
        tag: ["programming", "tutorial"],
        author: ["Alex Chen"]
    }
}
```

**Example:**
```js
let doc = db.get("rust-guide")
if (doc != nil) {
    print("Key:", doc.key)
    print("Tags:", doc.attributes.get("tag"))
    print("Author:", doc.attributes.get("author"))
}
```

### db.set(key, relation, values)

Replaces all values for a specific relation on a key. This is different from `add()` which appends values.

**Parameters:**
- `key` (string): The entity identifier
- `relation` (string): The attribute/relationship name
- `values` (list of strings): New values (replaces existing)

**Returns:** `nil`

**Example:**
```js
// Replace all tags
db.set("rust-guide", "tag", ["programming", "tutorial", "beginner"])
db.sync()

// Clear a relation (set to empty list)
db.set("rust-guide", "outdated", [])
db.sync()
```

### db.delete(key, relation, value)

Deletes a specific triple.

**Parameters:**
- `key` (string): The entity identifier
- `relation` (string): The attribute/relationship name
- `value` (string): The specific value to remove

**Returns:** `nil`

**Example:**
```js
// Remove one specific tag
db.delete("rust-guide", "tag", "tutorial")
db.sync()
```

### db.update(key, relation, oldValue, newValue)

Updates a value in a triple.

**Parameters:**
- `key` (string): The entity identifier
- `relation` (string): The attribute/relationship name
- `oldValue` (string): Current value
- `newValue` (string): New value

**Returns:** `nil`

**Example:**
```js
db.update("rust-guide", "tag", "tutorial", "beginner")
db.sync()
```

### db.sync()

Commits pending writes to the database. Operations like `add()`, `set()`, `delete()`, and `update()` are buffered until `sync()` is called.

**Returns:** `nil`

**Example:**
```js
db.add("doc1", "tag", "test")
db.add("doc2", "tag", "test")
db.add("doc3", "tag", "test")
db.sync()  // All three writes committed atomically
```

## Query Operations

### db.query(spec)

Performs a query with various filtering and sorting options. Most commonly used for **reverse queries** (finding keys by value).

**Parameters:**
- `spec` (map): Query specification

**Query Specification:**
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `terms` | list of strings | `[]` | Search terms (values to find) |
| `exclude` | list of strings | `[]` | Terms to exclude |
| `scope` | string | `""` | Scope filter |
| `filter` | string | `""` | Additional filter expression |
| `sort_by` | string | `""` | Sort field name |
| `reverse` | bool | `false` | **Reverse query**: find keys with these values |
| `expand` | bool | `false` | Include full document attributes |
| `offset` | int | `0` | Pagination offset |
| `limit` | int | `0` | Max results (0 = unlimited) |

**Returns:** Result object
```js
{
    rows: [/* array of documents */],
    total_count: 42
}
```

**Important:** Reverse queries only work for relations configured in the daemon's `ReverseRelations` list. The `"tag"` relation is in the default config.

**Examples:**

```js
// Find all keys with "programming" tag (reverse query)
let result = db.query({
    terms: ["programming"],
    reverse: true
})
print("Found", result.total_count, "documents")
result.rows.each(doc => {
    print(doc.key)
})

// Reverse query with full attributes
let result = db.query({
    terms: ["programming"],
    reverse: true,
    expand: true
})
result.rows.each(doc => {
    print(doc.key, ":", doc.attributes.get("tag"))
})

// Paginated query
let result = db.query({
    terms: ["programming"],
    reverse: true,
    offset: 0,
    limit: 10
})
print("Page 1 of", math.ceil(result.total_count / 10))

// Exclude certain terms
let result = db.query({
    terms: ["programming"],
    exclude: ["tutorial"],
    reverse: true
})

// Sorted results
let result = db.query({
    terms: ["programming"],
    reverse: true,
    sort_by: "key",
    reverse: false  // sort ascending
})
```

### db.expand(keys)

Retrieves multiple documents in a single call.

**Parameters:**
- `keys` (list of strings): Entity identifiers to fetch

**Returns:** List of document objects

**Example:**
```js
let docs = db.expand(["rust-guide", "go-concurrency", "python-async"])
docs.each(doc => {
    print(doc.key, ":", doc.attributes.get("tag"))
})
```

### db.associated(key)

Gets all triples that reference a key (both forward and backward associations).

**Parameters:**
- `key` (string): Entity identifier

**Returns:** List of triple maps

**Example:**
```js
let triples = db.associated("rust-guide")
triples.each(triple => {
    print(triple.key, "|", triple.value1, "|", triple.value2)
})
// Output: rust-guide | tag | programming
//         rust-guide | tag | tutorial
//         rust-guide | author | Alex Chen
```

### db.exists(key)

Checks if a key exists in the database.

**Parameters:**
- `key` (string): Entity identifier

**Returns:** `bool`

**Example:**
```js
if (db.exists("rust-guide")) {
    print("Document exists")
} else {
    print("Not found")
}
```

## Batch Operations

Batch operations allow efficient bulk writes by grouping multiple operations into a single transaction.

### db.batch()

Creates a new batch operation object.

**Returns:** `tie.batch` object

### batch.add(key, relation, value)

Adds a triple to the batch.

**Parameters:**
- `key` (string): Entity identifier
- `relation` (string): Attribute name
- `value` (string): Attribute value

**Returns:** `nil`

### batch.set(key, relation, values)

Adds a set operation to the batch (replaces all values for the relation).

**Parameters:**
- `key` (string): Entity identifier
- `relation` (string): Attribute name
- `values` (list of strings): New values

**Returns:** `nil`

### batch.delete(key, relation, value)

Adds a delete operation to the batch.

**Parameters:**
- `key` (string): Entity identifier
- `relation` (string): Attribute name
- `value` (string): Value to remove

**Returns:** `nil`

### batch.update(key, relation, oldValue, newValue)

Adds an update operation to the batch.

**Parameters:**
- `key` (string): Entity identifier
- `relation` (string): Attribute name
- `oldValue` (string): Current value
- `newValue` (string): New value

**Returns:** `nil`

### batch.run()

Executes all batched operations atomically.

**Returns:** `nil`

**Complete Example:**
```js
let batch = db.batch()

// Queue multiple operations
batch.add("python-async", "tag", "python")
batch.add("python-async", "tag", "programming")
batch.add("python-async", "tag", "async")
batch.set("python-async", "author", ["Pat Wilson"])
batch.update("rust-guide", "tag", "tutorial", "beginner")
batch.delete("old-doc", "tag", "obsolete")

// Execute all at once
batch.run()

print("Batch complete - all operations committed atomically")
```

**Benefits:**
- **Performance**: Reduces network round-trips
- **Atomicity**: All operations succeed or fail together
- **Efficiency**: Single transaction overhead

## Utility Operations

### db.drop()

Deletes the entire collection. **Use with extreme caution!**

**Returns:** `nil`

**Example:**
```js
// WARNING: This deletes all data in the collection
let confirm = dialog.ShowConfirm("Drop Collection", 
    "Delete all data?", 
    () => {
        db.drop()
        print("Collection dropped")
    }
)
```

## Tables

Store a spreadsheet/CSV as a table (ordered headers + rows) and read it back as
a grid, without hand-rolling a triple encoding. This is a client-side
convenience over the triple primitives — the daemon still stores only triples.
The same wire encoding is used by the Go and Python clients, so a table written
by one is readable by the others.

### db.insert_table(uid, headers, rows)

Writes `headers` + `rows` as a table entity and returns its uid.

- `uid`: pass `""` to mint a fresh uid (returned); pass an existing uid to
  replace that table in place (idempotent re-import — old rows are cleared first).
- `headers`: a list of column-name strings, **or** a list of header rows for a
  multi-row (hierarchical) header — see below.
- `rows`: a list of rows, each a list of cell strings in header order. Short rows
  are padded and cells past the header count are ignored. Empty cells are not
  stored (they read back as `""`).

All cells must be strings — convert numbers with `string(n)` first. Headers must
be unique and none may be named `"tie-type"` (it would collide with the row type
marker); either case raises an error. Targets sheet-sized tables (hundreds to low
thousands of rows).

**Returns:** `string` (the table uid)

**Example:**
```js
let uid = db.insert_table("", ["Name", "Age", "City"], [
    ["Alice", "30", "NYC"],
    ["Bob",   "25", "LA"],
])
```

### Multi-row (hierarchical) headers

Pass header **rows** instead of labels when the top row groups the columns below
it. The list is row-major, so `headers[i][j]` is level `i` of column `j` — the
order the source sheet reads in:

```js
let uid = db.insert_table("", [
    ["Sample", "Temperature (20°C)", "Temperature (20°C)"],
    ["",       "Replicate 1",        "Replicate 2"],
], [
    ["S1", "4.2", "4.4"],
])
```

Header rows must be rectangular, with a merged parent cell already repeated
across its columns and blanks explicit — working out where a merged cell ends is
file-parsing that belongs to your converter, not to storage.

Because a column's name is the key every one of its cells is stored under, each
column still needs exactly one unique string: its non-empty levels joined by
`\x1f` (Unit Separator). That joined key is what `headers` reports, so
`"Sample"` stays `"Sample"` while the replicate columns become
`"Temperature (20°C)\x1fReplicate 1"`. A level may contain neither `\x1f` nor
`\x00`. Join with `" / "` only for display — never as the key, or a real header
containing `" / "` would collide.

Passing a single header row stores exactly what the flat form stores, so there is
no migration and no need to decide up front which form a table uses.

### db.read_table(uid)

Reads a table back as a grid. Returns `nil` if `uid` holds no table.

**Returns:** a map `{headers: [...], header_levels: [[...], ...], rows: [[...], ...]}`
(row-major, header order), or `nil` if not found.

- `headers`: the column keys — one string per column, what each cell is keyed by.
- `header_levels`: the header rows, in the shape `insert_table` accepts. A table
  stored with a single header row reports exactly one row here, so rendering does
  not have to special-case depth.
- `rows`: the cells, row-major in header order.

**Example:**
```js
let t = db.read_table(uid)
print(t["headers"])          // ["Name", "Age", "City"]
print(len(t["header_levels"]))  // 1 for a flat header, 2 for a two-row header
t["rows"].each(row => {
    print(row[0], "is", row[1])
})
```

## Backup and Restore

### db.dump()

Returns all triples in the collection as a list.

**Returns:** List of triple arrays `[[key, relation, value], ...]`

**Example:**
```js
let triples = db.dump()
print("Collection has", len(triples), "triples")

triples.each(triple => {
    print(triple[0], "|", triple[1], "|", triple[2])
})

// Save to file (requires @io module)
io.write_file("backup.json", json.encode(triples))
```

### db.dump_stream(callback)

Streams all triples through a callback function. Memory-efficient for large collections.

**Parameters:**
- `callback` (function): Called for each triple with `[key, relation, value]`

**Returns:** `nil`

**Example:**
```js
let count = 0
let file = io.open("backup.txt", "w")

db.dump_stream((triple) => {
    count = count + 1
    file.write(triple[0] + "|" + triple[1] + "|" + triple[2] + "\n")
})

file.close()
print("Exported", count, "triples")
```

### db.restore(triples)

Restores triples from a backup.

**Parameters:**
- `triples` (list): List of triple arrays `[[key, relation, value], ...]`

**Returns:** `nil`

**Example:**
```js
// Load from file (requires @io module)
let data = io.read_file("backup.json")
let triples = json.decode(data)

db.restore(triples)
print("Restored", len(triples), "triples")
```

**Complete Backup/Restore Example:**
```js
// Backup
print("Backing up...")
let backup = db.dump()
io.write_file("tie-backup-" + string(time.now().unix()) + ".json", 
              json.encode(backup))
print("Backed up", len(backup), "triples")

// Restore
print("Restoring...")
let data = io.read_file("tie-backup-1234567890.json")
let triples = json.decode(data)

// Optional: drop existing data first
db.drop()

db.restore(triples)
print("Restored", len(triples), "triples")
```

## Data Model

### Triple Structure

Tie stores data as **triples**: `(key, relation, value)`

```
Subject (Key) | Predicate (Relation) | Object (Value)
-------------------------------------------------
rust-guide    | tag                  | programming
rust-guide    | tag                  | tutorial
rust-guide    | author               | Alex Chen
go-guide      | tag                  | programming
go-guide      | author               | Jordan Lee
```

### Forward vs Reverse Associations

**Forward association** (`db.get(key)`):
- Given a key, get all its attributes
- Example: "What tags does rust-guide have?"
- Always available for all relations

**Reverse association** (`db.query({terms, reverse: true})`):
- Given a value, find all keys that have it
- Example: "Which documents have the 'programming' tag?"
- **Only works for relations in daemon's `ReverseRelations` config**

### Default Reverse Relations

The tie daemon includes these relations in reverse indexes by default:
- `"tag"` - Most commonly used for categorization
- (check daemon config for complete list)

For other relations (like `"author"`), you must configure the daemon to index them for reverse queries.

### Multi-Value Attributes

A key can have multiple values for the same relation:

```js
db.add("rust-guide", "tag", "programming")
db.add("rust-guide", "tag", "tutorial")
db.add("rust-guide", "tag", "beginner")

let doc = db.get("rust-guide")
print(doc.attributes.get("tag"))  // ["programming", "tutorial", "beginner"]
```

## Complete Examples

### Example 1: Document Tagging System

```js
require(["v0.7", "@tie", "@strings"])

let db = tie.connect("http://localhost:1161")

// Add documents with multiple tags
db.add("rust-guide", "tag", "rust")
db.add("rust-guide", "tag", "programming")
db.add("rust-guide", "tag", "tutorial")
db.add("rust-guide", "author", "Alex Chen")
db.add("rust-guide", "year", "2026")

db.add("go-concurrency", "tag", "golang")
db.add("go-concurrency", "tag", "programming")
db.add("go-concurrency", "tag", "advanced")
db.add("go-concurrency", "author", "Jordan Lee")

db.sync()

// Find all programming documents
let result = db.query({
    terms: ["programming"],
    reverse: true,
    expand: true
})

print("Programming documents:", result.total_count)
result.rows.each(doc => {
    let tags = strings.join(doc.attributes.get("tag"), ", ")
    let author = doc.attributes.get("author")[0]
    print("-", doc.key, "by", author, ":", tags)
})

// Get specific document
let doc = db.get("rust-guide")
print("\nRust Guide details:")
print("  Tags:", strings.join(doc.attributes.get("tag"), ", "))
print("  Author:", doc.attributes.get("author")[0])
print("  Year:", doc.attributes.get("year")[0])
```

### Example 2: Task Management

```js
require(["v0.7", "@tie"])

let db = tie.connect("http://localhost:1161")

// Add tasks
let batch = db.batch()
batch.add("task-001", "title", "Fix login bug")
batch.set("task-001", "status", ["in-progress"])
batch.set("task-001", "priority", ["high"])
batch.set("task-001", "assignee", ["alice"])

batch.add("task-002", "title", "Update documentation")
batch.set("task-002", "status", ["todo"])
batch.set("task-002", "priority", ["low"])
batch.set("task-002", "assignee", ["bob"])

batch.add("task-003", "title", "Deploy to production")
batch.set("task-003", "status", ["in-progress"])
batch.set("task-003", "priority", ["critical"])
batch.set("task-003", "assignee", ["alice"])

batch.run()

// Find Alice's tasks
let aliceTasks = db.query({
    terms: ["alice"],
    reverse: true,
    expand: true
})

print("Alice's tasks:")
aliceTasks.rows.each(task => {
    let title = task.attributes.get("title")[0]
    let status = task.attributes.get("status")[0]
    let priority = task.attributes.get("priority")[0]
    print("-", title, "[" + status + "]", "(" + priority + ")")
})

// Find critical tasks
let critical = db.query({
    terms: ["critical"],
    reverse: true,
    expand: true
})

print("\nCritical tasks:", critical.total_count)
```

### Example 3: Knowledge Graph

```js
require(["v0.7", "@tie", "@strings"])

let db = tie.connect("http://localhost:1161")

// Build a simple knowledge graph
db.add("risor", "type", "programming-language")
db.add("risor", "created-by", "Curtis Myzie")
db.add("risor", "influenced-by", "go")
db.add("risor", "influenced-by", "python")
db.add("risor", "tag", "scripting")
db.add("risor", "tag", "embeddable")

db.add("fynerisor", "type", "library")
db.add("fynerisor", "uses", "risor")
db.add("fynerisor", "uses", "fyne")
db.add("fynerisor", "tag", "gui")

db.add("fyne", "type", "gui-framework")
db.add("fyne", "language", "go")
db.add("fyne", "tag", "cross-platform")

db.sync()

// Explore relationships
function explore(key) {
    let doc = db.get(key)
    if (doc == nil) {
        print("Not found:", key)
        return
    }
    
    print("\n===", key, "===")
    list(doc.attributes.keys()).each(relation => {
        let values = doc.attributes.get(relation)
        print(relation + ":", strings.join(values, ", "))
    })
}

explore("risor")
explore("fynerisor")
explore("fyne")

// Find all libraries
let libs = db.query({
    terms: ["library"],
    reverse: true,
    expand: true
})

print("\n=== Libraries ===")
libs.rows.each(lib => {
    let uses = lib.attributes.get("uses")
    if (uses != nil) {
        print(lib.key, "uses:", strings.join(uses, ", "))
    }
})
```

## Error Handling

Most tie operations return `nil` on success or throw errors on failure.

```js
// Check for nil returns
let doc = db.get("unknown-key")
if (doc == nil) {
    print("Document not found")
} else {
    print("Found:", doc.key)
}

// Query returns empty list if nothing found
let result = db.query({
    terms: ["nonexistent"],
    reverse: true
})
if (len(result.rows) == 0) {
    print("No results")
}

// Operations that can fail
try {
    db.add("key", "relation", "value")
    db.sync()
} catch (err) {
    print("ERROR:", err)
}
```

## Best Practices

### 1. Always Call sync()

Operations are buffered until `sync()`:

```js
// DON'T: Changes aren't committed
db.add("doc1", "tag", "test")
let doc = db.get("doc1")  // Won't see the change yet!

// DO: Commit before reading
db.add("doc1", "tag", "test")
db.sync()
let doc = db.get("doc1")  // Now it's there
```

### 2. Use Batches for Bulk Operations

```js
// DON'T: Multiple round-trips
db.add("doc1", "tag", "test")
db.sync()
db.add("doc2", "tag", "test")
db.sync()
db.add("doc3", "tag", "test")
db.sync()

// DO: Single batch transaction
let batch = db.batch()
batch.add("doc1", "tag", "test")
batch.add("doc2", "tag", "test")
batch.add("doc3", "tag", "test")
batch.run()
```

### 3. Check Reverse Relations Config

Only relations in the daemon's `ReverseRelations` config support reverse queries:

```js
// Works: "tag" is in default config
let docs = db.query({
    terms: ["programming"],
    reverse: true  // Finds docs with tag=programming
})

// Won't work: "author" not in default config
let authorDocs = db.query({
    terms: ["Alex Chen"],
    reverse: true  // Returns empty - relation not indexed!
})
```

If you need reverse queries on custom relations, configure the daemon.

### 4. Use expand for Complete Data

```js
// Without expand: only keys returned
let result = db.query({
    terms: ["programming"],
    reverse: true
})
result.rows.each(row => {
    // row only has: {key: "rust-guide"}
    let doc = db.get(row.key)  // Extra query needed!
})

// With expand: full documents returned
let result = db.query({
    terms: ["programming"],
    reverse: true,
    expand: true
})
result.rows.each(row => {
    // row has: {key: "rust-guide", attributes: {...}}
    print(row.attributes.get("tag"))  // No extra query!
})
```

### 5. Use dump_stream() for Large Datasets

```js
// DON'T: Load everything into memory
let triples = db.dump()  // Could be millions of triples!

// DO: Stream for large collections
let count = 0
db.dump_stream((triple) => {
    count = count + 1
    // Process one at a time
})
print("Processed", count, "triples")
```

### 6. Pagination for Large Result Sets

```js
let pageSize = 20
let page = 0

function loadPage() {
    let result = db.query({
        terms: ["programming"],
        reverse: true,
        expand: true,
        offset: page * pageSize,
        limit: pageSize
    })
    
    print("Page", page + 1, "of", 
          math.ceil(result.total_count / pageSize))
    
    return result
}

let results = loadPage()
```

### 7. Key Naming Conventions

Use consistent, descriptive keys:

```js
// Good: Descriptive and scoped
db.add("user-alice-123", "name", "Alice")
db.add("task-2026-001", "title", "Deploy app")
db.add("doc-rust-guide", "type", "tutorial")

// Avoid: Too generic or cryptic
db.add("alice", "name", "Alice")  // Could conflict
db.add("001", "title", "Deploy")  // What is this?
db.add("x123", "type", "tutorial")  // Unclear
```

---

## See Also

- [Tie Repository](https://github.com/uidbz/tie) - Upstream tie database
- [Example 36: Tie Headless](../examples/36-tie-headless/) - Complete CLI example
- [Example 37: Tie GUI](../examples/37-tie-gui/) - Interactive GUI browser
- [HTTP Module](HTTP_MODULE.md) - Remote tie connections
- [SQL Module](SQL_MODULE.md) - Alternative structured storage
