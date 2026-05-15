// Package strings provides string manipulation functions for Risor scripts.
package strings

import (
	"context"
	"strings"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Module returns the strings module for risor
func Module() *object.Module {
	return object.NewBuiltinsModule("strings", map[string]object.Object{
		"replace_all":  object.NewBuiltin("strings.replace_all", replaceAll),
		"trim_prefix":  object.NewBuiltin("strings.trim_prefix", trimPrefix),
		"trim_suffix":  object.NewBuiltin("strings.trim_suffix", trimSuffix),
		"trim":         object.NewBuiltin("strings.trim", trim),
		"to_lower":     object.NewBuiltin("strings.to_lower", toLower),
		"to_upper":     object.NewBuiltin("strings.to_upper", toUpper),
		"split":        object.NewBuiltin("strings.split", split),
		"join":         object.NewBuiltin("strings.join", join),
		"contains":     object.NewBuiltin("strings.contains", contains),
		"has_prefix":   object.NewBuiltin("strings.has_prefix", hasPrefix),
		"has_suffix":   object.NewBuiltin("strings.has_suffix", hasSuffix),
	})
}

func replaceAll(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 3 {
		return object.Errorf("strings.replace_all: expected 3 arguments, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	old, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	new, err := object.AsString(args[2])
	if err != nil {
		return nil, err
	}

	return object.NewString(strings.ReplaceAll(s, old, new)), nil
}

func trimPrefix(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("strings.trim_prefix: expected 2 arguments, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	prefix, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	return object.NewString(strings.TrimPrefix(s, prefix)), nil
}

func trimSuffix(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("strings.trim_suffix: expected 2 arguments, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	suffix, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	return object.NewString(strings.TrimSuffix(s, suffix)), nil
}

func trim(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("strings.trim: expected 2 arguments, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	cutset, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	return object.NewString(strings.Trim(s, cutset)), nil
}

func toLower(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("strings.to_lower: expected 1 argument, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	return object.NewString(strings.ToLower(s)), nil
}

func toUpper(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("strings.to_upper: expected 1 argument, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	return object.NewString(strings.ToUpper(s)), nil
}

func split(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("strings.split: expected 2 arguments, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	sep, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	parts := strings.Split(s, sep)
	items := make([]object.Object, len(parts))
	for i, part := range parts {
		items[i] = object.NewString(part)
	}

	return object.NewList(items), nil
}

func join(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("strings.join: expected 2 arguments, got %d", len(args)), nil
	}

	list, err := object.AsList(args[0])
	if err != nil {
		return nil, err
	}

	sep, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	parts := make([]string, len(list.Value()))
	for i, item := range list.Value() {
		s, err := object.AsString(item)
		if err != nil {
			return nil, err
		}
		parts[i] = s
	}

	return object.NewString(strings.Join(parts, sep)), nil
}

func contains(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("strings.contains: expected 2 arguments, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	substr, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	return object.NewBool(strings.Contains(s, substr)), nil
}

func hasPrefix(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("strings.has_prefix: expected 2 arguments, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	prefix, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	return object.NewBool(strings.HasPrefix(s, prefix)), nil
}

func hasSuffix(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("strings.has_suffix: expected 2 arguments, got %d", len(args)), nil
	}

	s, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	suffix, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	return object.NewBool(strings.HasSuffix(s, suffix)), nil
}
