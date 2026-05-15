package main

import (
	"context"
	"fmt"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

// UserDatabase is a custom type that will be exposed to Risor scripts
type UserDatabase struct {
	users map[string]User
}

type User struct {
	Name  string
	Email string
	Age   int
}

// NewUserDatabase creates a new user database
func NewUserDatabase() *UserDatabase {
	return &UserDatabase{
		users: make(map[string]User),
	}
}

// Add a user to the database
func (db *UserDatabase) AddUser(name, email string, age int) {
	db.users[name] = User{
		Name:  name,
		Email: email,
		Age:   age,
	}
}

// Get a user from the database
func (db *UserDatabase) GetUser(name string) (User, bool) {
	user, ok := db.users[name]
	return user, ok
}

// List all users
func (db *UserDatabase) ListUsers() []User {
	users := make([]User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}
	return users
}

// Count returns the number of users
func (db *UserDatabase) Count() int {
	return len(db.users)
}

// Delete a user
func (db *UserDatabase) DeleteUser(name string) bool {
	if _, ok := db.users[name]; ok {
		delete(db.users, name)
		return true
	}
	return false
}

// ============================================================================
// Risor Object Implementation
// ============================================================================

// UserDatabaseObject wraps UserDatabase for Risor
type UserDatabaseObject struct {
	db *UserDatabase
}

const UserDatabaseType object.Type = "user_database"

func (obj *UserDatabaseObject) Type() object.Type {
	return UserDatabaseType
}

func (obj *UserDatabaseObject) Inspect() string {
	return fmt.Sprintf("UserDatabase(%d users)", obj.db.Count())
}

func (obj *UserDatabaseObject) Interface() interface{} {
	return obj.db
}

func (obj *UserDatabaseObject) IsTruthy() bool {
	return true
}

func (obj *UserDatabaseObject) Cost() int {
	return 0
}

func (obj *UserDatabaseObject) Equals(other object.Object) bool {
	return obj == other
}

func (obj *UserDatabaseObject) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, fmt.Errorf("unsupported operation for UserDatabase: %v", opType)
}

func (obj *UserDatabaseObject) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("UserDatabase cannot be marshaled to JSON")
}

func (obj *UserDatabaseObject) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "add":
		return object.NewBuiltin("user_database.add", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("add() requires 3 arguments (name, email, age), got %d", len(args))
			}

			name, err := object.AsString(args[0])
			if err != nil {
				return nil, fmt.Errorf("name must be string: %w", err)
			}

			email, err := object.AsString(args[1])
			if err != nil {
				return nil, fmt.Errorf("email must be string: %w", err)
			}

			age, err := object.AsInt(args[2])
			if err != nil {
				return nil, fmt.Errorf("age must be int: %w", err)
			}

			obj.db.AddUser(name, email, int(age))
			return object.Nil, nil
		}), true

	case "get":
		return object.NewBuiltin("user_database.get", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("get() requires 1 argument (name), got %d", len(args))
			}

			name, err := object.AsString(args[0])
			if err != nil {
				return nil, fmt.Errorf("name must be string: %w", err)
			}

			user, ok := obj.db.GetUser(name)
			if !ok {
				return object.Nil, nil
			}

			// Return a map with user data
			return object.NewMap(map[string]object.Object{
				"name":  object.NewString(user.Name),
				"email": object.NewString(user.Email),
				"age":   object.NewInt(int64(user.Age)),
			}), nil
		}), true

	case "list":
		return object.NewBuiltin("user_database.list", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("list() takes no arguments, got %d", len(args))
			}

			users := obj.db.ListUsers()
			userObjects := make([]object.Object, len(users))
			for i, user := range users {
				userObjects[i] = object.NewMap(map[string]object.Object{
					"name":  object.NewString(user.Name),
					"email": object.NewString(user.Email),
					"age":   object.NewInt(int64(user.Age)),
				})
			}

			return object.NewList(userObjects), nil
		}), true

	case "count":
		return object.NewBuiltin("user_database.count", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("count() takes no arguments, got %d", len(args))
			}
			return object.NewInt(int64(obj.db.Count())), nil
		}), true

	case "delete":
		return object.NewBuiltin("user_database.delete", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("delete() requires 1 argument (name), got %d", len(args))
			}

			name, err := object.AsString(args[0])
			if err != nil {
				return nil, fmt.Errorf("name must be string: %w", err)
			}

			deleted := obj.db.DeleteUser(name)
			return object.NewBool(deleted), nil
		}), true
	}

	return nil, false
}

func (obj *UserDatabaseObject) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("UserDatabase attributes are read-only")
}

func (obj *UserDatabaseObject) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "add", Doc: "add(name, email, age) - Add a user to the database"},
		{Name: "get", Doc: "get(name) - Get a user by name"},
		{Name: "list", Doc: "list() - List all users"},
		{Name: "count", Doc: "count() - Get the number of users"},
		{Name: "delete", Doc: "delete(name) - Delete a user"},
	}
}

// NewUserDatabaseObject creates a Risor object wrapping the database
func NewUserDatabaseObject() *UserDatabaseObject {
	return &UserDatabaseObject{
		db: NewUserDatabase(),
	}
}
