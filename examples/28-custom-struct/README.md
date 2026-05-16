# Custom Go Struct Example

Demonstrates how to expose custom Go types with methods as global variables in Risor scripts.

## Overview

This example shows the complete pattern for:
1. Creating a custom Go type (UserDatabase)
2. Implementing the Risor `object.Object` interface
3. Exposing methods via `GetAttr()`
4. Adding it as a global variable with `WithGlobal()`
5. Using it from Risor scripts

**Note:** This example uses Risor v2 syntax with functional iteration (`.each()`) instead of C-style for loops.

## Architecture

### Go Side (custom_types.go)

**1. Define Your Type:**
```go
type UserDatabase struct {
    users map[string]User
}

func (db *UserDatabase) AddUser(name, email string, age int) {
    // Implementation
}
```

**2. Wrap it for Risor:**
```go
type UserDatabaseObject struct {
    db *UserDatabase
}

// Implement object.Object interface
func (obj *UserDatabaseObject) Type() object.Type { ... }
func (obj *UserDatabaseObject) Inspect() string { ... }
// ... etc (10 required methods)
```

**3. Expose Methods via GetAttr:**
```go
func (obj *UserDatabaseObject) GetAttr(name string) (object.Object, bool) {
    switch name {
    case "add":
        return object.NewBuiltin("user_database.add", func(...) {
            // Wrap Go method for Risor
        }), true
    case "get":
        // ...
    }
    return nil, false
}
```

### Application Side (main.go)

**Register as Global:**
```go
userDB := NewUserDatabaseObject()

w := fynerisor.NewApp("My App",
    fynerisor.WithGlobal("users", userDB),
)
```

### Script Side (script.risor)

**Use the Global:**
```js
// The 'users' global is now available
users.add("Alice", "alice@example.com", 28)
let user = users.get("Alice")
let count = users.count()

// Iterate using Risor v2 functional methods
let userList = users.list()
userList.each(user => {
    print(sprintf("%s - %s", user["name"], user["email"]))
})
```

## Required Interface Methods

To expose a custom type, implement these 10 methods from `object.Object`:

```go
Type() object.Type                           // Return your type name
Inspect() string                              // Return debug string
Interface() interface{}                       // Return wrapped object
IsTruthy() bool                               // For boolean context
Cost() int                                    // Memory cost (usually 0)
Equals(object.Object) bool                    // Equality check
RunOperation(op, object.Object) (...)         // Binary operations
MarshalJSON() ([]byte, error)                 // JSON serialization
GetAttr(string) (object.Object, bool)         // ⭐ Expose methods here
SetAttr(string, object.Object) error          // For writable attributes
Attrs() []object.AttrSpec                     // List available attributes
```

## Method Patterns

### Simple Method (no return value)
```go
case "add":
    return object.NewBuiltin("db.add", func(ctx context.Context, args ...object.Object) (object.Object, error) {
        // Parse arguments
        name, _ := object.AsString(args[0])
        
        // Call Go method
        obj.db.AddUser(name, ...)
        
        return object.Nil, nil
    }), true
```

### Method Returning Data
```go
case "get":
    return object.NewBuiltin("db.get", func(...) (object.Object, error) {
        user, ok := obj.db.GetUser(name)
        if !ok {
            return object.Nil, nil  // Not found
        }
        
        // Return as map
        return object.NewMap(map[string]object.Object{
            "name":  object.NewString(user.Name),
            "email": object.NewString(user.Email),
        }), nil
    }), true
```

### Method Returning List
```go
case "list":
    return object.NewBuiltin("db.list", func(...) (object.Object, error) {
        users := obj.db.ListUsers()
        
        objs := make([]object.Object, len(users))
        for i, user := range users {
            objs[i] = object.NewMap(...) // User as map
        }
        
        return object.NewList(objs), nil
    }), true
```

## Type Conversions

**Risor → Go:**
```go
str, err := object.AsString(arg)
num, err := object.AsInt(arg)
flt, err := object.AsFloat(arg)
bln, err := object.AsBool(arg)
```

**Go → Risor:**
```go
object.NewString("text")
object.NewInt(42)
object.NewFloat(3.14)
object.NewBool(true)
object.NewMap(map[string]object.Object{...})
object.NewList([]object.Object{...})
object.Nil  // For null/none
```

## Use Cases

1. **Database Connections** - Expose DB methods to scripts
2. **External APIs** - Wrap API clients
3. **Business Logic** - Domain-specific operations
4. **State Management** - Shared application state
5. **Configuration** - Complex config objects

## Run

```bash
go run main.go
```

The example demonstrates:
- Adding users via form
- Listing all users
- Searching by name
- Deleting users
- Demo data population

All operations call Go methods on the custom UserDatabase type!

## Debugging

Use `WithStatusCallback()` to capture script errors:

```go
w := fynerisor.NewApp("My App",
    fynerisor.WithGlobal("db", db),
    fynerisor.WithStatusCallback(func(status string) {
        log.Printf("Status: %s", status)
    }),
)
```

This helps identify Risor syntax errors or runtime issues.

## Key Takeaways

✅ Implement `object.Object` interface (10 methods)  
✅ Expose methods via `GetAttr()` returning builtins  
✅ Register with `WithGlobal(name, object)`  
✅ Convert types between Risor ↔ Go  
✅ Handle errors gracefully  
✅ Use Risor v2 functional iteration (`.each()`, `.map()`) instead of for loops  
✅ Add status callbacks for debugging script errors  

This pattern lets you extend fynerisor with **any** Go functionality!
