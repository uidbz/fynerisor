// Package v3bind is the simplified Risor v3 experiment for fynerisor.
//
// It tests the one role v3 is meant to keep: evaluated, statically checked
// binding expressions over page state. There are no event handlers, callbacks,
// scene documents or Fyne dependencies here. The full event-driven prototype
// was removed after the decision that handlers stay on v2.
package v3bind

import (
	"context"
	"fmt"
	"sort"

	risor "github.com/deepnoodle-ai/risor/go/v3"
)

// Set is a compiled group of binding expressions over one state type.
// Programs are immutable and may run concurrently.
type Set[T any] struct {
	names    []string
	programs map[string]*risor.Program
}

// Config tunes the environment.
type Config struct {
	// Steps is the per-run step budget. Zero keeps the v3 default.
	Steps int64
}

// Compile compiles one expression per binding name. Every expression reads
// the variable `state`, whose type is derived from T with risor.TypeOf.
//
//	Compile[State](map[string]string{
//	    "count": "state.count",
//	    "title": `"Count: " + string(state.count)`,
//	})
func Compile[T any](exprs map[string]string, cfg Config) (*Set[T], error) {
	stateType, err := risor.TypeOf[T]()
	if err != nil {
		return nil, fmt.Errorf("v3bind: state type: %w", err)
	}
	decls := []risor.Decl{risor.VarType("state", stateType)}
	if cfg.Steps > 0 {
		decls = append(decls, risor.Limits{Steps: cfg.Steps})
	}
	env, err := risor.NewEnv(decls...)
	if err != nil {
		return nil, fmt.Errorf("v3bind: environment: %w", err)
	}

	names := make([]string, 0, len(exprs))
	for name := range exprs {
		names = append(names, name)
	}
	sort.Strings(names)

	set := &Set[T]{names: names, programs: make(map[string]*risor.Program, len(exprs))}
	for _, name := range names {
		expr, err := risor.Parse(exprs[name])
		if err != nil {
			return nil, fmt.Errorf("v3bind: %s: %w", name, err)
		}
		prog, err := env.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("v3bind: %s: %w", name, err)
		}
		set.programs[name] = prog
	}
	return set, nil
}

// Names returns the binding names in sorted order.
func (s *Set[T]) Names() []string { return append([]string(nil), s.names...) }

// Run evaluates every binding against state. Evaluation stops at the first
// failure; a binding error carries the binding name.
func (s *Set[T]) Run(ctx context.Context, state T) (map[string]any, error) {
	out := make(map[string]any, len(s.programs))
	for _, name := range s.names {
		v, err := s.programs[name].Run(ctx, risor.Map{"state": state})
		if err != nil {
			return nil, fmt.Errorf("v3bind: %s: %w", name, err)
		}
		out[name] = v
	}
	return out, nil
}

// RunOne evaluates one binding. It is useful for a host that refreshes only
// the properties that changed.
func (s *Set[T]) RunOne(ctx context.Context, name string, state T) (any, error) {
	prog, ok := s.programs[name]
	if !ok {
		return nil, fmt.Errorf("v3bind: unknown binding %q", name)
	}
	return prog.Run(ctx, risor.Map{"state": state})
}
