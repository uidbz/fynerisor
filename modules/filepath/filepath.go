// Package filepath provides file path manipulation functions for Risor scripts.
package filepath

import (
	"context"
	"path/filepath"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Module returns the filepath module for risor
func Module() *object.Module {
	return object.NewBuiltinsModule("filepath", map[string]object.Object{
		"join":  object.NewBuiltin("filepath.join", join),
		"base":  object.NewBuiltin("filepath.base", base),
		"dir":   object.NewBuiltin("filepath.dir", dir),
		"ext":   object.NewBuiltin("filepath.ext", ext),
		"abs":   object.NewBuiltin("filepath.abs", abs),
		"clean": object.NewBuiltin("filepath.clean", clean),
	})
}

func join(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 {
		return object.Errorf("filepath.join: expected at least 1 argument, got %d", len(args)), nil
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		s, err := object.AsString(arg)
		if err != nil {
			return nil, err
		}
		parts[i] = s
	}

	return object.NewString(filepath.Join(parts...)), nil
}

func base(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("filepath.base: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	return object.NewString(filepath.Base(path)), nil
}

func dir(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("filepath.dir: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	return object.NewString(filepath.Dir(path)), nil
}

func ext(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("filepath.ext: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	return object.NewString(filepath.Ext(path)), nil
}

func abs(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("filepath.abs: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return object.NewError(err), nil
	}

	return object.NewString(absPath), nil
}

func clean(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("filepath.clean: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	return object.NewString(filepath.Clean(path)), nil
}
